package main

import (
	"os"

	"github.com/allmad/madtun/madtun_client"
	"github.com/allmad/madtun/madtun_server"
	"github.com/chzyer/flagly"
	"github.com/chzyer/logex"
)

type Madtun struct {
	Server *madtun_server.Config `flagly:"handler"`
	Client *madtun_client.Config `flagly:"handler"`
}

func main() {
	fset := flagly.New(os.Args[0])
	fset.Compile(&Madtun{})
	err := fset.Run(os.Args[1:])
	if err != nil {
		logex.Fatal(err)
	}
}
