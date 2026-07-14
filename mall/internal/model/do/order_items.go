// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// OrderItems is the golang structure of table order_items for DAO operations like Where/Data.
type OrderItems struct {
	g.Meta       `orm:"table:order_items, do:true"`
	Id           any         // è®¢å•æ˜Žç»†ID
	OrderId      any         // è®¢å•ID
	ProductId    any         // å•†å“ID
	ProductName  any         // ä¸‹å•æ—¶å•†å“åç§°å¿«ç…§
	PriceCent    any         // ä¸‹å•æ—¶å•ä»·ï¼Œå•ä½ä¸ºåˆ†
	Quantity     any         // è´­ä¹°æ•°é‡
	SubtotalCent any         // å°è®¡ï¼Œå•ä½ä¸ºåˆ†
	CreatedAt    *gtime.Time // åˆ›å»ºæ—¶é—´
}
