oci_registry?=tasselsd
version?=latest

git_hash:=$(shell git rev-parse HEAD)

all: agent plugin
plugin:
	docker build --build-arg VERSION=${version}-${git_hash} . -t ${oci_registry}/umeq-csi:${version};\
	docker push ${oci_registry}/umeq-csi:${version}
agent:
	cd cmd/agent;\
	go build -ldflags "-s -w";\
	scp agent root@192.168.3.11:/opt/umeq-csi/
agentctl:
	cd cmd/agentctl;\
	go build -ldflags "-s -w";\
	scp agentctl root@192.168.3.11:/usr/bin/
lint:
	cd cmd/plugin;go build -ldflags "-s -w";\
	cd ../agentctl;go build -ldflags "-s -w";\
	cd ../agent;go build -ldflags "-s -w"
dist:
	docker run --rm -e GOPROXY=https://goproxy.cn -v ${shell echo $GOPATH}:/go \
	-v ${shell pwd}:/app -w /app golang:1.19.2-buster sh -c \
	"env;cd cmd/agent;go build -ldflags \"-s -w\" -o agent-linux_amd64;cd ../agentctl;\
	go build -ldflags \"-s -w\" -o agentctl-linux_amd64"