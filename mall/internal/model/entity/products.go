// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Products is the golang structure for table products.
type Products struct {
	Id         uint64      `json:"id"         orm:"id"          description:"å•†å“ID"`                   // å•†å“ID
	CategoryId uint64      `json:"categoryId" orm:"category_id" description:"åˆ†ç±»ID"`                   // åˆ†ç±»ID
	Name       string      `json:"name"       orm:"name"        description:"å•†å“åç§°"`               // å•†å“åç§°
	PriceCent  uint64      `json:"priceCent"  orm:"price_cent"  description:"ä»·æ ¼ï¼Œå•ä½ä¸ºåˆ†"`      // ä»·æ ¼ï¼Œå•ä½ä¸ºåˆ†
	Stock      uint        `json:"stock"      orm:"stock"       description:"åº“å­˜"`                     // åº“å­˜
	Status     uint        `json:"status"     orm:"status"      description:"çŠ¶æ€ï¼š1ä¸Šæž¶ï¼Œ0ä¸‹æž¶"` // çŠ¶æ€ï¼š1ä¸Šæž¶ï¼Œ0ä¸‹æž¶
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"åˆ›å»ºæ—¶é—´"`               // åˆ›å»ºæ—¶é—´
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"æ›´æ–°æ—¶é—´"`               // æ›´æ–°æ—¶é—´
}
