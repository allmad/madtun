package madtun_server

import (
	"github.com/allmad/madtun/madtun_chan"
	"github.com/allmad/madtun/madtun_meta"
)

type MetaDelegate struct {
	s *Server
}

func NewMetaDelegate(s *Server) *MetaDelegate {
	return &MetaDelegate{
		s: s,
	}
}

func (s *MetaDelegate) OnNewClient(req *madtun_meta.MsgClientInfo) (*madtun_meta.MsgClientInfoResp, error) {
	ch := madtun_chan.NewChan("", madtun_chan.ModeServer, s.s.pathProvider)
	return &madtun_meta.MsgClientInfoResp{
		Connect: uri,
	}, nil
}
