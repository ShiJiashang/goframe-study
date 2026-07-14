package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "goframe-study/mall/api/config/v1"
)

func (c *ControllerV1) App(ctx context.Context, req *v1.AppReq) (res *v1.AppRes, err error) {
	cfg := g.Cfg()

	effectiveEnv, err := cfg.GetEffective(ctx, "app.env", "dev")
	if err != nil {
		return nil, err
	}
	productDefaultSort := cfg.MustGet(ctx, "product.defaultSort", "createdAtDesc").String()
	res = &v1.AppRes{
		Name:         cfg.MustGet(ctx, "app.name", "GoFrame Mall").String(),
		Env:          cfg.MustGet(ctx, "app.env", "dev").String(),
		EffectiveEnv: effectiveEnv.String(),
		Debug:        cfg.MustGet(ctx, "app.debug", false).Bool(),
		Address:      cfg.MustGet(ctx, "server.address", ":8000").String(),
		PageSize:     cfg.MustGet(ctx, "product.pageSize", 10).Int(),
		DefaultSort:  productDefaultSort,
	}
	return
}
