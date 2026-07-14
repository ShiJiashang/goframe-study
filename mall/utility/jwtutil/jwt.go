// Package jwtutil 封装 JWT 的签发、解析与撤销。
//
// 依赖：
//   - github.com/golang-jwt/jwt/v5：官方社区维护的 JWT 库
//   - g.Cfg().GetEffective：按“命令行参数 > 环境变量 > 配置文件”优先级读取密钥
//   - g.Redis()：撤销列表（黑名单）存储
package jwtutil

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/golang-jwt/jwt/v5"

	"goframe-study/mall/internal/consts"
	"goframe-study/mall/internal/model"
)

const revokedKeyPrefix = "mall:jwt:revoked:"

// LoadSecret 按 GoFrame 配置优先级读取 JWT 密钥。
// 配置键 auth.jwtSecret，对应环境变量 AUTH_JWTSECRET。
func LoadSecret(ctx context.Context) ([]byte, error) {
	value, err := g.Cfg().GetEffective(ctx, "auth.jwtSecret")
	if err != nil {
		return nil, gerror.Wrap(err, "读取 JWT 密钥失败")
	}
	if value.IsEmpty() {
		return nil, gerror.New("JWT 密钥未配置")
	}
	return value.Bytes(), nil
}

// Create 签发一张 JWT。tokenType 只能是 model.TokenTypeAccess 或 model.TokenTypeRefresh。
func Create(userID int64, role, tokenType string, ttl time.Duration, secret []byte) (string, error) {
	now := time.Now()
	claims := model.Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        guid.S(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// Parse 解析并验证 JWT，只允许 HS256 算法。
// 不检查撤销状态，撤销由 CheckRevoked 单独判断。
func Parse(tokenString string, secret []byte) (*model.Claims, error) {
	claims := new(model.Claims)
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(_ *jwt.Token) (any, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		return nil, gerror.NewCode(consts.CodeAuthInvalidToken)
	}
	return claims, nil
}

// Revoke 把 token 的 jti 写入 Redis 撤销列表，TTL 与 token 原本的剩余寿命一致。
// token 已过期时直接返回 nil，不用再写。
func Revoke(ctx context.Context, claims *model.Claims) error {
	if claims == nil || claims.ExpiresAt == nil {
		return nil
	}
	seconds := int64(time.Until(claims.ExpiresAt.Time).Seconds())
	if seconds <= 0 {
		return nil
	}
	key := revokedKeyPrefix + claims.ID
	return g.Redis().SetEX(ctx, key, 1, seconds)
}

// CheckRevoked 检查 jti 是否已被撤销。
// 返回 revoked=true 表示 token 已失效；err 表示查询过程本身出错。
func CheckRevoked(ctx context.Context, jti string) (bool, error) {
	value, err := g.Redis().Get(ctx, revokedKeyPrefix+jti)
	if err != nil {
		return false, gerror.WrapCode(consts.CodeAuthRevocationCheck, err, "检查 token 状态失败")
	}
	return !value.IsNil(), nil
}
