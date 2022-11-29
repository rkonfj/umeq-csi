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
	"os"
	"path/filepath"

	"github.com/tasselsd/umeq-csi/pkg/attach"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Socks      []attach.Sock
	Storage    map[string]string
	StatePath  string
	ServerPort int `yaml:"serverPort"`
}

var config *Config = &Config{}

func init() {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
	// Default Config
	if _, ok := config.Storage["default"]; !ok {
		panic("config.storage.default is required!")
	}
	if config.StatePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		config.StatePath = filepath.Join(cwd, ".state")
	}
	if config.ServerPort == 0 {
		config.ServerPort = 8080
	}

}
