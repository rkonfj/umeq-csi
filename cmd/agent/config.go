package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Qmp struct {
	Name string
	Sock string
}

type Etcd struct {
	Endpoints []string
	Cert      string
	Key       string
	Ca        string
}

type Config struct {
	Qmp        []Qmp
	ImagePath  string `yaml:"imagePath"`
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
	if config.ImagePath == "" {
		panic("config.imagePath is required!")
	}
	if len(config.Etcd.Endpoints) == 0 {
		panic("config.etcd.endpoints is required!")
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
