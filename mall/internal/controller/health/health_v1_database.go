package health

import (
	"context"

	v1 "goframe-study/mall/api/health/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

func (c *ControllerV1) Database(ctx context.Context, req *v1.DatabaseReq) (res *v1.DatabaseRes, err error) {
	start := gtime.Now()
	db := g.DB().Ctx(ctx)

	if err = db.PingMaster(); err != nil {
		return nil, gerror.WrapCode(
			gcode.CodeDbOperationError,
			err,
			"数据库连接失败",
		)
	}

	productModel := db.Model("products")
	productCount, err := productModel.Count()
	if err != nil {
		return nil, gerror.WrapCode(
			gcode.CodeDbOperationError,
			err,
			"统计商品数量失败",
		)
	}

	return &v1.DatabaseRes{
		Status:       "ok",
		ProductCount: productCount,
		Database:     "goframe_mail",
		CostMs:       int64(gtime.Now().Sub(start).Milliseconds()),
	}, nil
}
