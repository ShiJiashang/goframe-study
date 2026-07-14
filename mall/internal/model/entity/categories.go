// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Categories is the golang structure for table categories.
type Categories struct {
	Id        uint64      `json:"id"        orm:"id"         description:"åˆ†ç±»ID"`                   // åˆ†ç±»ID
	Name      string      `json:"name"      orm:"name"       description:"åˆ†ç±»åç§°"`               // åˆ†ç±»åç§°
	Sort      int         `json:"sort"      orm:"sort"       description:"æŽ’åºå€¼"`                  // æŽ’åºå€¼
	Status    uint        `json:"status"    orm:"status"     description:"çŠ¶æ€ï¼š1å¯ç”¨ï¼Œ0åœç”¨"` // çŠ¶æ€ï¼š1å¯ç”¨ï¼Œ0åœç”¨
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"åˆ›å»ºæ—¶é—´"`               // åˆ›å»ºæ—¶é—´
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"æ›´æ–°æ—¶é—´"`               // æ›´æ–°æ—¶é—´
}
