package main

import (
	"os"
	"path/filepath"

	"github.com/tasselsd/umeq-csi/internel/attach"
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
