// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure of table users for DAO operations like Where/Data.
type Users struct {
	g.Meta       `orm:"table:users, do:true"`
	Id           any         // ç”¨æˆ·ID
	Username     any         // ç™»å½•å
	PasswordHash any         // å¯†ç å“ˆå¸Œ
	Status       any         // çŠ¶æ€ï¼š1æ­£å¸¸ï¼Œ0ç¦ç”¨
	CreatedAt    *gtime.Time // åˆ›å»ºæ—¶é—´
	UpdatedAt    *gtime.Time // æ›´æ–°æ—¶é—´
}
