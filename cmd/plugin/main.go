package main

import (
	"os"

	"github.com/tasselsd/umeq-csi/internel/umeq"
)

func main() {
	endpoint := os.Getenv("CSI_ENDPOINT")
	nodeId := os.Getenv("NODE_NAME")
	agentServer := os.Getenv("AGENT_SERVER")
	if endpoint == "" {
		panic("system environment CSI_ENDPOINT is required!")
	}
	if nodeId == "" {
		panic("system environment NODE_NAME is required!")
	}
	if agentServer == "" {
		panic("system environment AGENT_SERVER is required!")
	}
	agent := umeq.NewAgentService(agentServer)
	csi := umeq.NewCsi(nodeId, "umeq-csi.xiaomakai.com", "1.0.0", agent)

	s := umeq.NewNonBlockingGRPCServer()
	s.Start("unix://"+endpoint, csi, csi, csi)
	s.Wait()
}
