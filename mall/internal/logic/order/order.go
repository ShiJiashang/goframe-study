package order

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"goframe-study/mall/internal/consts"
	"goframe-study/mall/internal/dao"
	"goframe-study/mall/internal/model"
	"goframe-study/mall/internal/model/do"
	"goframe-study/mall/internal/model/entity"
	"goframe-study/mall/internal/service"
)

func init() {
	service.RegisterOrder(New())
}

type sOrder struct{}

func New() *sOrder {
	return &sOrder{}
}

// Create creates an order and deducts stock in one transaction.
func (s *sOrder) Create(
	ctx context.Context,
	in model.OrderCreateInput,
) (out *model.OrderCreateOutput, err error) {
	if in.Quantity <= 0 {
		return nil, gerror.NewCode(
			gcode.CodeInvalidParameter,
			"购买数量必须大于0",
		)
	}

	orderNo := "M" + guid.S()
	var orderID int64
	var totalCent int64

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		productColumns := dao.Products.Columns()

		var product entity.Products
		queryErr := tx.Model(dao.Products.Table()).Ctx(ctx).
			Where(productColumns.Id, in.ProductID).
			LockUpdate().
			Scan(&product)
		if queryErr != nil {
			return gerror.WrapCode(
				gcode.CodeDbOperationError,
				queryErr,
				"查询商品失败",
			)
		}
		if product.Id == 0 {
			return gerror.NewCode(consts.CodeProductNotFound)
		}
		if int(product.Stock) < in.Quantity {
			return gerror.NewCode(consts.CodeOrderStockNotEnough)
		}

		stockResult, updateErr := tx.Model(dao.Products.Table()).Ctx(ctx).
			Where(productColumns.Id, in.ProductID).
			WhereGTE(productColumns.Stock, in.Quantity).
			Decrement(productColumns.Stock, in.Quantity)
		if updateErr != nil {
			return gerror.WrapCode(
				gcode.CodeDbOperationError,
				updateErr,
				"扣减库存失败",
			)
		}

		affected, affectedErr := stockResult.RowsAffected()
		if affectedErr != nil {
			return gerror.WrapCode(
				gcode.CodeDbOperationError,
				affectedErr,
				"读取扣库存结果失败",
			)
		}
		if affected != 1 {
			return gerror.NewCode(consts.CodeOrderStockNotEnough)
		}

		totalCent = int64(product.PriceCent) * int64(in.Quantity)

		orderID, queryErr = tx.Model(dao.Orders.Table()).Ctx(ctx).
			Data(do.Orders{
				OrderNo:   orderNo,
				UserId:    in.UserID,
				TotalCent: totalCent,
				Status:    1,
			}).
			InsertAndGetId()
		if queryErr != nil {
			return gerror.WrapCode(
				gcode.CodeDbOperationError,
				queryErr,
				"创建订单失败",
			)
		}

		_, queryErr = tx.Model(dao.OrderItems.Table()).Ctx(ctx).
			Data(do.OrderItems{
				OrderId:      orderID,
				ProductId:    product.Id,
				ProductName:  product.Name,
				PriceCent:    product.PriceCent,
				Quantity:     in.Quantity,
				SubtotalCent: totalCent,
			}).
			Insert()
		if queryErr != nil {
			return gerror.WrapCode(
				gcode.CodeDbOperationError,
				queryErr,
				"创建订单明细失败",
			)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &model.OrderCreateOutput{
		OrderID:   orderID,
		OrderNo:   orderNo,
		TotalCent: totalCent,
	}, nil
}
