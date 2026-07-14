package product

import (
	"context"

	v1 "goframe-study/mall/api/product/v1"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/internal/service"
)

func (c *ControllerV1) Detail(
	ctx context.Context,
	req *v1.DetailReq,
) (res *v1.DetailRes, err error) {
	out, err := service.Product().Detail(ctx, model.ProductDetailInput{
		ID: req.ID,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DetailRes{
		Product: v1.ProductItem{
			ID:         out.ID,
			CategoryID: out.CategoryID,
			Name:       out.Name,
			PriceCent:  out.PriceCent,
			Stock:      out.Stock,
			Status:     out.Status,
		},
	}, nil
}
