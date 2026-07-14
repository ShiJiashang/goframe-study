// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Products is the golang structure of table products for DAO operations like Where/Data.
type Products struct {
	g.Meta     `orm:"table:products, do:true"`
	Id         any         // å•†å“ID
	CategoryId any         // åˆ†ç±»ID
	Name       any         // å•†å“åç§°
	PriceCent  any         // ä»·æ ¼ï¼Œå•ä½ä¸ºåˆ†
	Stock      any         // åº“å­˜
	Status     any         // çŠ¶æ€ï¼š1ä¸Šæž¶ï¼Œ0ä¸‹æž¶
	CreatedAt  *gtime.Time // åˆ›å»ºæ—¶é—´
	UpdatedAt  *gtime.Time // æ›´æ–°æ—¶é—´
}
