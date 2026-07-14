package main

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type CreateProductReq struct {
	g.Meta    `path:"/products" method:"post" tags:"Product" summary:"Create product"`
	Name      string `json:"name"`
	PriceCent int64  `json:"priceCent"`
}

type CreateProductRes struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	PriceCent int64  `json:"priceCent"`
}

type GetProductReq struct {
	g.Meta `path:"/products/:id" method:"get" tags:"Product" summary:"Get product"`

	ID            int64  `json:"id" in:"path"`
	Page          int    `json:"page" in:"query"`
	Authorization string `json:"Authorization" in:"header"`
}

type GetProductRes struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Page          int    `json:"page"`
	Authorization string `json:"authorization"`
}

type ProductController struct{}

func (controller *ProductController) Create(
	ctx context.Context,
	req *CreateProductReq,
) (res *CreateProductRes, err error) {
	res = &CreateProductRes{
		ID:        1001,
		Name:      req.Name,
		PriceCent: req.PriceCent,
	}
	return
}

func (controller *ProductController) Get(
	ctx context.Context,
	req *GetProductReq,
) (res *GetProductRes, err error) {
	res = &GetProductRes{
		ID:            req.ID,
		Name:          "GoFrame Book",
		Page:          req.Page,
		Authorization: req.Authorization,
	}
	return
}

func main() {
	server := g.Server()
	server.SetPort(8002)

	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(&ProductController{})
	})

	server.Run()
}
