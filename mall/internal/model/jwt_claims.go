package model

import "github.com/golang-jwt/jwt/v5"

// TokenType 用于区分 Access Token 与 Refresh Token。
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims 是 mall 项目的 JWT 载荷类型。
//
// 内嵌 jwt.RegisteredClaims 之后会自动带上 iat/exp/jti 等标准字段。
type Claims struct {
	UserID    int64  `json:"userId"`
	Role      string `json:"role"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}
