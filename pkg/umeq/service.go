package umeq

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/mount"
)

type Csi struct {
	NodeID        string
	DriverName    string
	VendorVersion string
}

func (c *Csi) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	log.Printf("NodePublishVolume:%v", req)
	// Check arguments
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	targetPath := req.GetTargetPath()

	path, err := getDevPath(req.VolumeId)
	if err != nil {
		log.Printf("ERR:%s", err)
	}

	mounter := mount.New("")

	notMnt, err := mount.IsNotMountPoint(mount.New(""), targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return nil, fmt.Errorf("create target path: %w", err)
			}
			notMnt = true
		} else {
			return nil, fmt.Errorf("check target path: %w", err)
		}
	}

	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	fsType := req.GetVolumeCapability().GetMount().GetFsType()

	deviceId := ""
	if req.GetPublishContext() != nil {
		deviceId = req.GetPublishContext()["deviceID"]
	}

	readOnly := req.GetReadonly()
	volumeId := req.GetVolumeId()
	attrib := req.GetVolumeContext()
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()

	log.Printf("target %v\nfstype %v\ndevice %v\nreadonly %v\nvolumeId %v\nattributes %v\nmountflags %v\n",
		targetPath, fsType, deviceId, readOnly, volumeId, attrib, mountFlags)

	options := []string{}
	if readOnly {
		options = append(options, "ro")
	}

	if out, err := exec.Command("blkid", path).Output(); err != nil {
		if out, err = exec.Command("mkfs.ext4", path).Output(); err != nil {
			log.Println("mkfs.ext4 ERR:", err)
		} else {
			log.Println(string(out))
		}
	} else {
		log.Println(string(out))
	}

	if err := mounter.Mount(path, targetPath, "", options); err != nil {
		var errList strings.Builder
		errList.WriteString(err.Error())
		return nil, fmt.Errorf("failed to mount device: %s at %s: %s", path, targetPath, errList.String())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (c *Csi) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	log.Printf("NodeUnpublishVolume:%v", req)
	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	targetPath := req.GetTargetPath()

	// Unmount only if the target path is really a mount point.
	if notMnt, err := mount.IsNotMountPoint(mount.New(""), targetPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("check target path: %w", err)
		}
	} else if !notMnt {
		// Unmounting the image or filesystem.
		err = mount.New("").Unmount(targetPath)
		if err != nil {
			return nil, fmt.Errorf("unmount target path: %w", err)
		}
	}
	// Delete the mount point.
	// Does not return error for non-existent path, repeated calls OK for idempotency.
	if err := os.RemoveAll(targetPath); err != nil {
		return nil, fmt.Errorf("remove target path: %w", err)
	}
	glog.V(4).Infof("volume %s has been unpublished.", targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (c *Csi) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	log.Printf("NodeGetInfo:%v", req)
	resp := &csi.NodeGetInfoResponse{
		NodeId:            c.NodeID,
		MaxVolumesPerNode: 100,
	}

	resp.AccessibleTopology = &csi.Topology{
		Segments: map[string]string{"TopologyKeyNode": c.NodeID},
	}

	resp.MaxVolumesPerNode = 100
	log.Printf("NodeGetInfo Resp:%v", resp)
	return resp, nil
}

func (c *Csi) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	log.Printf("NodeGetCapabilities:%v", req)
	caps := []*csi.NodeServiceCapability{
		{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_VOLUME_CONDITION,
				},
			},
		},
		{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
				},
			},
		},
	}

	return &csi.NodeGetCapabilitiesResponse{Capabilities: caps}, nil
}

func (c *Csi) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	log.Printf("NodeStageVolume:%v", req)
	return &csi.NodeStageVolumeResponse{}, nil
}

func (c *Csi) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	log.Printf("NodeUnstageVolume:%v", req)
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (c *Csi) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	log.Printf("NodeGetVolumeStats:%v", in)
	return &csi.NodeGetVolumeStatsResponse{}, nil
}

