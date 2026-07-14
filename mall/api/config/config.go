// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package config

import (
	"context"

	"goframe-study/mall/api/config/v1"
)

type IConfigV1 interface {
	App(ctx context.Context, req *v1.AppReq) (res *v1.AppRes, err error)
}
