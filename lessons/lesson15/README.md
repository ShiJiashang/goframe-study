# 第 15 课：事务、库存与并发

## 本课目标

这一课实现“创建订单并扣减库存”，重点不是多写几个接口，而是保证：

- 扣库存、创建订单、创建订单明细要么全部成功，要么全部回滚。
- 多个请求同时购买时库存不会变成负数。
- 不会创建“有订单但没扣库存”或“扣了库存却没订单”的脏数据。

核心 API：

- `gdb.DB.Transaction`
- `gdb.TX`
- `tx.Model`
- `LockUpdate`
- `Decrement`
- `RowsAffected`

## 和 Gin 的对应关系

Gin 不管理数据库事务，通常直接使用 `database/sql`：

```go
tx, err := db.BeginTx(c.Request.Context(), nil)
if err != nil {
    return
}
defer tx.Rollback()

// 所有 SQL 都必须使用 tx

if err = tx.Commit(); err != nil {
    return
}
```

GoFrame 推荐使用闭包事务：

```go
err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
    // 所有数据库操作都使用 tx
    return nil
})
```

闭包返回 `nil` 自动提交，返回非 `nil error` 自动回滚。

## 为什么普通判断会超卖

错误写法：

```go
product := queryProduct()
if product.Stock >= quantity {
    updateStock(product.Stock - quantity)
}
```

库存只有 1 时，两个请求可能同时读取到 1：

```text
请求A读取 stock=1
请求B读取 stock=1
请求A判断通过
请求B判断通过
请求A更新
请求B更新
```

因此仅仅写 `if stock >= quantity` 不能处理并发。

## `Transaction`

`gdb.DB` 的方法签名：

```go
Transaction(
    ctx context.Context,
    f func(ctx context.Context, tx gdb.TX) error,
) error
```

参数：

- 外层 `ctx`：当前请求上下文。
- `f`：事务闭包。
- 闭包参数 `ctx`：事务内部应该继续使用的上下文。
- 闭包参数 `tx`：当前事务对象，类型为 `gdb.TX`。

返回：

- 闭包成功且提交成功：`nil`。
- 闭包返回错误：自动回滚并返回该错误。
- 提交或回滚失败：返回数据库错误。

不要在闭包里手动调用：

```go
tx.Commit()
tx.Rollback()
```

闭包事务会自动管理它们。

## `gdb.TX`

`gdb.TX` 是事务接口，常用方法：

```go
type TX interface {
    Ctx(ctx context.Context) TX
    Model(tableNameQueryOrStruct ...any) *gdb.Model
    Raw(sql string, args ...any) *gdb.Model
    Commit() error
    Rollback() error
}
```

本课使用闭包事务，所以只使用 `Model`，不手动提交和回滚。

事务中的错误写法：

```go
g.DB().Model("products") // 可能跑到事务外
```

正确写法：

```go
tx.Model("products")
```

只要有一步误用 `g.DB()`，这一步就可能无法随着事务回滚。

## `LockUpdate()`：行锁

```go
err := tx.Model("products").Ctx(ctx).
    Where("id", productID).
    LockUpdate().
    Scan(&product)
```

签名：

```go
func (m *Model) LockUpdate() *Model
```

它生成类似 SQL：

```sql
SELECT * FROM products WHERE id = ? FOR UPDATE;
```

作用：事务结束之前，其他事务不能同时修改这行商品。

注意：

- 必须放在事务中才有意义。
- 表应使用 InnoDB。
- 锁持有时间要短，不要在事务里调用支付等外部网络服务。

## `Decrement()` 与条件扣减

```go
result, err := tx.Model("products").Ctx(ctx).
    Where("id", productID).
    WhereGTE("stock", quantity).
    Decrement("stock", quantity)
```

签名：

```go
func (m *Model) Decrement(column string, amount any) (sql.Result, error)
```

产生类似 SQL：

```sql
UPDATE products
SET stock = stock - ?
WHERE id = ? AND stock >= ?;
```

`stock >= quantity` 是最后一道并发保护。

检查是否真正扣减：

```go
affected, err := result.RowsAffected()
if affected != 1 {
    return gerror.NewCode(consts.CodeOrderStockNotEnough)
}
```

## 业务输入输出模型

新增 `mall/internal/model/order.go`：

```go
package model

type OrderCreateInput struct {
    UserID    int64
    ProductID int64
    Quantity  int
}

type OrderCreateOutput struct {
    OrderID   int64
    OrderNo   string
    TotalCent int64
}
```

`Quantity` 即使在 API 层校验过，Logic 仍可以再次检查，因为 Logic 未来可能被定时任务或其他入口调用。

## 可运行样例：创建订单事务

