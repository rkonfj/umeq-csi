# CSI driver for qemu vm 

use `qemu-img` create `qcow2` format disk

use `qmp` attach disk to node

CSI: controller create volume  
CSI: controller publish volume  
CSI: node publish volume  
CSI: node unpublish volume  
CSI: controller unpublish volume  
CSI: controller deleteVolume  

`host-agent` controller backend  
`plugin` csi implement (also should use [csi sidecars](https://kubernetes-csi.github.io/docs/sidecar-containers.html) to provide full stack csi driver)
