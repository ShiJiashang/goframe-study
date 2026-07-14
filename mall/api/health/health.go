// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package health

import (
	"context"

	"goframe-study/mall/api/health/v1"
)

type IHealthV1 interface {
	Database(ctx context.Context, req *v1.DatabaseReq) (res *v1.DatabaseRes, err error)
}
