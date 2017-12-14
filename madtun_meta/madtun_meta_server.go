package madtun_meta

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chzyer/logex"
)

var (
	ErrNilDelegater = logex.Define("delegater is nil")
)

type ServerConfig struct {
	Addr     string
	Delegate ServerDelegate
}

type Server struct {
	cfg *ServerConfig
	mux *http.ServeMux
	svr *http.Server
}

type ServerDelegate interface {
	OnNewClient(*MsgClientInfo) (*MsgClientInfoResp, error)
}

func NewServer(cfg *ServerConfig) *Server {
	svr := &Server{
		cfg: cfg,
	}
	svr.init()
	return svr
}

func (s *Server) init() {
	s.mux = http.NewServeMux()
	s.svr = &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           s.mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	s.mux.HandleFunc("/newClient", s.onNewClientHandler)
}

func (s *Server) unmarshal(req *http.Request, val interface{}) error {
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, val)
}

type Error struct {
	Msg string
}

func (s *Server) responseAll(w http.ResponseWriter, val interface{}, err error) {
	if err != nil {
		s.response(w, err)
	} else {
		s.response(w, val)
	}
}

func (s *Server) response(w http.ResponseWriter, val interface{}) {
	isError := false
	if err, ok := val.(error); ok {
		isError = true
		val = Error{Msg: err.Error()}
	}
	body, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	if isError {
		w.WriteHeader(400)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// ---------------------------------------------------------------------------

func (s *Server) onNewClientHandler(w http.ResponseWriter, req *http.Request) {
	var n MsgClientInfo
	if err := s.unmarshal(req, &n); err != nil {
		s.response(w, err)
		return
	}
	resp, err := s.cfg.Delegate.OnNewClient(&n)
	s.responseAll(w, resp, err)
}

// ---------------------------------------------------------------------------

func (s *Server) Run() error {
	if s.cfg.Delegate == nil {
		return ErrNilDelegater.Trace()
	}
	logex.Infof("[meta] listen on: %q", s.cfg.Addr)
	return s.svr.ListenAndServe()
}
