package product

import (
	"context"

	v1 "goframe-study/mall/api/product/v1"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/internal/service"
)

func (c *ControllerV1) Create(
	ctx context.Context,
	req *v1.CreateReq,
) (res *v1.CreateRes, err error) {
	out, err := service.Product().Create(ctx, model.ProductCreateInput{
		CategoryID: req.CategoryID,
		Name:       req.Name,
		PriceCent:  req.PriceCent,
		Stock:      req.Stock,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{ID: out.ID}, nil
}
