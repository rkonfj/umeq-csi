package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/kataras/iris/v12"
	"github.com/tasselsd/umeq-csi/internel/attach"
	"github.com/tasselsd/umeq-csi/internel/state"
)

var gracefulShutdowns []func() error

func main() {
	app := iris.New()

	attacherKv := state.NewFsKvStore(filepath.Join(config.StatePath, "attacher"))
	var attacher attach.Attacher
	if len(config.Socks) == 0 {
		log.Println("Using virsh attacher")
		attacher = attach.NewVirshAttacher(attacherKv)
	} else {
		log.Println("Using qmp attacher")
		attacher = attach.NewQmpAttacher(attacherKv, config.Socks)
	}

	agentKv := state.NewFsKvStore(filepath.Join(config.StatePath, "agent"))
	agent := NewAgent(config.Storage, agentKv, attacher)

	Routing(app, agent)

	defer func() {
		for _, hook := range gracefulShutdowns {
			err := hook()
			if err != nil {
				log.Println(err)
			}
		}
	}()

	app.Listen(fmt.Sprintf("0.0.0.0:%d", config.ServerPort))
}
