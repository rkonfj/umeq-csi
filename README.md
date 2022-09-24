# CSI driver for qemu vm 

> DON'T USED FOR PRODUCTION!

### Summary
use `qemu-img` create `qcow2` format disk

use `qmp protocol`(for qemu) or use `virsh command`(for libvirt) attach disk to node 

CSI: controller create volume  
CSI: controller publish volume  
CSI: node publish volume  
CSI: node unpublish volume  
CSI: controller unpublish volume  
CSI: controller deleteVolume  

`agent` the controller backend  
`plugin` the csi implemention (should use [csi sidecars](https://kubernetes-csi.github.io/docs/sidecar-containers.html) to provide full stack csi driver)

### Get Started
> Requirements:
> a kubernetes cluster that running on qemu based vm like this:
> ```
> > qemu-system-x86_64 -name k1 -enable-kvm -nographic -cpu host -drive format=qcow2,file=/fs/trust/vm/k1.qcow2,if=virtio -qmp unix:/run/k1.mon.sock,server,nowait -net bridge,br=br0 -m 8g -smp 8 -net nic,macaddr=52:54:00:00:00:01,model=virtio -vnc :1
> 
> > qemu-system-x86_64 -name k2 -enable-kvm -nographic -cpu host -drive format=qcow2,file=/fs/trust/vm/k2.qcow2,if=virtio -qmp unix:/run/k2.mon.sock,server,nowait -net bridge,br=br0 -m 8g -smp 8 -net nic,macaddr=52:54:00:00:00:02,model=virtio -vnc :2
>
> > qemu-system-x86_64 -name k3 -enable-kvm -nographic -cpu host -drive format=qcow2,file=/fs/trust/vm/k3.qcow2,if=virtio -qmp unix:/run/k3.mon.sock,server,nowait -net bridge,br=br0 -m 8g -smp 8 -net nic,macaddr=52:54:00:00:00:03,model=virtio -vnc :3
> ```

1. Deploy csi agent to qemu host machine
```
> cat <EOF > config.yml
serverPort: 8080
socks:
- name: k1
  path: /run/k1.mon.sock
- name: k2
  path: /run/k2.mon.sock
- name: k3
  path: /run/k3.mon.sock
storage:
  default: /fs/trust/vm/csi/
EOF

> systemctl enable --now umeq-csi-agent
``` 
2. Deploy csi to kubernetes
> Tips: replace environment variable `AGENT_SERVER` as your correct host csi agent rest api address
```
kubectl create -f https://raw.githubusercontent.com/tasselsd/umeq-csi/master/distro/deployment.yml
```

3. Create StorageClass use kind `default`
```
kubectl create -f https://raw.githubusercontent.com/tasselsd/umeq-csi/master/distro/storageclass.yml
```
4. Try to create PersistentVolumeClaim
```
kubectl create -f https://raw.githubusercontent.com/tasselsd/umeq-csi/master/distro/pvc-test.yml
```
