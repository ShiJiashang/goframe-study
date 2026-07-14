package v1

import "github.com/gogf/gf/v2/frame/g"

type ProductItem struct {
	ID         int64  `json:"id"`
	CategoryID int64  `json:"categoryId"`
	Name       string `json:"name"`
	PriceCent  int64  `json:"priceCent"`
	Stock      int    `json:"stock"`
	Status     int    `json:"status"`
}

type CreateReq struct {
	g.Meta     `path:"/admin/products" method:"post" tags:"Product" summary:"新增商品"`
	CategoryID int64  `json:"categoryId" v:"required|min:1#分类不能为空|分类不正确"`
	Name       string `json:"name" v:"required|length:2,128#商品名不能为空|商品名长度为2到128"`
	PriceCent  int64  `json:"priceCent" v:"min:1#价格必须大于0"`
	Stock      int    `json:"stock" v:"min:0#库存不能小于0"`
}

type CreateRes struct {
	ID int64 `json:"id"`
}

type DetailReq struct {
	g.Meta `path:"/products/{id}" method:"get" tags:"Product" summary:"商品详情"`
	ID     int64 `json:"id" in:"path" v:"min:1#商品ID不正确"`
}

type DetailRes struct {
	Product ProductItem `json:"product"`
}

type ListReq struct {
	g.Meta     `path:"/products" method:"get" tags:"Product" summary:"商品列表"`
	Page       int    `json:"page" in:"query" d:"1" v:"min:1#页码必须大于0"`
	Size       int    `json:"size" in:"query" d:"10" v:"between:1,100#每页数量必须在1到100之间"`
	Name       string `json:"name" in:"query"`
	CategoryID int64  `json:"categoryId" in:"query"`
	MinPrice   int64  `json:"minPrice" in:"query" v:"min:0#最低价格不能小于0"`
	MaxPrice   int64  `json:"maxPrice" in:"query" v:"min:0#最高价格不能小于0"`
}

type ListRes struct {
	List  []ProductItem `json:"list"`
	Total int           `json:"total"`
}

type UpdateReq struct {
	g.Meta     `path:"/admin/products/{id}" method:"put" tags:"Product" summary:"更新商品"`
	ID         int64  `json:"id" in:"path" v:"min:1#商品ID不正确"`
	CategoryID int64  `json:"categoryId" v:"min:1#分类不正确"`
	Name       string `json:"name" v:"required|length:2,128#商品名不能为空|商品名长度为2到128"`
	PriceCent  int64  `json:"priceCent" v:"min:1#价格必须大于0"`
	Stock      int    `json:"stock" v:"min:0#库存不能小于0"`
	Status     int    `json:"status" v:"in:0,1#状态只能是0或1"`
}

type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/admin/products/{id}" method:"delete" tags:"Product" summary:"删除商品"`
	ID     int64 `json:"id" in:"path" v:"min:1#商品ID不正确"`
}

type DeleteRes struct{}
