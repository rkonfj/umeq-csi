// Copyright 2022 rkonfj@fnla.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/kataras/iris/v12"
	"github.com/tasselsd/umeq-csi/pkg/attach"
	"github.com/tasselsd/umeq-csi/pkg/state"
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
