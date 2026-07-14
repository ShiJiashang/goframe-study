// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Orders is the golang structure of table orders for DAO operations like Where/Data.
type Orders struct {
	g.Meta    `orm:"table:orders, do:true"`
	Id        any         // è®¢å•ID
	OrderNo   any         // è®¢å•å·
	UserId    any         // ç”¨æˆ·ID
	TotalCent any         // è®¢å•æ€»é‡‘é¢ï¼Œå•ä½ä¸ºåˆ†
	Status    any         // çŠ¶æ€ï¼š1å¾…æ”¯ä»˜ï¼Œ2å·²æ”¯ä»˜ï¼Œ3å·²å–æ¶ˆ
	CreatedAt *gtime.Time // åˆ›å»ºæ—¶é—´
	UpdatedAt *gtime.Time // æ›´æ–°æ—¶é—´
}
