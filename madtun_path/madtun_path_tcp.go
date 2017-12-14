package madtun_path

import (
	"net"
	"time"

	"github.com/allmad/madtun/madtun_api"
	"github.com/chzyer/logex"
)

var _ madtun_api.Path = &Tcp{}

type Tcp struct {
	conn net.Conn
	ln   net.Listener
	*BasePath
}

func NewTcp() (*Tcp, error) {
	tcp := &Tcp{}
	bp := NewBasePath(tcp)
	tcp.BasePath = bp
	return tcp, nil
}

func (t *Tcp) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return logex.Trace(err)
	}
	t.conn = conn
	return nil
}

func (t *Tcp) Listen(addr string) (int, error) {
	if addr == "" {
		addr = ":0"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, logex.Trace(err)
	}

	lnAddr := ln.Addr().(*net.TCPAddr)
	t.ln = ln
	return lnAddr.Port, nil
}

func (t *Tcp) Wait(timeout time.Duration) error {
	if timeout > 0 {
		t.ln.(*net.TCPListener).SetDeadline(time.Now().Add(timeout))
	}
	conn, err := t.ln.Accept()
	if err != nil {
		return err
	}
	t.ln.Close()
	t.conn = conn
	return nil
}

func (t *Tcp) Write(buf []byte) (int, error) {
	return t.conn.Write(buf)
}

func (t *Tcp) Read(buf []byte) (int, error) {
	return t.conn.Read(buf)
}

func (t *Tcp) Close() error {
	if t.conn != nil {
		t.conn.Close()
	}
	return nil
}
