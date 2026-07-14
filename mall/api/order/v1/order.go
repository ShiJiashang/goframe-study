package v1

import "github.com/gogf/gf/v2/frame/g"

type CreateReq struct {
	g.Meta    `path:"/orders" method:"post" tags:"Order" summary:"创建订单"`
	UserID    int64 `json:"userId" v:"min:1#用户ID不正确"`
	ProductID int64 `json:"productId" v:"min:1#商品ID不正确"`
	Quantity  int   `json:"quantity" v:"between:1,100#购买数量必须在1到100之间"`
}

type CreateRes struct {
	OrderID   int64  `json:"orderId"`
	OrderNo   string `json:"orderNo"`
	TotalCent int64  `json:"totalCent"`
}
