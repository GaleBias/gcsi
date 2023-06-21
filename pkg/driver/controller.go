package driver

import (
	"context"
	"fmt"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	ecsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
	evsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (d *Driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	fmt.Println()
	fmt.Println("************* CreateVolume of controller service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()

	// name is present
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "createVolume must be called with a req name")
	}
	fmt.Println(req.Name)

	// extract required memory
	// make sure th value here is not less than or = 0
	// requiredBytes is not more than limitedBytes
	sizeBytes := req.CapacityRange.GetRequiredBytes()

	// make sure volume capabilities have been specified
	if req.VolumeCapabilities == nil || len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapabilities have not been specified")
	}

	// validate volume capabilities
	// make sure accessMode that has been specified by the PVC is actually supported by SP
	// make sure volumeMode that has been specified by the PVC is supported by us

	// create the request struct
	chargingModeBssParam := evsModel.GetBssParamForCreateVolumeChargingModeEnum().POST_PAID
	passthrough := "true"
	count, gb := int32(1), 1024*1024*1024
	volReq := &evsModel.CreateVolumeRequest{
		Body: &evsModel.CreateVolumeRequestBody{
			Volume: &evsModel.CreateVolumeOption{
				AvailabilityZone: "ap-southeast-3b",
				Name:             &req.Name,
				Size:             int32(sizeBytes / int64(gb)),
				Count:            &count,
				VolumeType:       evsModel.GetCreateVolumeOptionVolumeTypeEnum().SAS,
				Metadata: map[string]string{
					"hw:passthrough": passthrough,
				},
			},
			BssParam: &evsModel.BssParamForCreateVolume{
				ChargingMode: &chargingModeBssParam,
			},
		},
	}

	// check if VolumeContentSource is specified
	// you will also have to make sure that this snapshot is actually present
	// if req.VolumeContentSource.GetSnapshot().SnapshotId != "" {
	// 	volReq.Body.Volume.ImageRef = &req.VolumeContentSource.GetSnapshot().SnapshotId
	// }

	// if this user have not exceeded the limit
	// if this user can provision the requested amount etc

	// handle AccessibilityRequirement

	// actually all DO api to create the volume
	volResp, err := d.storage.CreateVolume(volReq)
	if err != nil {
		fmt.Printf("%+v\n", volResp)
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed provisioning the volume, error: %s\n", err.Error()))
	}
	fmt.Println((*volResp.VolumeIds)[0])

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			CapacityBytes: sizeBytes,
			VolumeId:      (*volResp.VolumeIds)[0],
			// specify content source, but only in cases where its specified in the PVC
			VolumeContext: map[string]string{
				"hw:passthrough": passthrough,
			},
		},
	}, nil
}

func (d *Driver) DeleteVolume(context.Context, *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	fmt.Println()
	fmt.Println("************* ControllerPublishVolume of controller service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()

	// check if volumeId is present and volume is available on SP
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "VolumeId is mandatory in ControllerPublishVolume request")
	}

	// check nodeId is set, and node is actually present on SP
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "nodeId is mandatory in ControllerPublishVolume request")
	}

	passthrough := req.GetVolumeContext()["hw:passthrough"]

	_, err := d.ecs.AttachServerVolume(&ecsModel.AttachServerVolumeRequest{
		ServerId: req.NodeId,
		Body: &ecsModel.AttachServerVolumeRequestBody{
			VolumeAttachment: &ecsModel.AttachServerVolumeOption{
				VolumeId:      req.VolumeId,
				Hwpassthrough: &passthrough,
			},
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("volume %s attach to ecs %s failed, err: %s\n", req.VolumeId, req.NodeId, err.Error()))
	}

	var ecsAttachIdentity string
	if err := d.getEvsAttachIdentity(req.VolumeId, &ecsAttachIdentity); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("waiting volume %s attach to ecs %s , 挂载点为: %s, err: %s\n", req.VolumeId, req.NodeId, ecsAttachIdentity, err.Error()))
	}

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			"csi.gale.com/volume-name": ecsAttachIdentity,
		},
	}, nil
}

func (d *Driver) getEvsAttachIdentity(volumeId string, identity *string) error {
	err := wait.Poll(1*time.Second, 5*time.Minute, func() (done bool, err error) {

		vol, err := d.storage.ShowVolume(&model.ShowVolumeRequest{VolumeId: volumeId})
		if err != nil {
			return false, fmt.Errorf("show volume %s failed,err:%s\n", volumeId, err.Error())
		}

		if passthrough, ok := vol.Volume.Metadata["hw:passthrough"]; ok && passthrough == "true" {
			*identity = fmt.Sprintf("scsi-3%s", *vol.Volume.Wwn)
			if *identity == "" {
				return false, fmt.Errorf("未获取到挂载点信息")
			}
			return true, nil
		}
		*identity = fmt.Sprintf("virtio-%s", volumeId[:20])
		if *identity == "" {
			return false, fmt.Errorf("未获取到挂载点信息")
		}
		return true, nil
	})
	return err
}

func (d *Driver) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, nil
}

func (d *Driver) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, nil
}

func (d *Driver) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, nil
}

func (d *Driver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	fmt.Println()
	fmt.Println("************* ControllerGetCapabilities of controller service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()

	capabilities := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	}

	caps := []*csi.ControllerServiceCapability{}
	for _, c := range capabilities {
		caps = append(caps, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: c,
				},
			},
		})
	}

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}, nil
}

func (d *Driver) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, nil
}

func (d *Driver) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, nil
}

func (d *Driver) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, nil
}

func (d *Driver) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) ControllerGetVolume(context.Context, *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, nil
}
