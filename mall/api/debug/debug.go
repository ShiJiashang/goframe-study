// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package debug

import (
	"context"

	"goframe-study/mall/api/debug/v1"
)

type IDebugV1 interface {
	Error(ctx context.Context, req *v1.ErrorReq) (res *v1.ErrorRes, err error)
	PaymentParse(ctx context.Context, req *v1.PaymentParseReq) (res *v1.PaymentParseRes, err error)
}
