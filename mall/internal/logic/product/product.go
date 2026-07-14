package product

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"goframe-study/mall/internal/consts"
	"goframe-study/mall/internal/dao"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/internal/model/do"
	"goframe-study/mall/internal/model/entity"
	"goframe-study/mall/internal/service"
)

const productDetailCacheTTL = 600

func productDetailCacheKey(id int64) string {
	return fmt.Sprintf("mall:product:detail:%d", id)
}

func init() {
	service.RegisterProduct(New())
}

type sProduct struct{}

func New() *sProduct {
	return &sProduct{}
}

// List returns filtered and paginated products.
func (s *sProduct) List(
	ctx context.Context,
	in model.ProductListInput,
) (out *model.ProductListOutput, err error) {
	columns := dao.Products.Columns()
	query := dao.Products.Ctx(ctx)

	if in.Name != "" {
		query = query.WhereLike(columns.Name, "%"+in.Name+"%")
	}
	if in.CategoryID > 0 {
		query = query.Where(columns.CategoryId, in.CategoryID)
	}
	if in.MinPrice > 0 {
		query = query.WhereGTE(columns.PriceCent, in.MinPrice)
	}
	if in.MaxPrice > 0 {
		query = query.WhereLTE(columns.PriceCent, in.MaxPrice)
	}

	total, err := query.Count()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "统计商品失败")
	}

	var rows []entity.Products
	err = query.
		OrderDesc(columns.Id).
		Page(in.Page, in.Size).
		Scan(&rows)
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "查询商品失败")
	}

	list := make([]model.ProductListItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, model.ProductListItem{
			ID:         int64(row.Id),
			CategoryID: int64(row.CategoryId),
			Name:       row.Name,
			PriceCent:  int64(row.PriceCent),
			Stock:      int(row.Stock),
			Status:     int(row.Status),
		})
	}

	return &model.ProductListOutput{
		List:  list,
		Total: total,
	}, nil
}

// Create inserts a new product and returns its ID.
func (s *sProduct) Create(
	ctx context.Context,
	in model.ProductCreateInput,
) (out *model.ProductCreateOutput, err error) {
	result, err := dao.Products.Ctx(ctx).
		Data(do.Products{
			CategoryId: in.CategoryID,
			Name:       in.Name,
			PriceCent:  in.PriceCent,
			Stock:      in.Stock,
			Status:     1,
		}).
		Insert()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "新增商品失败")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "获取商品ID失败")
	}
	return &model.ProductCreateOutput{ID: id}, nil
}

// Detail returns a single product by ID, with Redis cache-aside.
func (s *sProduct) Detail(
	ctx context.Context,
	in model.ProductDetailInput,
) (out *model.ProductDetailOutput, err error) {
	key := productDetailCacheKey(in.ID)

	// 1) 先查缓存
	cached, cacheErr := g.Redis().Get(ctx, key)
	if cacheErr != nil {
		g.Log().Warningf(ctx, "读取商品缓存失败 key=%s err=%v", key, cacheErr)
	} else if cached != nil && !cached.IsNil() && cached.String() != "" {
		out = new(model.ProductDetailOutput)
		if decodeErr := gjson.DecodeTo(cached.Bytes(), out); decodeErr == nil {
			return out, nil
		}
		// 缓存内容损坏时删除，继续查数据库
		_, _ = g.Redis().Del(ctx, key)
	}

	// 2) 缓存未命中 → 查数据库
	var product *entity.Products
	columns := dao.Products.Columns()
	err = dao.Products.Ctx(ctx).
		Where(columns.Id, in.ID).
		Scan(&product)
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "查询商品失败")
	}
	if product == nil {
		return nil, gerror.NewCode(consts.CodeProductNotFound)
	}

	out = &model.ProductDetailOutput{
		ID:         int64(product.Id),
		CategoryID: int64(product.CategoryId),
		Name:       product.Name,
		PriceCent:  int64(product.PriceCent),
		Stock:      int(product.Stock),
		Status:     int(product.Status),
	}

	// 3) 回填缓存（失败仅告警，不影响返回）
	encoded, encodeErr := gjson.Encode(out)
	if encodeErr != nil {
		g.Log().Warningf(ctx, "编码商品缓存失败 key=%s err=%v", key, encodeErr)
		return out, nil
	}
	if setErr := g.Redis().SetEX(ctx, key, encoded, productDetailCacheTTL); setErr != nil {
		g.Log().Warningf(ctx, "写入商品缓存失败 key=%s err=%v", key, setErr)
	}

	return out, nil
}

// Update modifies an existing product by ID and invalidates its cache.
func (s *sProduct) Update(
	ctx context.Context,
	in model.ProductUpdateInput,
) (out *model.ProductUpdateOutput, err error) {
	columns := dao.Products.Columns()
	result, err := dao.Products.Ctx(ctx).
		Where(columns.Id, in.ID).
		Data(do.Products{
			CategoryId: in.CategoryID,
			Name:       in.Name,
			PriceCent:  in.PriceCent,
			Stock:      in.Stock,
			Status:     in.Status,
		}).
		Update()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "更新商品失败")
	}

	if _, err = result.RowsAffected(); err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "读取更新结果失败")
	}

	// 数据库更新成功 → 删除缓存
	key := productDetailCacheKey(in.ID)
	if _, delErr := g.Redis().Del(ctx, key); delErr != nil {
		return nil, gerror.WrapCode(
			gcode.CodeOperationFailed,
			delErr,
			"商品已更新但缓存失效失败",
		)
	}
	return &model.ProductUpdateOutput{}, nil
}

// Delete removes a product by ID.
func (s *sProduct) Delete(
	ctx context.Context,
	in model.ProductDeleteInput,
) (out *model.ProductDeleteOutput, err error) {
	columns := dao.Products.Columns()
	result, err := dao.Products.Ctx(ctx).
		Where(columns.Id, in.ID).
		Delete()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "删除商品失败")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "读取删除结果失败")
	}
	if affected == 0 {
		return nil, gerror.NewCode(consts.CodeProductNotFound)
	}
	return &model.ProductDeleteOutput{}, nil
}
