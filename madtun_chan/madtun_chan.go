package madtun_chan

import (
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/allmad/madtun/madtun_api"
	"github.com/allmad/madtun/madtun_session"
	"github.com/chzyer/logex"
)

var (
	ErrNoPathFound  = logex.Define("no available path found")
	ErrAllPathLeave = logex.Define("all path leave")
)

type Mode int

const (
	ModeServer Mode = iota
	ModeClient
)

// 调度由客户端来负责, 服务端负责满足客户端的请求
// 服务端可以通过Join把Path加入到Chan的管理
type Chan struct {
	session  *madtun_session.SessionMgr
	mode     Mode
	id       string
	provider PathProvider
	path     []madtun_api.Path
	source   rand.Source
	wl       sync.Mutex
	buf      *madtun_api.Buffer

	receiveChan chan *madtun_api.Packet
	errorChan   chan error
	stopChan    chan struct{}
	newPathChan chan struct{}
}

type PathProvider interface {
	AllocChanPath(id string) (madtun_api.Path, error)
}

type ChanPacket struct {
	Data []byte
}

func NewChan(id string, mode Mode, p PathProvider) *Chan {
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	buf := make([]byte, 4<<20)
	return &Chan{
		mode:        mode,
		session:     madtun_session.NewSessionMgr(0x10000000),
		buf:         madtun_api.NewBuffer(buf),
		id:          id,
		provider:    p,
		source:      source,
		receiveChan: make(chan *madtun_api.Packet, 1024),
		errorChan:   make(chan error, 1),
		newPathChan: make(chan struct{}, 1),
	}
}

func (c *Chan) markError(err error) {
	select {
	case c.errorChan <- err:
		close(c.errorChan)
	default:
	}
}

func (c *Chan) ServerMode(addr string) error {
	path, err := c.provider.AllocChanPath(addr)
	if err != nil {
		return err
	}
	c.Join(path)
	return nil
}

func (c *Chan) ClientMode(addr string) error {
	return nil
}

func (c *Chan) Run() error {
	go func() {
		for range time.Tick(100 * time.Millisecond) {
			select {
			case <-c.newPathChan:
				if c.mode != ModeClient {
					break
				}
				c.lock()
				pathLength := len(c.path)
				c.unlock()
				if pathLength < 4 {
					reply, err := c.Call(&madtun_api.Packet{
						Type: madtun_api.TypeNewPath,
					})
					if err != nil {
						logex.Error(err)
						continue
					}

					// received from server the new port to connect
					path, err := c.provider.AllocChanPath(c.id)
					if err != nil {
						logex.Error(err)
						return
					}
					if err := path.Dial(fmt.Sprintf("%v:%s", c.id, reply.Payload)); err != nil {
						logex.Error(err)
						return
					}
					c.Join(path)
				}
			}
		}
	}()
	return <-c.errorChan
}

func (c *Chan) lock() {
	c.wl.Lock()
}

func (c *Chan) unlock() {
	c.wl.Unlock()
}

func (c *Chan) Leave(p madtun_api.Path) {
	p.Close()
	c.lock()
	for idx, pp := range c.path {
		if pp == p {
			c.path = append(c.path[:idx], c.path[idx+1:]...)
			break
		}
	}
	c.markError(ErrAllPathLeave.Trace())
	c.unlock()
}

func (c *Chan) Join(p madtun_api.Path) {
	c.lock()
	for _, pp := range c.path {
		if pp == p {
			c.unlock()
			panic("path is already registered")
		}
	}
	c.path = append(c.path, p)
	c.unlock()
	select {
	case c.newPathChan <- struct{}{}:
	default:
	}

	go func() {
		p.ReadToChan(c.receiveChan)
		c.Leave(p)
	}()
}

func (c *Chan) handleInternalPacket(p *madtun_api.Packet) {
	if c.mode == ModeClient {
		switch p.Type {
		case madtun_api.TypeNewPathResp:
			c.session.Push(p.ReqId, p)
		}
	} else if c.mode == ModeServer {
		switch p.Type {
		case madtun_api.TypeNewPath:
			// recevied from client want to open a new path
			path, err := c.provider.AllocChanPath(c.id)
			if err != nil {
				logex.Error(err)
				return
			}
			port, err := path.Listen("")
			if err != nil {
				logex.Error(err)
				return
			}
			c.WritePacket(&madtun_api.Packet{
				ReqId:   p.ReqId,
				Type:    madtun_api.TypeNewPathResp,
				Payload: []byte(fmt.Sprint(port)),
			})
			path.Wait(time.Second)
			c.Join(path)
		}
	}
}

func (c *Chan) ReadPacket() (*madtun_api.Packet, error) {
reread:
	select {
	case packet := <-c.receiveChan:
		if packet == nil {
			return nil, io.EOF
		}
		switch packet.Type {
		case madtun_api.TypeData, madtun_api.TypeDataResp:
			return packet, nil
		default:
			go c.handleInternalPacket(packet)
			goto reread
		}
	case <-c.stopChan:
	}
	return nil, io.EOF
}

func (c *Chan) findPathLocked() madtun_api.Path {
	length := len(c.path)
	randPos := int(c.source.Int63() % int64(length))
	for i := 0; i < length; i++ {
		offset := randPos + i
		if offset >= length {
			offset -= length
		}
		p := c.path[offset]
		if p.IsAvailable() {
			return p
		}
	}
	return nil
}

func (c *Chan) Call(packet *madtun_api.Packet) (*madtun_api.Packet, error) {
	c.session.Register(&packet.ReqId)
	if err := c.WritePacket(packet); err != nil {
		return nil, logex.Trace(err)
	}
	reply, _ := c.session.Pop(packet.ReqId, time.Second)
	if reply == nil {
		return nil, fmt.Errorf("timeout")
	}
	return reply, nil
}

func (c *Chan) WritePacket(packet *madtun_api.Packet) error {
	c.lock()
	defer c.unlock()

	p := c.findPathLocked()
	if p == nil {
		return ErrNoPathFound.Trace()
	}
	return p.WritePacket(packet)
}

func (c *Chan) Close() error {
	close(c.receiveChan)
	for _, p := range c.path {
		p.Close()
	}
	return nil
}
