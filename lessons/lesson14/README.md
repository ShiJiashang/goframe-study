# 第 14 课：Controller、Service、Logic 协作

## 本课目标

第 12 课把 SQL 直接写在 Controller 中，是为了先认识 ORM。本课把职责拆开：

```text
HTTP 请求
   ↓
Controller：绑定参数、调用 Service、转换响应
   ↓
Service：定义业务能力接口
   ↓
Logic：实现业务规则
   ↓
DAO：访问数据库
```

目标是把商品 CRUD 从 Controller 移到 Logic。

## 和 Gin 的对应关系

Gin 项目也常见：

```text
handler → service → repository
```

GoFrame 推荐结构对应为：

```text
controller → service → logic → dao
```

区别是 GoFrame 的 `gf gen service` 可以从 Logic 方法生成 Service 接口和注册函数。

## 每层具体负责什么

### API

目录：

```text
api/product/v1
```

负责：

- 路由 `g.Meta`。
- path/query/body 参数。
- `v` 校验标签。
- OpenAPI 字段描述。
- 对外 JSON 结构。

### Controller

目录：

```text
internal/controller/product
```

负责：

- 接收 `ctx` 和 `req`。
- 把 API Req 转成业务 Input。
- 调用 `service.Product()`。
- 把业务 Output 转成 API Res。

不负责：

- 拼 SQL。
- 判断库存。
- 开事务。
- 实现复杂业务流程。

### Service

目录：

```text
internal/service
```

这里保存的是业务接口和实现注册表，不是具体业务代码。

Controller 只依赖接口：

```go
service.Product().List(ctx, input)
```

### Logic

目录：

```text
internal/logic/product
```

负责：

- 实现 Service 接口。
- 组织业务规则。
- 调用 DAO。
- 包装业务错误。
- 管理事务。

### Model

目录：

```text
internal/model
```

保存 Logic 的输入输出模型。它们不携带 HTTP 路由和参数来源。

### DAO

负责数据库访问，不负责“为什么要这样做”的业务决策。

## 依赖方向

正确方向：

```text
api ← controller → service ← logic → dao
                         ↘ model ↙
```

Controller 不应该直接导入：

```go
goframe-study/mall/internal/logic/product
```

Controller 应该导入：

```go
goframe-study/mall/internal/service
```

这样测试时可以注册另一个 Service 实现。

## 第一步：定义业务输入输出模型

新增 `mall/internal/model/product.go`：

```go
package model

type ProductListInput struct {
    Page       int
    Size       int
    Name       string
    CategoryID int64
    MinPrice   int64
    MaxPrice   int64
}

type ProductListItem struct {
    ID         int64
    CategoryID int64
    Name       string
    PriceCent  int64
    Stock      int
    Status     int
}

type ProductListOutput struct {
    List  []ProductListItem
    Total int
}
```

这些类型不写：

```go
g.Meta
in:"query"
json:"..."
v:"..."
```

原因是业务层不应该知道输入来自 HTTP query、消息队列还是定时任务。

## 第二步：先写 Logic 实现

第一次生成 Service 时，先创建 `mall/internal/logic/product/product.go`，暂时不要写注册代码：

```go
package product

import (
    "context"

    "github.com/gogf/gf/v2/errors/gcode"
    "github.com/gogf/gf/v2/errors/gerror"

    "goframe-study/mall/internal/dao"
    "goframe-study/mall/internal/model"
    "goframe-study/mall/internal/model/entity"
)

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
```

关键类型和变量：

- `sProduct`：商品 Service 的实际实现类型，`s` 是 GoFrame 常用命名约定。
- `New() *sProduct`：创建实现对象。
- `in`：与 HTTP 无关的业务输入。
- `out`：业务输出。
- `columns`：DAO 生成的字段名集合。
- `query`：绑定 `ctx` 的商品 ORM Model。
- `rows`：数据库 Entity 切片。
- `list`：转换后的业务商品切片。

## 第三步：生成 Service

进入 `mall` 执行：

```bash
gf gen service
```

查看命令说明：

```bash
gf gen service -h
```

常用形式：

```bash
gf gen service
gf gen service -p product
gf gen service -s internal/logic -d internal/service
```

- `-s`：扫描 Logic 的目录，默认 `internal/logic`。
- `-d`：生成 Service 的目录，默认 `internal/service`。
- `-p`：只生成指定 Logic 包。

