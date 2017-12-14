package madtun_api

import "time"

type Path interface {
	ReadToChan(chan<- *Packet)
	WritePacket(*Packet) error
	Listen(addr string) (port int, err error)
	Wait(timeout time.Duration) error
	Dial(string) error
	Close() error
	IsAvailable() bool
}
