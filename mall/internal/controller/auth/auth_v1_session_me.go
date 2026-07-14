package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	v1 "goframe-study/mall/api/auth/v1"
	"goframe-study/mall/internal/consts"
)

func (c *ControllerV1) SessionMe(
	ctx context.Context,
	req *v1.SessionMeReq,
) (res *v1.SessionMeRes, err error) {
	r := ghttp.RequestFromCtx(ctx)

	userID, err := r.Session.Get("userId")
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeOperationFailed, err, "读取Session失败")
	}
	if userID == nil || userID.IsNil() || userID.Int64() <= 0 {
		return nil, gerror.NewCode(consts.CodeAuthUnauthorized)
	}

	username, _ := r.Session.Get("username", "")
	role, _ := r.Session.Get("role", "")

	return &v1.SessionMeRes{
		UserID:   userID.Int64(),
		Username: username.String(),
		Role:     role.String(),
	}, nil
}
