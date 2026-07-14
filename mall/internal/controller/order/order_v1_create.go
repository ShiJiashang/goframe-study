package order

import (
	"context"

	v1 "goframe-study/mall/api/order/v1"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/internal/service"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	out, err := service.Order().Create(ctx, model.OrderCreateInput{
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{
		OrderID:   out.OrderID,
		OrderNo:   out.OrderNo,
		TotalCent: out.TotalCent,
	}, nil
}
