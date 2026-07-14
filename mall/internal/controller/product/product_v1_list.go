package product

import (
	"context"
	v1 "goframe-study/mall/api/product/v1"

	"goframe-study/mall/internal/model"
	"goframe-study/mall/internal/service"
)

func (c *ControllerV1) List(
	ctx context.Context,
	req *v1.ListReq,
) (res *v1.ListRes, err error) {
	out, err := service.Product().List(ctx, model.ProductListInput{
		Page:       req.Page,
		Size:       req.Size,
		Name:       req.Name,
		CategoryID: req.CategoryID,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
	})
	if err != nil {
		return nil, err
	}

	list := make([]v1.ProductItem, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, v1.ProductItem{
			ID:         item.ID,
			CategoryID: item.CategoryID,
			Name:       item.Name,
			PriceCent:  item.PriceCent,
			Stock:      item.Stock,
			Status:     item.Status,
		})
	}

	return &v1.ListRes{
		List:  list,
		Total: out.Total,
	}, nil
}
