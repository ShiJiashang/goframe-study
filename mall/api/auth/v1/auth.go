package v1

import "github.com/gogf/gf/v2/frame/g"

// ============= Session 系列（lesson17）=============

type SessionLoginReq struct {
	g.Meta   `path:"/auth/session/login" method:"post" tags:"Auth" summary:"Session登录"`
	Username string `json:"username" v:"required#用户名不能为空"`
	Password string `json:"password" v:"required#密码不能为空"`
}

type SessionLoginRes struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
}

type SessionMeReq struct {
	g.Meta `path:"/auth/session/me" method:"get" tags:"Auth" summary:"当前Session用户"`
}

type SessionMeRes struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type SessionLogoutReq struct {
	g.Meta `path:"/auth/session/logout" method:"post" tags:"Auth" summary:"Session退出"`
}

type SessionLogoutRes struct{}

// ============= JWT 系列（lesson18）=============

type JWTLoginReq struct {
	g.Meta   `path:"/auth/jwt/login" method:"post" tags:"Auth" summary:"JWT登录"`
	Username string `json:"username" v:"required#用户名不能为空"`
	Password string `json:"password" v:"required#密码不能为空"`
}

type JWTLoginRes struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"` // 秒
	UserID       int64  `json:"userId"`
	Username     string `json:"username"`
	Role         string `json:"role"`
}

type JWTRefreshReq struct {
	g.Meta       `path:"/auth/jwt/refresh" method:"post" tags:"Auth" summary:"用RefreshToken换AccessToken"`
	RefreshToken string `json:"refreshToken" v:"required#refreshToken不能为空"`
}

type JWTRefreshRes struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

// JWTLogoutReq 从 Authorization 头读取当前 Access Token。
// 客户端可额外提供 refreshToken 一并撤销。
type JWTLogoutReq struct {
	g.Meta       `path:"/auth/jwt/logout" method:"post" tags:"Auth" summary:"JWT退出登录"`
	RefreshToken string `json:"refreshToken"` // 可选：一并撤销 refresh token
}

type JWTLogoutRes struct{}
