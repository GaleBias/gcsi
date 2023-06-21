package storage

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"

	evs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2"
	evsRegion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/region"

	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsRegion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"
)

func NewEvsClient(pro, re, ak, sk string) (*evs.EvsClient, *ecs.EcsClient) {

	auth := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		WithProjectId(pro).
		Build()

	storage := evs.NewEvsClient(
		evs.EvsClientBuilder().
			WithRegion(evsRegion.ValueOf(re)).
			WithCredential(auth).
			Build())

	ecs := ecs.NewEcsClient(
		ecs.EcsClientBuilder().
			WithRegion(ecsRegion.ValueOf("ap-southeast-3")).
			WithCredential(auth).
			Build())
	return storage, ecs
}
