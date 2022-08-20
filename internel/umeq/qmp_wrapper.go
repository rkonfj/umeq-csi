package umeq

import (
	"fmt"
	"log"
	"time"

	"github.com/tasselsd/umeq-csi/internel/qmp"
)

var mons map[string]*qmp.Monitor

func init() {
	mons = make(map[string]*qmp.Monitor)
	k1, err := qmp.NewMonitor("/run/k1.mon.sock", 60*time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	k2, err := qmp.NewMonitor("/run/k2.mon.sock", 60*time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	k3, err := qmp.NewMonitor("/run/k3.mon.sock", 60*time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	mons["k1"] = k1
	mons["k2"] = k2
	mons["k3"] = k3
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
