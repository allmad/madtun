package madtun_path

import (
	"bufio"
	"sync"
	"time"

	"github.com/allmad/madtun/madtun_api"
	"github.com/chzyer/logex"
)

type Pather interface {
	Listen(addr string) (port int, err error)
	Dial(string) error
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Close() error
}

type BasePath struct {
	pather   Pather
	buf      *madtun_api.Buffer
	sendTick chan struct{}
	m        sync.Mutex
	closed   bool
}

func NewBasePath(p Pather) *BasePath {
	bp := &BasePath{
		buf:      madtun_api.NewBuffer(make([]byte, 4<<20)),
		pather:   p,
		sendTick: make(chan struct{}, 1),
	}
	go func() {
		for {
			<-bp.sendTick
			time.Sleep(time.Millisecond)
			if err := bp.flush(); err != nil {
				logex.Error(err)
				panic("check error")
				return
			}
		}
	}()
	return bp
}

func (b *BasePath) close() {
	b.closed = true
}

func (b *BasePath) IsAvailable() bool {
	return !b.closed
}

func (b *BasePath) ReadToChan(ch chan<- *madtun_api.Packet) {
	defer b.pather.Close()
	r := bufio.NewReader(b.pather)
	data := make([]byte, 4<<20)
	for b.IsAvailable() {
		var packet madtun_api.Packet
		if err := packet.Decode(r, data); err != nil {
			logex.Error(err)
			break
		}
		ch <- &packet
	}
}

func (b *BasePath) flush() error {
	b.m.Lock()
	defer b.m.Unlock()
	if len(b.buf.Bytes()) == 0 {
		return nil
	}
	_, err := b.pather.Write(b.buf.Bytes())
	if err != nil {
		return err
	}
	b.buf.Reset()
	return nil
}

func (b *BasePath) WritePacket(packet *madtun_api.Packet) error {
	b.m.Lock()
	defer b.m.Unlock()

	if b.buf.Avail() < packet.Size() {
		if err := b.flush(); err != nil {
			return logex.Trace(err)
		}
	}

	packet.Marshal(b.buf)
	if b.buf.Len()*2 > b.buf.Cap() {
		return logex.Trace(b.flush())
	}
	select {
	case b.sendTick <- struct{}{}:
	default:
	}
	return nil
}
