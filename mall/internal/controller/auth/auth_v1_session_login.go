package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	v1 "goframe-study/mall/api/auth/v1"
	"goframe-study/mall/internal/consts"
)

func (c *ControllerV1) SessionLogin(
	ctx context.Context,
	req *v1.SessionLoginReq,
) (res *v1.SessionLoginRes, err error) {
	// 本课固定学习账号；下一课接入数据库 + bcrypt
	if req.Username != "demo" || req.Password != "demo123" {
		return nil, gerror.NewCode(
			consts.CodeAuthUnauthorized,
			"用户名或密码错误",
		)
	}

	r := ghttp.RequestFromCtx(ctx)
	if err = r.Session.SetMap(map[string]any{
		"userId":   int64(1),
		"username": "demo",
		"role":     "admin",
	}); err != nil {
		return nil, gerror.WrapCode(
			gcode.CodeOperationFailed,
			err,
			"保存登录状态失败",
		)
	}

	return &v1.SessionLoginRes{
		UserID:   1,
		Username: "demo",
	}, nil
}
