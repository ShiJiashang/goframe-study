package main

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ListProductsReq struct {
	g.Meta `path:"/products" method:"get" tags:"Product" summary:"List products"`

	CategoryID int64 `json:"categoryId" in:"query" dc:"分类ID，不传表示全部分类"`
	Page       int   `json:"page" in:"query" d:"1" v:"min:1#页码必须大于0" dc:"页码"`
	Size       int   `json:"size" in:"query" d:"10" v:"between:1,100#每页数量必须在1到100之间" dc:"每页数量"`
}

type ProductItem struct {
	ID        int64  `json:"id" dc:"商品ID"`
	Name      string `json:"name" dc:"商品名称"`
	PriceCent int64  `json:"priceCent" dc:"商品价格，单位：分"`
	Stock     int    `json:"stock" dc:"库存数量"`
}

type ListProductsRes struct {
	List  []ProductItem `json:"list" dc:"商品列表"`
	Total int           `json:"total" dc:"总数量"`
	Page  int           `json:"page" dc:"当前页码"`
	Size  int           `json:"size" dc:"每页数量"`
}

type CreateProductReq struct {
	g.Meta `path:"/products" method:"post" tags:"Product" summary:"Create product"`

	Name       string `json:"name" v:"required|length:2,60#商品名不能为空|商品名长度必须在2到60个字符之间" dc:"商品名称"`
	CategoryID int64  `json:"categoryId" v:"min:1#分类ID必须大于0" dc:"分类ID"`
	PriceCent  int64  `json:"priceCent" v:"min:1#价格必须大于0" dc:"商品价格，单位：分"`
	Stock      int    `json:"stock" v:"min:0#库存不能小于0" dc:"库存数量"`
}

type CreateProductRes struct {
	ID int64 `json:"id" dc:"新商品ID"`
}

type ProductController struct{}

func (controller *ProductController) List(
	ctx context.Context,
	req *ListProductsReq,
) (res *ListProductsRes, err error) {
	res = &ListProductsRes{
		List: []ProductItem{
			{
				ID:        1001,
				Name:      "GoFrame Book",
				PriceCent: 9900,
				Stock:     20,
			},
		},
		Total: 1,
		Page:  req.Page,
		Size:  req.Size,
	}
	return
}

func (controller *ProductController) Create(
	ctx context.Context,
	req *CreateProductReq,
) (res *CreateProductRes, err error) {
	res = &CreateProductRes{
		ID: 1002,
	}
	return
}

func main() {
	server := g.Server()
	server.SetPort(8005)
	server.SetOpenApiPath("/api.json")
	server.SetSwaggerPath("/swagger")

	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(&ProductController{})
	})

	server.Run()
}
