package debug

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"

	"goframe-study/mall/api/debug/v1"
	"goframe-study/mall/internal/consts"
)

func (c *ControllerV1) Error(ctx context.Context, req *v1.ErrorReq) (res *v1.ErrorRes, err error) {
	traceID := gtrace.GetTraceID(ctx)
	g.Log().Infof(ctx, "debug error request type=%s traceId=%s", req.Type, traceID)

	switch req.Type {
	case "notfound":
		err = gerror.NewCode(consts.CodeProductNotFound)
		g.Log().Warning(ctx, err)
		return nil, err

	case "wrap":
		baseErr := errors.New("database query returned empty result")
		err = gerror.WrapCode(consts.CodeProductNotFound, baseErr, "查询商品失败")
		g.Log().Error(ctx, err)
		return nil, err

	case "panic":
		panic("模拟 panic：生产代码不要这样处理业务错误")
	}

	res = &v1.ErrorRes{
		TraceID: traceID,
		Message: "ok",
	}
	return
}
