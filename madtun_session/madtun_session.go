package madtun_session

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/allmad/madtun/madtun_api"
)

type SessionMgr struct {
	reqIdSeq uint32
	prefix   uint32
	session  map[uint32]chan *madtun_api.Packet
	m        sync.Mutex
}

func NewSessionMgr(prefix uint32) *SessionMgr {
	return &SessionMgr{
		session: make(map[uint32]chan *madtun_api.Packet),
	}
}

func (s *SessionMgr) alloc() uint32 {
	return atomic.AddUint32(&s.reqIdSeq, 1) & s.prefix
}

func (s *SessionMgr) Register(n *uint32) chan *madtun_api.Packet {
	if *n != 0 {
		return nil
	}
	*n = s.alloc()
	reply := make(chan *madtun_api.Packet, 1)
	s.m.Lock()
	s.session[*n] = reply
	s.m.Unlock()
	return reply
}

func (s *SessionMgr) Push(n uint32, p *madtun_api.Packet) bool {
	s.m.Lock()
	reply := s.session[n]
	s.m.Unlock()
	if reply == nil {
		return false
	}
	reply <- p
	return true
}

func (s *SessionMgr) Pop(n uint32, timeout time.Duration) (*madtun_api.Packet, bool) {
	s.m.Lock()
	reply := s.session[n]
	s.m.Unlock()
	if reply == nil {
		return nil, false
	}
	select {
	case <-time.After(timeout):
		return nil, true
	case pkt := <-reply:
		return pkt, true
	}
}
