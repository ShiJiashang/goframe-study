package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	v1 "goframe-study/mall/api/auth/v1"
)

func (c *ControllerV1) SessionLogout(
	ctx context.Context,
	req *v1.SessionLogoutReq,
) (res *v1.SessionLogoutRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	if err = r.Session.RemoveAll(); err != nil {
		return nil, gerror.WrapCode(gcode.CodeOperationFailed, err, "退出失败")
	}
	return &v1.SessionLogoutRes{}, nil
}
