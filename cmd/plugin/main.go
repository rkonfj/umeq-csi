package main

import (
	"flag"
	"log"
	"os"

	"github.com/openxiaoma/umeq-csi/pkg/umeq"
)

func main() {
	flag.Parse()
	s := umeq.NewNonBlockingGRPCServer()
	csi := umeq.Csi{
		NodeID:        os.Getenv("NODE_NAME"),
		DriverName:    "umeq-csi.xiaomakai.com",
		VendorVersion: "1.0.0",
	}
	s.Start("unix://"+os.Getenv("CSI_ENDPOINT"), &csi, &csi, &csi)
	log.Println("listen on unix://" + os.Getenv("CSI_ENDPOINT"))
	s.Wait()
}
