all: agent plugin
plugin:
	docker build . -t tasselsd/umeq-csi:0.0.1;\
	docker push tasselsd/umeq-csi:0.0.1
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