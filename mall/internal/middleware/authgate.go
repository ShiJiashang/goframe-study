package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"goframe-study/mall/internal/consts"
	"goframe-study/mall/internal/model"
)

// authScopePath 定义需要鉴权的路径前缀。
// 除白名单路径外，命中此前缀才做 Session/JWT 校验。
const authScopePath = "/admin/"

// authWhitelist 定义不需要鉴权的公开接口。
// 目前登录、刷新等 auth 相关接口都在 /auth/ 下，天然放行。
var authWhitelist = []string{}

// writeAuthContext 把当前登录身份写入 Request 参数，供 Controller 读取。
func writeAuthContext(r *ghttp.Request, userID int64, role, username, jti string, claims *model.Claims) {
	if userID > 0 {
		r.SetParam("currentUserId", userID)
		r.SetParam("userId", userID) // 兼容 JWT 章节的命名
	}
	if role != "" {
		r.SetParam("role", role)
		r.SetParam("currentUserRole", role)
	}
	if username != "" {
		r.SetParam("currentUsername", username)
	}
	if jti != "" {
		r.SetParam("currentTokenJTI", jti)
	}
	if claims != nil {
		r.SetParam("tokenClaims", claims)
	}
}

// AuthGate 是 Session 与 JWT 二选一的组合鉴权。
//   - 不在 /admin/ 前缀内：直接放行；
//   - Session 已登录：放行；
//   - 请求带 Authorization Bearer 且 JWT 合法：放行；
//   - 上述都不满足：返回未登录错误。
//
// 顺序为“Session 优先，JWT 兜底”，方便浏览器场景与 App 场景共用一套后端。
func AuthGate(r *ghttp.Request) {
	if !strings.HasPrefix(r.URL.Path, authScopePath) {
		r.Middleware.Next()
		return
	}
	for _, prefix := range authWhitelist {
		if strings.HasPrefix(r.URL.Path, prefix) {
			r.Middleware.Next()
			return
		}
	}

	if session, ok, err := tryParseSession(r); err == nil && ok {
		writeAuthContext(r, session.userID, session.role, session.username, "", nil)
		r.Middleware.Next()
		return
	} else if err != nil {
		r.SetError(err)
		return
	}

	jwtResult, hasHeader, jwtErr := tryParseJWT(r)
	if hasHeader {
		if jwtErr != nil {
			r.SetError(jwtErr)
			return
		}
		writeAuthContext(r, jwtResult.userID, jwtResult.role, "", jwtResult.tokenJTI, jwtResult.claims)
		r.Middleware.Next()
		return
	}

	r.SetError(gerror.NewCode(consts.CodeAuthUnauthorized))
}

// AdminOnly 校验当前身份必须是 admin，仅对 /admin/* 路径生效。
// 必须挂在 AuthGate 之后（依赖 role 参数已被写入）。
func AdminOnly(r *ghttp.Request) {
	if !strings.HasPrefix(r.URL.Path, authScopePath) {
		r.Middleware.Next()
		return
	}
	if r.Get("role").String() != "admin" {
		r.SetError(gerror.NewCode(consts.CodeAuthAdminRequired))
		return
	}
	r.Middleware.Next()
}
