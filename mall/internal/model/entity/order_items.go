// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// OrderItems is the golang structure for table order_items.
type OrderItems struct {
	Id           uint64      `json:"id"           orm:"id"            description:"è®¢å•æ˜Žç»†ID"`                 // è®¢å•æ˜Žç»†ID
	OrderId      uint64      `json:"orderId"      orm:"order_id"      description:"è®¢å•ID"`                       // è®¢å•ID
	ProductId    uint64      `json:"productId"    orm:"product_id"    description:"å•†å“ID"`                       // å•†å“ID
	ProductName  string      `json:"productName"  orm:"product_name"  description:"ä¸‹å•æ—¶å•†å“åç§°å¿«ç…§"`    // ä¸‹å•æ—¶å•†å“åç§°å¿«ç…§
	PriceCent    uint64      `json:"priceCent"    orm:"price_cent"    description:"ä¸‹å•æ—¶å•ä»·ï¼Œå•ä½ä¸ºåˆ†"` // ä¸‹å•æ—¶å•ä»·ï¼Œå•ä½ä¸ºåˆ†
	Quantity     uint        `json:"quantity"     orm:"quantity"      description:"è´­ä¹°æ•°é‡"`                   // è´­ä¹°æ•°é‡
	SubtotalCent uint64      `json:"subtotalCent" orm:"subtotal_cent" description:"å°è®¡ï¼Œå•ä½ä¸ºåˆ†"`          // å°è®¡ï¼Œå•ä½ä¸ºåˆ†
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"åˆ›å»ºæ—¶é—´"`                   // åˆ›å»ºæ—¶é—´
}
