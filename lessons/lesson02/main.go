package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	server := g.Server()
	server.SetPort(8001)

	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.GET("/products/:id", func(request *ghttp.Request) {
			productID := request.GetRouter("id").Int()
			currency := request.GetQuery("currency", "CNY").String()

			request.Response.WriteJson(g.Map{
				"id":       productID,
				"name":     "GoFrame Learning Keyboard",
				"currency": currency,
			})
		})

		group.POST("/products/preview", func(request *ghttp.Request) {
			name := request.Get("name").String()
			priceCent := request.Get("priceCent").Int64()

			request.Response.WriteJson(g.Map{
				"name":      name,
				"priceCent": priceCent,
			})
		})
		group.POST("/products/calculate", func(request *ghttp.Request) {
			priceCent := request.Get("priceCent").Int64()
			quantity := request.Get("quantity").Int()
			request.Response.WriteJson(g.Map{
				"priceCent": priceCent,
				"quantity":  quantity,
				"totalCent": int64(priceCent) * int64(quantity),
			})
		})
	})

	server.Run()
}