CLI 默认识别名称符合 `sXxx` 的结构体，例如：

```go
type sProduct struct{}
```

它会根据 `sProduct` 的导出方法生成 `internal/service/product.go`，还会生成 `internal/logic/logic.go` 用于统一空白导入各 Logic 包。

## 生成的 Service 内容

`internal/service/product.go` 大致是：

```go
type IProduct interface {
    List(
        ctx context.Context,
        in model.ProductListInput,
    ) (out *model.ProductListOutput, err error)
}

var localProduct IProduct

func Product() IProduct {
    if localProduct == nil {
        panic("implement not found for interface IProduct, forgot register?")
    }
    return localProduct
}

func RegisterProduct(i IProduct) {
    localProduct = i
}
```

逐项解释：

- `IProduct`：商品业务能力接口。
- `localProduct`：当前注册的接口实现，类型是 `IProduct`。
- `Product()`：让 Controller 取得当前实现。
- `RegisterProduct(i)`：把 Logic 实现注册进来。
- 参数 `i IProduct`：任何实现了全部接口方法的类型都能注册。

这个生成文件带有 `DO NOT EDIT`，Logic 方法变化后重新运行 `gf gen service`。

## 第四步：注册 Logic 实现

Service 文件生成后，在 `internal/logic/product/product.go` 中加入：

```go
import "goframe-study/mall/internal/service"

func init() {
    service.RegisterProduct(New())
}
```

为什么第一次要后加？因为第一次生成前 `internal/service/product.go` 还不存在，Logic 无法导入它。

还要在 `mall/internal/cmd/cmd.go` 中加入空白导入：

```go
import (
    _ "goframe-study/mall/internal/logic"
)
```

执行过程：

```text
cmd 导入 internal/logic
    ↓
logic.go 空白导入 logic/product
    ↓
product 包的 init() 执行
    ↓
service.RegisterProduct(New())
```

漏掉注册或空白导入时，调用 `service.Product()` 会 panic：

```text
implement not found for interface IProduct, forgot register?
```

## 第五步：Controller 只调用 Service

重写商品列表 Controller：

```go
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
```

Controller 中已经没有：

```go
g.DB()
dao.Products
Where(...)
Scan(...)
```

这就是本课重构的判断标准。

## 为什么不把 API Req 直接传给 Logic

不要这样：

```go
service.Product().List(ctx, req)
```

否则 Logic 会依赖 `api/product/v1`，业务层就和 HTTP 版本绑定。

正确做法：

```go
API ListReq
    ↓ Controller转换
model.ProductListInput
    ↓ Logic处理
model.ProductListOutput
    ↓ Controller转换
API ListRes
```

虽然转换代码多一点，但边界更稳定，也更容易测试。

## 本课练习：重构其余商品 CRUD

按 List 的模式处理 `Create`、`Detail`、`Update`、`Delete`。

照着做：

1. 在 `internal/model/product.go` 定义每个用例的 Input/Output。
2. 在 `sProduct` 上增加对应导出方法。
3. 把原 Controller 中的 DAO/ORM 代码移动到 Logic。
4. 执行 `gf gen service` 更新 `IProduct`。
5. Controller 只负责 Req/Input 和 Output/Res 转换。
6. 不要修改生成的 `internal/service/product.go`。

例如创建用例模型可以从这里开始：

```go
type ProductCreateInput struct {
    CategoryID int64
    Name       string
    PriceCent  int64
    Stock      int
}

type ProductCreateOutput struct {
    ID int64
}
```

Logic 方法外形：

```go
func (s *sProduct) Create(
    ctx context.Context,
    in model.ProductCreateInput,
) (out *model.ProductCreateOutput, err error) {
    // 把第12课的新增数据库代码移动到这里。
    return
}
```

## 验收条件

- `gf gen service` 生成 `IProduct`、`Product()`、`RegisterProduct()`。
- Logic 在 `init()` 中完成注册。
- `cmd` 空白导入生成的 `internal/logic` 聚合包。
- Controller 不导入 `gdb`、`dao`、`do`、`entity`。
- Logic 不导入任何 `api/.../v1` 包。
- 商品 CRUD 接口行为和重构前一致。
- 修改 Logic 方法后重新生成 Service，项目仍能编译。
- `go test ./...` 和 `go vet ./...` 通过。

完成后把目录树、生成的 Service 接口、一个 Controller 和对应 Logic 发给我，我先评审再给参考答案。
