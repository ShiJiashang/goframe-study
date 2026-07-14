// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure for table users.
type Users struct {
	Id           uint64      `json:"id"           orm:"id"            description:"ç”¨æˆ·ID"`                   // ç”¨æˆ·ID
	Username     string      `json:"username"     orm:"username"      description:"ç™»å½•å"`                  // ç™»å½•å
	PasswordHash string      `json:"passwordHash" orm:"password_hash" description:"å¯†ç å“ˆå¸Œ"`               // å¯†ç å“ˆå¸Œ
	Status       uint        `json:"status"       orm:"status"        description:"çŠ¶æ€ï¼š1æ­£å¸¸ï¼Œ0ç¦ç”¨"` // çŠ¶æ€ï¼š1æ­£å¸¸ï¼Œ0ç¦ç”¨
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"åˆ›å»ºæ—¶é—´"`               // åˆ›å»ºæ—¶é—´
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"æ›´æ–°æ—¶é—´"`               // æ›´æ–°æ—¶é—´
}
