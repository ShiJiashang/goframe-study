package consts

import "github.com/gogf/gf/v2/errors/gcode"

var (
	CodeProductNotFound     = gcode.New(10001, "商品不存在", nil)
	CodeOrderStockNotEnough = gcode.New(20001, "库存不足", nil)

	// Session / 通用鉴权
	CodeAuthUnauthorized = gcode.New(30001, "未登录", nil)

	// JWT 相关
	CodeAuthMissingBearer   = gcode.New(30101, "缺少 Bearer token", nil)
	CodeAuthInvalidToken    = gcode.New(30102, "无效或已过期的 token", nil)
	CodeAuthWrongTokenType  = gcode.New(30103, "token 类型不匹配", nil)
	CodeAuthTokenRevoked    = gcode.New(30104, "token 已退出登录", nil)
	CodeAuthRevocationCheck = gcode.New(30105, "检查 token 状态失败", nil)
	CodeAuthAdminRequired   = gcode.New(30201, "需要管理员权限", nil)
)
