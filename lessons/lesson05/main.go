package main

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/util/gconv"
)

type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	TraceID string `json:"traceId"`
	CostMs  int64  `json:"costMs"`
}

type GetProductReq struct {
	g.Meta `path:"/products/:id" method:"get" tags:"Product" summary:"Get product"`

	ID int64 `json:"id" in:"path"`
}

type GetProductRes struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	PriceCent int64  `json:"priceCent"`
}

type ProductController struct{}

func (controller *ProductController) Get(
	ctx context.Context,
	req *GetProductReq,
) (res *GetProductRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	userID := r.GetParam("userId").Int64()

	if req.ID == 404 {
		return nil, gerror.NewCode(gcode.CodeNotFound, "商品不存在")
	}

	res = &GetProductRes{
		ID:        req.ID,
		Name:      "GoFrame Book for user " + gconv.String(userID),
		PriceCent: 9900,
	}
	return
}

func AccessLogMiddleware(r *ghttp.Request) {
	start := time.Now()

	r.Middleware.Next()

	duration := time.Since(start)
	g.Log().Infof(
		r.Context(),
		"%s %s cost=%s",
		r.Method,
		r.URL.Path,
		duration,
	)
}

func ResponseMiddleware(r *ghttp.Request) {
	start := time.Now()

	r.Middleware.Next()

	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}

	var (
		err                = r.GetError()
		res                = r.GetHandlerResponse()
		code    gcode.Code = gcode.CodeOK
		message            = code.Message()
	)

	if err != nil {
		code = gerror.Code(err)
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
		message = err.Error()
	}

	r.Response.WriteJson(ApiResponse{
		Code:    code.Code(),
		Message: message,
		Data:    res,
		TraceID: gtrace.GetTraceID(r.Context()),
		CostMs:  time.Since(start).Milliseconds(),
	})
}

func AuthMiddleware(r *ghttp.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, "未登录"))
		return
	}

	r.SetParam("userId", int64(12))
	r.Middleware.Next()
}

func main() {
	server := g.Server()
	server.SetPort(8004)

	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(
			ghttp.MiddlewareCORS,
			AccessLogMiddleware,
			ResponseMiddleware,
			AuthMiddleware,
		)
		group.Bind(&ProductController{})
	})

	server.Run()
}
