// Copyright 2022 rkonfj@fnla.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package umeq

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/mount"
)

type Csi struct {
	NodeID        string
	DriverName    string
	VendorVersion string
	Agent         *AgentService
	lock          sync.Locker
}

func NewCsi(nodeId, driverName, venderVersion string, agent *AgentService) *Csi {
	return &Csi{
		NodeID:        nodeId,
		DriverName:    driverName,
		VendorVersion: venderVersion,
		Agent:         agent,
		lock:          &sync.Mutex{},
	}
}

func (c *Csi) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Printf("NodePublishVolume %v", req)
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
probe:
	path, err := c.Agent.GetDevPath(req.VolumeId)
	if err != nil {
		log.Println("[warn] get devpath error, try publish volume", req.VolumeId, c.NodeID)
		err = c.Agent.PublishVolume(req.VolumeId, c.NodeID)
		if err == nil {
			time.Sleep(time.Millisecond * 300)
			goto probe
		}
		log.Println("[warn] publishVolume error", err)
		return nil, fmt.Errorf("get dev-path err: %w", err)
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Println("[warn] get block device error, try publish volume", req.VolumeId, c.NodeID)
		err = c.Agent.PublishVolume(req.VolumeId, c.NodeID)
		if err == nil {
			time.Sleep(time.Millisecond * 300)
			goto probe
		}
		log.Println("[warn] publishVolume error", err)
		return nil, fmt.Errorf("%s not ready yet", path)
	}

	notMnt, err := mount.IsNotMountPoint(mount.New(""), targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return nil, fmt.Errorf("create target path: %w", err)
			}
			log.Println("[info] created target path:", targetPath)
			notMnt = true
		} else {
			return nil, fmt.Errorf("check target path: %w", err)
		}
	}

	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	readOnly := req.GetReadonly()

	options := []string{}
	if readOnly {
		options = append(options, "ro")
	}

	if out, err := exec.Command("fsck.ext4", "-n", path).Output(); err != nil {
		if out, err = exec.Command("mkfs.ext4", path).Output(); err != nil {
			log.Println("[error] mkfs.ext4 failed", err)
			return nil, err
		} else {
			log.Println(string(out))
		}
	} else {
		log.Println(string(out))
	}

	if err := mount.New("").Mount(path, targetPath, "ext4", options); err != nil {
		var errList strings.Builder
		errList.WriteString(err.Error())
		return nil, fmt.Errorf("failed to mount device: %s at %s: %s", path, targetPath, errList.String())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (c *Csi) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	log.Printf("NodeUnpublishVolume %v", req)
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
	log.Printf("NodeGetInfo %v", req)
	resp := &csi.NodeGetInfoResponse{
		NodeId:            c.NodeID,
		MaxVolumesPerNode: 100,
	}

	resp.AccessibleTopology = &csi.Topology{
		Segments: map[string]string{"TopologyKeyNode": c.NodeID},
	}

	resp.MaxVolumesPerNode = 100
	return resp, nil
}

func (c *Csi) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	log.Printf("NodeGetCapabilities %v", req)
	caps := []*csi.NodeServiceCapability{
		{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_VOLUME_CONDITION,
				},
			},
		},
	}

	return &csi.NodeGetCapabilitiesResponse{Capabilities: caps}, nil
}

func (c *Csi) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	log.Printf("NodeStageVolume %v", req)
	return &csi.NodeStageVolumeResponse{}, nil
}

func (c *Csi) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	log.Printf("NodeUnstageVolume %v", req)
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (c *Csi) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	log.Printf("NodeGetVolumeStats %v", in)
	return &csi.NodeGetVolumeStatsResponse{}, nil
}

func (c *Csi) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	log.Printf("NodeExpandVolume %v", req)
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	if out, err := exec.Command("resize2fs", req.VolumePath).Output(); err != nil {
		log.Println("resize2fs ERR:", err)
	} else {
		log.Println(out)
	}
	return &csi.NodeExpandVolumeResponse{}, nil
}

func (c *Csi) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	log.Printf("GetPluginInfo %v", req)
	if c.DriverName == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if c.VendorVersion == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}
	resp := &csi.GetPluginInfoResponse{
		Name:          c.DriverName,
		VendorVersion: c.VendorVersion,
	}
	return resp, nil
}

func (c *Csi) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	err := c.Agent.Probe()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &csi.ProbeResponse{}, nil
}

func (c *Csi) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
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
	resp := &csi.GetPluginCapabilitiesResponse{Capabilities: caps}
	return resp, nil
}

func (c *Csi) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (resp *csi.CreateVolumeResponse, finalErr error) {
	log.Printf("CreateVolume %v", req)
	if err := c.Agent.CreateVolume(req.Parameters["kind"], req.Name, req.CapacityRange.RequiredBytes); err != nil {
		log.Println(err)
		return nil, err
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
	log.Printf("DeleteVolume %v", req)
	err := c.Agent.DeleteVolume(req.VolumeId)
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
	return res, nil
}

func (c *Csi) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	log.Printf("ValidateVolumeCapabilities %v", req)
	return &csi.ValidateVolumeCapabilitiesResponse{}, nil
}

func (c *Csi) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	log.Printf("ControllerPublishVolume %v", req)
	err := c.Agent.PublishVolume(req.VolumeId, req.NodeId)
	if err != nil {
		return nil, err
	}
	return &csi.ControllerPublishVolumeResponse{
		PublishContext: req.VolumeContext,
	}, nil
}

func (c *Csi) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	log.Printf("ControllerUnpublishVolume %v", req)
	err := c.Agent.UnpublishVolume(req.VolumeId, req.NodeId)
	if err != nil {
		log.Printf("ERR:%s", err)
		return nil, err
	}
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (c *Csi) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	log.Printf("GetCapacity %v", req)
	cap, err := c.Agent.GetCapacity()
	if err != nil {
		log.Println("getCapacity ERR:", err)
		return &csi.GetCapacityResponse{}, nil
	}
	return &csi.GetCapacityResponse{
		AvailableCapacity: cap.Available,
		MaximumVolumeSize: wrapperspb.Int64(cap.MaximumVolumeSize),
		MinimumVolumeSize: wrapperspb.Int64(cap.MinimumVolumeSize),
	}, nil
}

func (c *Csi) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	log.Printf("ListVolumes %v", req)
	return &csi.ListVolumesResponse{}, nil
}

func (c *Csi) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	log.Printf("ControllerGetVolume %v", req)
	return &csi.ControllerGetVolumeResponse{}, nil
}

func (c *Csi) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	log.Printf("CreateSnapshot %v", req)
	return &csi.CreateSnapshotResponse{}, nil
}

func (c *Csi) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	log.Printf("DeleteSnapshot %v", req)
	return &csi.DeleteSnapshotResponse{}, nil
}

func (c *Csi) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	log.Printf("ListSnapshots %v", req)
	return &csi.ListSnapshotsResponse{}, nil
}

func (c *Csi) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	log.Printf("ControllerExpandVolume %v", req)
	err := c.Agent.ExpandVolume(req.VolumeId, req.CapacityRange.RequiredBytes)
	if err != nil {
		log.Println("expandVolume ERR:", err)
	}
	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes: req.CapacityRange.RequiredBytes,
	}, nil
}
