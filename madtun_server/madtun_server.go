package madtun_server

import (
	"github.com/allmad/madtun/madtun_common"
	"github.com/allmad/madtun/madtun_meta"
	"github.com/chzyer/logex"
)

func Run(cfg *Config) error {
	svr := &Server{cfg: cfg}
	svr.init()
	return svr.Run()
}

type Server struct {
	cfg *Config

	meta *madtun_meta.Server
}

func (s *Server) init() {
	metaDelegate := NewMetaDelegate(s)
	s.meta = madtun_meta.NewServer(&madtun_meta.ServerConfig{
		Addr:     s.cfg.Meta,
		Delegate: metaDelegate,
	})
}

func (s *Server) Run() error {
	wg := madtun_common.NewWaitGroup()
	wg.Run(s.meta.Run)
	if err := wg.Wait(); err != nil {
		return logex.Trace(err)
	}
	return nil
}
