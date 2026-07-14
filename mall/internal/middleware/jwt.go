package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"goframe-study/mall/internal/consts"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/utility/jwtutil"
)

// bearerPrefix 遵循 RFC 6750，大小写不敏感。
const bearerPrefix = "Bearer"

// jwtAuthResult 保存 JWT 中间件解析出的身份，用于跨中间件传递。
type jwtAuthResult struct {
	userID   int64
	role     string
	tokenJTI string
	claims   *model.Claims
}

// tryParseJWT 从 Authorization 头解析 JWT，校验签名、算法、类型和撤销状态。
// ok=true 表示解析出可用身份；hasHeader=true 表示至少提供了 Authorization。
// 供 AuthGate 组合调用；单独用见 JWTAuth。
func tryParseJWT(r *ghttp.Request) (result *jwtAuthResult, hasHeader bool, err error) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return nil, false, nil
	}

	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], bearerPrefix) {
		return nil, true, gerror.NewCode(consts.CodeAuthMissingBearer)
	}
	tokenString := strings.TrimSpace(parts[1])
	if tokenString == "" {
		return nil, true, gerror.NewCode(consts.CodeAuthMissingBearer)
	}

	ctx := r.Context()
	secret, err := jwtutil.LoadSecret(ctx)
	if err != nil {
		return nil, true, err
	}
	claims, err := jwtutil.Parse(tokenString, secret)
	if err != nil {
		return nil, true, err
	}
	if claims.TokenType != model.TokenTypeAccess {
		return nil, true, gerror.NewCode(consts.CodeAuthWrongTokenType, "该 token 不能访问业务接口")
	}

	revoked, err := jwtutil.CheckRevoked(ctx, claims.ID)
	if err != nil {
		return nil, true, err
	}
	if revoked {
		return nil, true, gerror.NewCode(consts.CodeAuthTokenRevoked)
	}

	return &jwtAuthResult{
		userID:   claims.UserID,
		role:     claims.Role,
		tokenJTI: claims.ID,
		claims:   claims,
	}, true, nil
}

// JWTAuth 是仅使用 JWT 的鉴权中间件（保留独立使用能力）。
// 一般路由建议用 AuthGate（Session 或 JWT 二选一）。
func JWTAuth(r *ghttp.Request) {
	result, hasHeader, err := tryParseJWT(r)
	if err != nil {
		r.SetError(err)
		return
	}
	if !hasHeader || result == nil {
		r.SetError(gerror.NewCode(consts.CodeAuthMissingBearer))
		return
	}
	writeAuthContext(r, result.userID, result.role, "", result.tokenJTI, result.claims)
	r.Middleware.Next()
}
