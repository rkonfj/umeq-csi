all: agent plugin
plugin:
	cd cmd/plugin;go build -ldflags "-s -w";\
	docker build . -t kcr.xfs.pub/umeq-csi-plugin:1.0.14;\
	docker push kcr.xfs.pub/umeq-csi-plugin:1.0.14
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
	cd ../agent;go build -ldflags "-s -w"