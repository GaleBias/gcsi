package driver

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (d *Driver) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	fmt.Println()
	fmt.Println("************* GetPluginInfo of identity service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()
	return &csi.GetPluginInfoResponse{
		Name: d.name,
	}, nil
}

func (d *Driver) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	fmt.Println()
	fmt.Println("************* GetPluginCapabilities of identity service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			&csi.PluginCapability{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}

func (d *Driver) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	fmt.Println()
	fmt.Println("************* Probe of identity service have been called *************")
	fmt.Printf("req: %+#v\n", *req)
	fmt.Println()
	return &csi.ProbeResponse{
		Ready: &wrapperspb.BoolValue{
			Value: d.ready,
		},
	}, nil
}
