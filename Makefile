all: agent plugin
plugin:
	cd cmd/plugin;go build -ldflags "-s -w";\
	docker build . -t kcr.xfs.pub/umeq-csi-plugin:1.0.3;\
	docker push kcr.xfs.pub/umeq-csi-plugin:1.0.3
agent:
	cd cmd/agent;\
	go build -ldflags "-s -w";\
	scp agent root@192.168.3.11:/opt/umeq-csi/
