// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductsDao is the data access object for the table products.
type ProductsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ProductsColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ProductsColumns defines and stores column names for the table products.
type ProductsColumns struct {
	Id         string // å•†å“ID
	CategoryId string // åˆ†ç±»ID
	Name       string // å•†å“åç§°
	PriceCent  string // ä»·æ ¼ï¼Œå•ä½ä¸ºåˆ†
	Stock      string // åº“å­˜
	Status     string // çŠ¶æ€ï¼š1ä¸Šæž¶ï¼Œ0ä¸‹æž¶
	CreatedAt  string // åˆ›å»ºæ—¶é—´
	UpdatedAt  string // æ›´æ–°æ—¶é—´
}

// productsColumns holds the columns for the table products.
var productsColumns = ProductsColumns{
	Id:         "id",
	CategoryId: "category_id",
	Name:       "name",
	PriceCent:  "price_cent",
	Stock:      "stock",
	Status:     "status",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
}

// NewProductsDao creates and returns a new DAO object for table data access.
func NewProductsDao(handlers ...gdb.ModelHandler) *ProductsDao {
	return &ProductsDao{
		group:    "default",
		table:    "products",
		columns:  productsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductsDao) Columns() ProductsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
