package main

import (
	"context"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gvalid"
)

type CreateProductReq struct {
	g.Meta `path:"/products" method:"post" tags:"Product" summary:"Create product with validation"`

	Name       string `json:"name" v:"required|length:2,60#商品名不能为空|商品名长度必须在2到60个字符之间"`
	PriceCent  int64  `json:"priceCent" v:"min:1#价格必须大于0"`
	Stock      int    `json:"stock" v:"min:0#库存不能小于0"`
	CategoryID int64  `json:"categoryId" v:"min:1#分类ID必须大于0"`
}

type CreateProductRes struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	PriceCent  int64  `json:"priceCent"`
	Stock      int    `json:"stock"`
	CategoryID int64  `json:"categoryId"`
}

type CheckPriceReq struct {
	g.Meta `path:"/tools/check-price" method:"get" tags:"Tool" summary:"Check price with gvalid"`

	Price string `json:"price" in:"query"`
}

type CheckPriceRes struct {
	PriceCent int64 `json:"priceCent"`
}

type DynamicValueReq struct {
	g.Meta `path:"/tools/dynamic" method:"post" tags:"Tool" summary:"Inspect dynamic value"`

	Value any `json:"value"`
}

type DynamicValueRes struct {
	StringValue string `json:"stringValue"`
	IntValue    int    `json:"intValue"`
	BoolValue   bool   `json:"boolValue"`
}

type ProductController struct{}

func (controller *ProductController) Create(
	ctx context.Context,
	req *CreateProductReq,
) (res *CreateProductRes, err error) {
	res = &CreateProductRes{
		ID:         1001,
		Name:       req.Name,
		PriceCent:  req.PriceCent,
		Stock:      req.Stock,
		CategoryID: req.CategoryID,
	}
	return
}

type ToolController struct{}

func (controller *ToolController) CheckPrice(
	ctx context.Context,
	req *CheckPriceReq,
) (res *CheckPriceRes, err error) {
	priceCent := gconv.Int64(req.Price)

	if err = gvalid.New().
		Data(priceCent).
		Rules("min:1").
		Messages("价格必须大于0").
		Run(ctx); err != nil {
		return nil, err
	}

	res = &CheckPriceRes{
		PriceCent: priceCent,
	}
	return
}

func (controller *ToolController) Dynamic(
	ctx context.Context,
	req *DynamicValueReq,
) (res *DynamicValueRes, err error) {
	value := gvar.New(req.Value)

	res = &DynamicValueRes{
		StringValue: value.String(),
		IntValue:    value.Int(),
		BoolValue:   value.Bool(),
	}
	return
}

func main() {
	server := g.Server()
	server.SetPort(8003)

	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(
			&ProductController{},
			&ToolController{},
		)
	})

	server.Run()
}
