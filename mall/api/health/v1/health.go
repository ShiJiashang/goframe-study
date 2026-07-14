package v1

import "github.com/gogf/gf/v2/frame/g"

type DatabaseReq struct {
	g.Meta `path:"/health/database" method:"get" tags:"Health" summary:"数据库健康检查"`
}

type DatabaseRes struct {
	Status       string `json:"status" dc:"数据库状态"`
	ProductCount int    `json:"productCount" dc:"商品数量"`
	Database     string `json:"database" dc:"数据库状态"`
	CostMs       int64  `json:"costMs" dc:"耗时"`
}
