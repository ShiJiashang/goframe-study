// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"goframe-study/mall/internal/model"
)

type (
	IProduct interface {
		// List returns filtered and paginated products.
		List(ctx context.Context, in model.ProductListInput) (out *model.ProductListOutput, err error)
		// Create inserts a new product and returns its ID.
		Create(ctx context.Context, in model.ProductCreateInput) (out *model.ProductCreateOutput, err error)
		// Detail returns a single product by ID.
		Detail(ctx context.Context, in model.ProductDetailInput) (out *model.ProductDetailOutput, err error)
		// Update modifies an existing product by ID.
		Update(ctx context.Context, in model.ProductUpdateInput) (out *model.ProductUpdateOutput, err error)
		// Delete removes a product by ID.
		Delete(ctx context.Context, in model.ProductDeleteInput) (out *model.ProductDeleteOutput, err error)
	}
)

var (
	localProduct IProduct
)

func Product() IProduct {
	if localProduct == nil {
		panic("implement not found for interface IProduct, forgot register?")
	}
	return localProduct
}

func RegisterProduct(i IProduct) {
	localProduct = i
}
