package product

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	v1 "goframe-study/mall/api/product/v1"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/internal/service"
)

func (c *ControllerV1) Update(
	ctx context.Context,
	req *v1.UpdateReq,
) (res *v1.UpdateRes, err error) {
	operator := ghttp.RequestFromCtx(ctx).GetParam("currentUserId").Int64()
	g.Log().Infof(ctx, "商品更新 operator=%d productId=%d", operator, req.ID)

	_, err = service.Product().Update(ctx, model.ProductUpdateInput{
		ID:         req.ID,
		CategoryID: req.CategoryID,
		Name:       req.Name,
		PriceCent:  req.PriceCent,
		Stock:      req.Stock,
		Status:     req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
