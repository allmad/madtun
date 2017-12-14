package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/allmad/madtun/madtun_api"
	"github.com/allmad/madtun/madtun_chan"
	"github.com/allmad/madtun/madtun_path"
)

type PathProvider struct{}

func (*PathProvider) AllocChanPath(id string) (madtun_api.Path, error) {
	tcp, err := madtun_path.NewTcp()
	if err != nil {
		return nil, err
	}
	return tcp, nil
}

func main() {
	rootConn, err := madtun_path.NewTcp()
	if err != nil {
		panic(err)
	}
	defer rootConn.Close()
	pp := &PathProvider{}

	mode := madtun_chan.ModeServer
	dialAddr := ""
	host := ""
	if len(os.Args) > 1 {
		dialAddr = os.Args[1]
		if dialAddr != "" && dialAddr[0] != ':' {
			mode = madtun_chan.ModeClient
			host, _, _ = net.SplitHostPort(dialAddr)
		}
	}

	ch := madtun_chan.NewChan(host, mode, pp)
	defer ch.Close()

	if dialAddr == "" || dialAddr[0] == ':' {
		port, err := rootConn.Listen(dialAddr)
		if err != nil {
			panic(err)
		}

		fmt.Println(port)
		if err := rootConn.Wait(10 * time.Second); err != nil {
			panic(err)
		}
	} else {
		if err := rootConn.Dial(dialAddr); err != nil {
			panic(err)
		}
	}
	ch.Join(rootConn)

	errChan := make(chan error)

	go func() {
		for {
			packet, err := ch.ReadPacket()
			if err != nil {
				errChan <- err
			}
			println(string(packet.Payload))
		}
	}()

	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			line, _, err := r.ReadLine()
			if err != nil {
				errChan <- err
			}

			if err := ch.WritePacket(&madtun_api.Packet{
				Type:    madtun_api.TypeData,
				Payload: []byte(string(line)),
			}); err != nil {
				errChan <- err
			}
		}
	}()

	ch.Run()
	<-errChan
}
