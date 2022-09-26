oci_registry?=tasselsd

all: agent plugin
plugin:
	docker build . -t ${oci_registry}/umeq-csi:0.0.2;\
	docker push ${oci_registry}/umeq-csi:0.0.2
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