新增 `mall/internal/logic/order/order.go`：

```go
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
```

变量作用：

- `orderNo`：事务外生成的业务订单号。
- `orderID`：订单插入后生成的主键，事务提交后返回。
- `totalCent`：商品单价乘数量，全程使用整数分。
- `tx`：同一个数据库事务，所有写操作必须使用它。
- `product`：加行锁后读取的商品快照。
- `stockResult`：扣库存 SQL 的 `sql.Result`。
- `affected`：实际修改行数；必须等于 1。
- `queryErr/updateErr/affectedErr`：分别保存当前步骤的错误，避免覆盖含义。

## 自动提交和回滚过程

成功：

```text
锁定商品
→ 扣库存
→ 插入 orders
→ 插入 order_items
→ 闭包返回 nil
→ 自动 COMMIT
```

任一步失败：

```text
扣库存成功
→ 插入订单失败
→ 闭包返回 error
→ 自动 ROLLBACK
→ 库存恢复到事务开始前
```

这和中间件的 `return` 不同：中间件 `return` 不能撤销已经写入的数据；数据库事务回滚可以撤销同一事务中的数据库修改。

## API 与 Controller

新增 `mall/api/order/v1/order.go`：

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

type CreateReq struct {
    g.Meta   `path:"/orders" method:"post" tags:"Order" summary:"创建订单"`
    UserID    int64 `json:"userId" v:"min:1#用户ID不正确"`
    ProductID int64 `json:"productId" v:"min:1#商品ID不正确"`
    Quantity  int   `json:"quantity" v:"between:1,100#购买数量必须在1到100之间"`
}

type CreateRes struct {
    OrderID   int64  `json:"orderId"`
    OrderNo   string `json:"orderNo"`
    TotalCent int64  `json:"totalCent"`
}
```

生成代码：

```bash
gf gen ctrl
gf gen service
```

如果 `RegisterOrder` 尚未生成，第一次可以先省略 `init()` 和 `service` 导入，执行 `gf gen service` 后再加回来；这和第 14 课第一次生成 `Product` Service 的顺序相同。

Controller 核心：

```go
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
```

还要像商品 Controller 一样，在 `cmd` 中绑定 `order.NewV1()`。

## 本课练习：并发下单验收

### 第一步：重置数据

```bash
docker exec goframe-mysql mysql -uroot -p12345678 -e '
USE goframe_mall;
DELETE FROM order_items;
DELETE FROM orders;
UPDATE products SET stock=10 WHERE id=1;
'
```

### 第二步：同时发起 20 个请求

```bash
seq 1 20 | xargs -P20 -I{} curl -s \
  -X POST 'http://127.0.0.1:8000/orders' \
  -H 'Content-Type: application/json' \
  -d '{"userId":1,"productId":1,"quantity":1}'
```

`-P20` 表示最多并行运行 20 个 curl。

### 第三步：检查数据库

```bash
docker exec goframe-mysql mysql -uroot -p12345678 -e '
USE goframe_mall;
SELECT id,stock FROM products WHERE id=1;
SELECT COUNT(*) AS order_count FROM orders;
SELECT COUNT(*) AS item_count FROM order_items;
'
```

预期：

```text
stock       = 0
order_count = 10
item_count  = 10
```

另外 10 个请求应该返回库存不足，而不是继续成功。

### 第四步：验证回滚

临时在“扣库存成功”和“插入订单”之间加入：

```go
return errors.New("模拟订单插入失败")
```

发起一次请求，确认：

- 接口返回失败。
- 库存没有减少。
- orders 和 order_items 没有新增。

验证后删除这行模拟错误。

## 常见错误

事务中误用普通 DB：

```go
g.DB().Model(...) // 错
tx.Model(...)     // 对
```

只读库存但不加锁或条件更新：

```go
if product.Stock > 0 {
    // 并发不安全
}
```

事务里调用外部支付：

```go
// 锁会持有很久，不要这样做
callPaymentService()
```

多个商品下单时还要统一加锁顺序，例如按商品 ID 从小到大锁定，降低死锁概率。

## 验收条件

- 所有事务内 SQL 都使用同一个 `tx`。
- 闭包返回错误时库存、订单、明细全部回滚。
- 库存 10、并发请求 20 次时只能成功 10 次。
- 最终库存为 0，永远不会小于 0。
- 成功订单数与订单明细数都是 10。
- 库存不足返回 `CodeOrderStockNotEnough`。
- 事务代码没有手动 `Commit()` 或 `Rollback()`。
- `go test ./...` 和 `go vet ./...` 通过。

完成后把并发请求统计、数据库查询结果和事务核心代码发给我。我先评审并指出竞态或事务边界问题，再给参考答案。
