// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package auth

import (
	"context"

	"goframe-study/mall/api/auth/v1"
)

type IAuthV1 interface {
	SessionLogin(ctx context.Context, req *v1.SessionLoginReq) (res *v1.SessionLoginRes, err error)
	SessionMe(ctx context.Context, req *v1.SessionMeReq) (res *v1.SessionMeRes, err error)
	SessionLogout(ctx context.Context, req *v1.SessionLogoutReq) (res *v1.SessionLogoutRes, err error)

	JWTLogin(ctx context.Context, req *v1.JWTLoginReq) (res *v1.JWTLoginRes, err error)
	JWTRefresh(ctx context.Context, req *v1.JWTRefreshReq) (res *v1.JWTRefreshRes, err error)
	JWTLogout(ctx context.Context, req *v1.JWTLogoutReq) (res *v1.JWTLogoutRes, err error)
}
