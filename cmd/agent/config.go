package main

import (
	"os"
	"path/filepath"

	"github.com/tasselsd/umeq-csi/internel/attach"
	"gopkg.in/yaml.v3"
)

type Etcd struct {
	Endpoints []string
	Cert      string
	Key       string
	Ca        string
}

type Config struct {
	Socks      []attach.Sock
	Storage    map[string]string
	StatePath  string
	Etcd       Etcd
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
	if len(config.Etcd.Endpoints) == 0 {
		panic("config.etcd.endpoints is required!")
	}
	if config.StatePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		config.StatePath = filepath.Join(cwd, ".state")
	}
	if config.Etcd.Cert == "" {
		config.Etcd.Cert = "etcd.crt"
	}
	if config.Etcd.Key == "" {
		config.Etcd.Key = "etcd.key"
	}
	if config.Etcd.Ca == "" {
		config.Etcd.Ca = "etcd-ca.crt"
	}
	if config.ServerPort == 0 {
		config.ServerPort = 8080
	}

}
