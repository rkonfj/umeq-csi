package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/tasselsd/umeq-csi/internel/qmp"
)

type Attacher interface {
	Exec(node, cmd string) error
}

type QmpAttacher struct {
	mons map[string]*qmp.Monitor
}

func NewQmpAttacher(qs []Qmp) *QmpAttacher {
	attacher := QmpAttacher{
		mons: make(map[string]*qmp.Monitor),
	}
	for _, q := range qs {
		mon, err := qmp.NewMonitor(q.Sock, 60*time.Second)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Registered %s -> %s\n", q.Name, q.Sock)
		attacher.mons[q.Name] = mon
	}
	return &attacher
}

func (q *QmpAttacher) Exec(node, cmd string) error {
	var out string

	mon := q.mons[node]

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
	log.Printf("[qmp]node: %s, cmd: %s, out: %s\n", node, cmd, out)
	return nil
}

type VirshAttacher struct {
}

func NewVirshAttacher() *VirshAttacher {
	return &VirshAttacher{}
}

func (v *VirshAttacher) Exec(node, cmd string) error {
	out, err := exec.Command("virsh", "qemu-monitor-command", "--hmp", node, fmt.Sprintf("'%s'", cmd)).Output()
	if err != nil {
		return err
	}
	log.Printf("[virsh]node: %s, cmd: %s, out: %s\n", node, cmd, string(out))
	return nil
}
