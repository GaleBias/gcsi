package main

import (
	"csi/pkg/driver"
	"flag"
	"fmt"
)

func main() {
	var (
		endpoint = flag.String("endpoint", "unix:///var/lib/csi/sockets/csi.sock", "Endpoint gRPC server server would run at")

		project = flag.String("project", "2fb7832179084025b2eadab146ad3cb0", "project of cloud platform")
		region  = flag.String("region", "ap-southeast-3", "region where the volumes are going to be provisioned")

		ak = flag.String("ak", "xxxxxxxxx", "ak of the storage provider")
		sk = flag.String("sk", "xxxxxxxxx", "ak of the storage provider")
		// zone: ap-southeast-3a、b、c、d、e
	)
	flag.Parse()
	fmt.Println(*endpoint, *project, *region, *ak, *sk)

	driver, err := driver.NewDriver(driver.InputParams{
		Name:     driver.DefaultName,
		Endpoint: *endpoint,
		Project:  *project,
		Region:   *region,
		AK:       *ak,
		SK:       *sk,
	})
	if err != nil {
		fmt.Printf("Error %s, creating new instance of driver\n", err.Error())
	}
	if err := driver.Run(); err != nil {
		fmt.Printf("Error %s, running the driver", err.Error())
	}
}
