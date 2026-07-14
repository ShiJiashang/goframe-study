package auth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	v1 "goframe-study/mall/api/auth/v1"
	"goframe-study/mall/internal/consts"
	"goframe-study/mall/utility/jwtutil"
)

// JWTLogout 撤销当前 Access Token（从 Authorization 头读取），
// 客户端如果同时提供 refreshToken，也一并撤销。
func (c *ControllerV1) JWTLogout(
	ctx context.Context,
	req *v1.JWTLogoutReq,
) (res *v1.JWTLogoutRes, err error) {
	r := ghttp.RequestFromCtx(ctx)

	// 1. 解析 Access Token
	authorization := r.Header.Get("Authorization")
	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, gerror.NewCode(consts.CodeAuthMissingBearer)
	}
	accessToken := strings.TrimSpace(parts[1])

	secret, err := jwtutil.LoadSecret(ctx)
	if err != nil {
		return nil, err
	}
	accessClaims, err := jwtutil.Parse(accessToken, secret)
	if err != nil {
		return nil, err
	}
	if err = jwtutil.Revoke(ctx, accessClaims); err != nil {
		return nil, gerror.Wrap(err, "撤销 access token 失败")
	}

	// 2. 如果客户端提供了 refreshToken，一并撤销
	if req.RefreshToken != "" {
		refreshClaims, parseErr := jwtutil.Parse(req.RefreshToken, secret)
		if parseErr == nil {
			if err = jwtutil.Revoke(ctx, refreshClaims); err != nil {
				g.Log().Warningf(ctx, "撤销 refresh token 失败 err=%v", err)
			}
		} else {
			g.Log().Warningf(ctx, "退出登录时 refresh token 无效 err=%v", parseErr)
		}
	}

	g.Log().Infof(ctx, "JWT 退出登录 userId=%d jti=%s", accessClaims.UserID, accessClaims.ID)
	return &v1.JWTLogoutRes{}, nil
}
