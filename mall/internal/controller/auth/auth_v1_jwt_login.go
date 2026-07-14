package auth

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"

	v1 "goframe-study/mall/api/auth/v1"
	"goframe-study/mall/internal/consts"
	"goframe-study/mall/internal/dao"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/internal/model/entity"
	"goframe-study/mall/utility/jwtutil"
)

const (
	// accessTTL / refreshTTL 与 lesson18 README 一致
	accessTTL  = 15 * time.Minute
	refreshTTL = 7 * 24 * time.Hour
)

// resolveRole 简易 role 判定：username=admin → admin，否则 user。
// 实际项目应把 role 存到用户表或独立的 roles 表。
func resolveRole(username string) string {
	if username == "admin" {
		return "admin"
	}
	return "user"
}

// findAndVerifyUser 从数据库查用户并用 bcrypt 验证密码。
func findAndVerifyUser(ctx context.Context, username, password string) (*entity.Users, error) {
	var user *entity.Users
	if err := dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Username, username).
		Scan(&user); err != nil {
		return nil, gerror.Wrap(err, "查询用户失败")
	}
	if user == nil {
		return nil, gerror.NewCode(consts.CodeAuthUnauthorized, "用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, gerror.NewCode(consts.CodeAuthUnauthorized, "用户名或密码错误")
	}
	if user.Status != 1 {
		return nil, gerror.NewCode(consts.CodeAuthUnauthorized, "账号已被禁用")
	}
	return user, nil
}

func (c *ControllerV1) JWTLogin(
	ctx context.Context,
	req *v1.JWTLoginReq,
) (res *v1.JWTLoginRes, err error) {
	user, err := findAndVerifyUser(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	secret, err := jwtutil.LoadSecret(ctx)
	if err != nil {
		return nil, err
	}

	role := resolveRole(user.Username)
	access, err := jwtutil.Create(int64(user.Id), role, model.TokenTypeAccess, accessTTL, secret)
	if err != nil {
		return nil, gerror.Wrap(err, "创建 access token 失败")
	}
	refresh, err := jwtutil.Create(int64(user.Id), role, model.TokenTypeRefresh, refreshTTL, secret)
	if err != nil {
		return nil, gerror.Wrap(err, "创建 refresh token 失败")
	}

	g.Log().Infof(ctx, "JWT 登录成功 userId=%d role=%s", user.Id, role)

	return &v1.JWTLoginRes{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(accessTTL.Seconds()),
		UserID:       int64(user.Id),
		Username:     user.Username,
		Role:         role,
	}, nil
}
