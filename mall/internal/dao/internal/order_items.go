// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OrderItemsDao is the data access object for the table order_items.
type OrderItemsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  OrderItemsColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// OrderItemsColumns defines and stores column names for the table order_items.
type OrderItemsColumns struct {
	Id           string // è®¢å•æ˜Žç»†ID
	OrderId      string // è®¢å•ID
	ProductId    string // å•†å“ID
	ProductName  string // ä¸‹å•æ—¶å•†å“åç§°å¿«ç…§
	PriceCent    string // ä¸‹å•æ—¶å•ä»·ï¼Œå•ä½ä¸ºåˆ†
	Quantity     string // è´­ä¹°æ•°é‡
	SubtotalCent string // å°è®¡ï¼Œå•ä½ä¸ºåˆ†
	CreatedAt    string // åˆ›å»ºæ—¶é—´
}

// orderItemsColumns holds the columns for the table order_items.
var orderItemsColumns = OrderItemsColumns{
	Id:           "id",
	OrderId:      "order_id",
	ProductId:    "product_id",
	ProductName:  "product_name",
	PriceCent:    "price_cent",
	Quantity:     "quantity",
	SubtotalCent: "subtotal_cent",
	CreatedAt:    "created_at",
}

// NewOrderItemsDao creates and returns a new DAO object for table data access.
func NewOrderItemsDao(handlers ...gdb.ModelHandler) *OrderItemsDao {
	return &OrderItemsDao{
		group:    "default",
		table:    "order_items",
		columns:  orderItemsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *OrderItemsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *OrderItemsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *OrderItemsDao) Columns() OrderItemsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *OrderItemsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *OrderItemsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *OrderItemsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
