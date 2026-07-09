package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	server := g.Server()
	server.SetPort(8000)

	server.BindHandler("/hello", func(request *ghttp.Request) {
		request.Response.Write("Hello, GoFrame!")
	})
	server.BindHandler("/ping", func(request *ghttp.Request) {
		request.Response.WriteJson(map[string]string{"status": "ok"})
	})
	server.Run()
}
