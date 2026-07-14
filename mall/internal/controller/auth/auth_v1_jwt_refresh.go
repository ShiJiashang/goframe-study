package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "goframe-study/mall/api/auth/v1"
	"goframe-study/mall/internal/consts"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/utility/jwtutil"
)

func (c *ControllerV1) JWTRefresh(
	ctx context.Context,
	req *v1.JWTRefreshReq,
) (res *v1.JWTRefreshRes, err error) {
	secret, err := jwtutil.LoadSecret(ctx)
	if err != nil {
		return nil, err
	}

	claims, err := jwtutil.Parse(req.RefreshToken, secret)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != model.TokenTypeRefresh {
		return nil, gerror.NewCode(consts.CodeAuthWrongTokenType, "必须使用 refresh token")
	}

	revoked, err := jwtutil.CheckRevoked(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, gerror.NewCode(consts.CodeAuthTokenRevoked, "refresh token 已失效")
	}

	access, err := jwtutil.Create(claims.UserID, claims.Role, model.TokenTypeAccess, accessTTL, secret)
	if err != nil {
		return nil, gerror.Wrap(err, "创建 access token 失败")
	}
	return &v1.JWTRefreshRes{
		AccessToken: access,
		ExpiresIn:   int64(accessTTL.Seconds()),
	}, nil
}
