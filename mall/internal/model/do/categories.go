// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Categories is the golang structure of table categories for DAO operations like Where/Data.
type Categories struct {
	g.Meta    `orm:"table:categories, do:true"`
	Id        any         // åˆ†ç±»ID
	Name      any         // åˆ†ç±»åç§°
	Sort      any         // æŽ’åºå€¼
	Status    any         // çŠ¶æ€ï¼š1å¯ç”¨ï¼Œ0åœç”¨
	CreatedAt *gtime.Time // åˆ›å»ºæ—¶é—´
	UpdatedAt *gtime.Time // æ›´æ–°æ—¶é—´
}
