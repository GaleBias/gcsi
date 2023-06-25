package driver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
)

func (d *Driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	fmt.Println()
	fmt.Println("************* NodeStageVolume of node service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "VolumeID must be present in the NodeStageVolumeReq")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "StagingTargetPath must be present in the NodeSVolReq")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "VolumeCaps must be present in the NodeSVolReq")
	}
	fmt.Println("req.VolumeCapability.GetAccessMode() *************", req.VolumeCapability.GetAccessMode())
	fmt.Println("req.VolumeCapability.GetAccessType() *************", req.VolumeCapability.GetAccessType())
	fmt.Println("req.VolumeCapability.GetMount() ******************", req.VolumeCapability.GetMount())

	switch req.VolumeCapability.AccessType.(type) {
	case *csi.VolumeCapability_Block:
		return &csi.NodeStageVolumeResponse{}, nil
	}

	source, target := "", req.StagingTargetPath
	if vol, ok := req.PublishContext["csi.gale.com/volume-name"]; !ok {
		return nil, status.Error(codes.InvalidArgument, "VolumeName is not present in the publish context of request")
	} else {
		source = fmt.Sprintf("/dev/disk/by-id/%s", vol)
	}

	mnt := req.VolumeCapability.GetMount()
	fsType := "ext4"
	if mnt.FsType != "" {
		fsType = mnt.FsType
	}

	// format disk
	if err := formatAndMakeFs(source, fsType); err != nil {
		fmt.Printf("unable to create fs error %s\n", err.Error())
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to create fs error %s\n", err.Error()))
	}

	// mount source to target
	if err := mount(source, target, fsType, mnt.MountFlags); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error %s, mounting the source %s to target %s\n", err.Error(), source, target))
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func mount(source, target, fsType string, options []string) error {
	_, err := exec.LookPath("mount")
	if err != nil {
		return fmt.Errorf("unable to find the mount cmd errors is %s", err.Error())
	}
	if err := os.MkdirAll(target, 0777); err != nil {
		return fmt.Errorf("error: %s, creating the target dir\n", err.Error())
	}

	mountArgs := []string{}
	mountArgs = append(mountArgs, "-t", fsType)

	// check of options and then append them at the end of the mount command
	if len(options) > 0 {
		mountArgs = append(mountArgs, "-o", strings.Join(options, ","))
	}

	mountArgs = append(mountArgs, source)
	mountArgs = append(mountArgs, target)

	out, err := exec.Command("mount", mountArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error %s, mounting the source %s to tar %s. Output: %s\n", err.Error(), source, target, out)
	}
	return nil
}

func formatAndMakeFs(source, fsType string) error {
	mkfsCmd := fmt.Sprintf("mkfs.%s", fsType)

	_, err := exec.LookPath(mkfsCmd)
	if err != nil {
		return fmt.Errorf("unable to find the mkfs (%s) utiltiy errors is %s", mkfsCmd, err.Error())
	}

	// actually run mkfs.ext4 -F source
	mkfsArgs := []string{"-F", source}

	out, err := exec.Command(mkfsCmd, mkfsArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("create fs command failed output: %s, and err: %s\n", out, err.Error())
	}
	return nil

	// mkfsCmd := fmt.Sprintf("mkfs.%s", fsType)
	// _, err := exec.LookPath(mkfsCmd)
	// if err != nil {
	// 	return fmt.Errorf("unable to find the mkfs (%s) errors is %s", mkfsCmd, err.Error())
	// }

	// mkfsArgs := []string{"-F", source}
	// out, err := exec.Command(mkfsCmd, mkfsArgs...).CombinedOutput()
	// if err != nil {
	// 	return fmt.Errorf("create fs command failed output: %s, and err: %s\n", out, err.Error())
	// }
	// return nil
}

func (d *Driver) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	fmt.Println()
	fmt.Println("************* NodePublishVolume of node service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()

	// make sure the required fields are set and not empty

	options := []string{"bind"}
	if req.Readonly {
		options = append(options, "rwx")
	}

	// get req.VolumeCaps and make sure that you handle request for block mode as well
	// here we are just handling request for filesystem mode
	// in case of block mode, the source is going to be the device dir where volume was attached form ControllerPubVolume RPC

	fsType := "ext4"
	if req.VolumeCapability.GetMount().FsType != "" {
		fsType = req.VolumeCapability.GetMount().FsType
	}

	source, target := req.StagingTargetPath, req.TargetPath

	// we want to run mount -t fsType source target -o bind,rw
	err := mount(source, target, fsType, options)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error %s, mounting the volume from staging dir to target dir", err.Error()))
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (d *Driver) NodeUnpublishVolume(context.Context, *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, nil
}

func (d *Driver) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	fmt.Println()
	fmt.Println("************* GetPluginCapabilities of identity service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			&csi.NodeServiceCapability{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (d *Driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	fmt.Println()
	fmt.Println("************* GetPluginCapabilities of identity service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()

	var ecsId string

	name := os.Getenv("nodeName")
	es, err := d.ecs.ListServersDetails(&model.ListServersDetailsRequest{
		Name: &name,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("get ecs id failed, err: %s\n", err.Error()))
	}
	for _, ecs := range *es.Servers {
		if ecs.Name == name {
			ecsId = ecs.Id
		}
	}

	return &csi.NodeGetInfoResponse{
		NodeId:            ecsId,
		MaxVolumesPerNode: 5,
		AccessibleTopology: &csi.Topology{
			Segments: map[string]string{
				"zone": "ap-southeast-3b",
			},
		},
	}, nil
}
