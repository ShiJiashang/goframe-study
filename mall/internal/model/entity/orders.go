// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Orders is the golang structure for table orders.
type Orders struct {
	Id        uint64      `json:"id"        orm:"id"         description:"è®¢å•ID"`                                      // è®¢å•ID
	OrderNo   string      `json:"orderNo"   orm:"order_no"   description:"è®¢å•å·"`                                     // è®¢å•å·
	UserId    uint64      `json:"userId"    orm:"user_id"    description:"ç”¨æˆ·ID"`                                      // ç”¨æˆ·ID
	TotalCent uint64      `json:"totalCent" orm:"total_cent" description:"è®¢å•æ€»é‡‘é¢ï¼Œå•ä½ä¸ºåˆ†"`                // è®¢å•æ€»é‡‘é¢ï¼Œå•ä½ä¸ºåˆ†
	Status    uint        `json:"status"    orm:"status"     description:"çŠ¶æ€ï¼š1å¾…æ”¯ä»˜ï¼Œ2å·²æ”¯ä»˜ï¼Œ3å·²å–æ¶ˆ"` // çŠ¶æ€ï¼š1å¾…æ”¯ä»˜ï¼Œ2å·²æ”¯ä»˜ï¼Œ3å·²å–æ¶ˆ
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"åˆ›å»ºæ—¶é—´"`                                  // åˆ›å»ºæ—¶é—´
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"æ›´æ–°æ—¶é—´"`                                  // æ›´æ–°æ—¶é—´
}