func (c *Csi) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	log.Printf("NodeExpandVolume:%v", req)
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	glog.V(4).Infof("NodeExpandVolume Req: %v", req)
	return &csi.NodeExpandVolumeResponse{}, nil
}

func (c *Csi) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	log.Printf("GetPluginInfo:%v", req)
	glog.V(5).Infof("Using default GetPluginInfo")

	if c.DriverName == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if c.VendorVersion == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	return &csi.GetPluginInfoResponse{
		Name:          c.DriverName,
		VendorVersion: c.VendorVersion,
	}, nil
}

func (c *Csi) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	log.Printf("Probe:%v", req)
	return &csi.ProbeResponse{}, nil
}

func (c *Csi) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	log.Printf("GetPluginCapabilities:%v", req)
	glog.V(5).Infof("Using default capabilities")
	caps := []*csi.PluginCapability{
		{
			Type: &csi.PluginCapability_Service_{
				Service: &csi.PluginCapability_Service{
					Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
				},
			},
		},
		{
			Type: &csi.PluginCapability_Service_{
				Service: &csi.PluginCapability_Service{
					Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
				},
			},
		},
	}
	return &csi.GetPluginCapabilitiesResponse{Capabilities: caps}, nil
}

func (c *Csi) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (resp *csi.CreateVolumeResponse, finalErr error) {
	log.Printf("CreateVolume:%v", req)

	err := createVolume(req.Name, req.CapacityRange.RequiredBytes)
	if err != nil {
		log.Printf("ERR:%s", err)
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      req.Name,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			VolumeContext: req.GetParameters(),
			ContentSource: req.GetVolumeContentSource(),
			AccessibleTopology: []*csi.Topology{
				{Segments: map[string]string{"TopologyKeyNode": c.NodeID}},
			},
		},
	}, nil
}

func (c *Csi) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	log.Printf("DeleteVolume:%v", req)
	err := deleteVolume(req.VolumeId)
	if err != nil {
		log.Printf("ERR:%s", err)
	}
	return &csi.DeleteVolumeResponse{}, nil
}
func (c *Csi) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	log.Printf("ControllerGetCapabilities:%v", req)
	res := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
					},
				},
			},
		},
	}
	log.Printf("ControllerGetCapabilities Res:%v", res)
	return res, nil
}

func (c *Csi) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	log.Printf("ValidateVolumeCapabilities:%v", req)
	return &csi.ValidateVolumeCapabilitiesResponse{}, nil
}

func (c *Csi) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	log.Printf("ControllerPublishVolume:%v", req)
	err := publishVolume(req.VolumeId, req.NodeId)
	if err != nil {
		log.Printf("ERR:%s", err)
	}
	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{},
	}, nil
}

func (c *Csi) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	log.Printf("ControllerUnpublishVolume:%v", req)
	err := unpublishVolume(req.VolumeId, req.NodeId)
	if err != nil {
		log.Printf("ERR:%s", err)
	}
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (c *Csi) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	log.Printf("GetCapacity:%v", req)
	return &csi.GetCapacityResponse{}, nil
}

func (c *Csi) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	log.Printf("ListVolumes:%v", req)
	return &csi.ListVolumesResponse{}, nil
}

func (c *Csi) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	log.Printf("ControllerGetVolume:%v", req)
	return &csi.ControllerGetVolumeResponse{}, nil
}

// CreateSnapshot uses tar command to create snapshot for hostpath volume. The tar command can quickly create
// archives of entire directories. The host image must have "tar" binaries in /bin, /usr/sbin, or /usr/bin.
func (c *Csi) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	log.Printf("CreateSnapshot:%v", req)
	return &csi.CreateSnapshotResponse{}, nil
}

func (c *Csi) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	log.Printf("DeleteSnapshot:%v", req)
	return &csi.DeleteSnapshotResponse{}, nil
}

func (c *Csi) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	log.Printf("ListSnapshots:%v", req)
	return &csi.ListSnapshotsResponse{}, nil
}

func (c *Csi) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	log.Printf("ControllerExpandVolume:%v", req)
	return &csi.ControllerExpandVolumeResponse{}, nil
}
