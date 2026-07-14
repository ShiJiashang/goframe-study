package middleware

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// sessionAuthResult 内部结构，只在 middleware 包内使用。
type sessionAuthResult struct {
	userID   int64
	username string
	role     string
}

// tryParseSession 尝试从 Session 读取登录信息。
// ok=true 表示 Session 里确实有登录用户；否则 err 描述失败原因。
// 该函数不写入错误，也不 Next，供 AuthGate 组合调用。
func tryParseSession(r *ghttp.Request) (*sessionAuthResult, bool, error) {
	userID, err := r.Session.Get("userId")
	if err != nil {
		g.Log().Warningf(r.Context(), "读取 Session 失败 err=%v", err)
		return nil, false, gerror.WrapCode(gcode.CodeOperationFailed, err, "读取登录状态失败")
	}
	if userID == nil || userID.IsNil() || userID.Int64() <= 0 {
		return nil, false, nil
	}
	username, _ := r.Session.Get("username", "")
	role, _ := r.Session.Get("role", "")
	return &sessionAuthResult{
		userID:   userID.Int64(),
		username: username.String(),
		role:     role.String(),
	}, true, nil
}
