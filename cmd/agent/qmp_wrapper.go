package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tasselsd/umeq-csi/internel/qmp"
)

var mons map[string]*qmp.Monitor = make(map[string]*qmp.Monitor)

func initMons(qs []Qmp) {
	for _, q := range qs {
		mon, err := qmp.NewMonitor(q.Sock, 60*time.Second)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Registered %s -> %s\n", q.Name, q.Sock)
		mons[q.Name] = mon
	}
}

func Exec(node, cmd string) error {
	var out string

	mon := mons[node]

	if mon == nil {
		return fmt.Errorf("mon %s not found", node)
	}

	if err := mon.Run(qmp.Command{
		Name: "human-monitor-command",
		Arguments: &qmp.HumanCommand{
			Cmd: cmd,
		},
	}, &out); err != nil {
		return err
	}
	log.Printf("node: %s, cmd: %s, out: %s\n", node, cmd, out)
	return nil
}
