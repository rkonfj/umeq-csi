package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kataras/iris/v12"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/pkg/transport"
)

var gracefulShutdowns []func() error

func main() {
	app := iris.New()

	tlsInfo := transport.TLSInfo{
		CertFile:      config.Etcd.Cert,
		KeyFile:       config.Etcd.Key,
		TrustedCAFile: config.Etcd.Ca,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Etcd.Endpoints,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	agent := NewAgent(cli, config.ImagePath)

	Routing(app, agent)

	defer func() {
		for _, hook := range gracefulShutdowns {
			err = hook()
			if err != nil {
				log.Println(err)
			}
		}
	}()

	app.Listen(fmt.Sprintf("0.0.0.0:%d", config.ServerPort))
}
