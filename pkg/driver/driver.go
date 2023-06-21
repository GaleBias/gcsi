package driver

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"

	"csi/pkg/storage"

	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	evs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2"
)

const (
	DefaultName string = "csi.gale.com"
)

type Driver struct {
	name     string
	endpoint string

	srv     *grpc.Server
	storage *evs.EvsClient
	ecs     *ecs.EcsClient

	ready bool
}
type InputParams struct {
	Name     string
	Endpoint string

	Project string
	Region  string
	AK      string
	SK      string
}

func NewDriver(params InputParams) (*Driver, error) {
	if params.AK == "" || params.SK == "" {
		return nil, fmt.Errorf("AK and SK must be specified")
	}

	// client := storage.NewEvsClient(params.Project, params.Region, params.AK, params.SK)
	storage, ecs := storage.NewEvsClient(params.Project, params.Region, params.AK, params.SK)

	return &Driver{
		name:     params.Name,
		endpoint: params.Endpoint,
		storage:  storage,
		ecs:      ecs,
	}, nil
}

func (d *Driver) Run() error {
	url, err := url.Parse(d.endpoint)
	if err != nil {
		return fmt.Errorf("parasing the endpoint %s\n", err.Error())
	}

	if url.Scheme != "unix" {
		return fmt.Errorf("only supported scheme is unix, but provided %s\n", url.Scheme)
	}

	grpcAddr := path.Join(url.Host, filepath.FromSlash(url.Path))
	if url.Host == "" {
		grpcAddr = filepath.FromSlash(url.Path)
	}
	if err := os.Remove(grpcAddr); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove listen sock %s\n", err.Error())
	}

	listener, err := net.Listen(url.Scheme, grpcAddr)
	if err != nil {
		return fmt.Errorf("listen failed %s\n", err.Error())
	}

	d.srv = grpc.NewServer()

	csi.RegisterIdentityServer(d.srv, d)
	csi.RegisterControllerServer(d.srv, d)
	csi.RegisterNodeServer(d.srv, d)

	d.ready = true

	return d.srv.Serve(listener)
}
