package product

import (
	"context"

	v1 "goframe-study/mall/api/product/v1"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/internal/service"
)

func (c *ControllerV1) Delete(
	ctx context.Context,
	req *v1.DeleteReq,
) (res *v1.DeleteRes, err error) {
	_, err = service.Product().Delete(ctx, model.ProductDeleteInput{
		ID: req.ID,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
