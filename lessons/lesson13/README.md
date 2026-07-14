# 第 13 课：DAO 代码生成与模型区别

## 本课目标

这一课不追求写更多业务，而是把数据库表变成 GoFrame 推荐的数据访问层：

- 使用 `gf gen dao` 生成 DAO、DO、Entity。
- 看懂 `dao / dao/internal / model/do / model/entity`。
- 分清 API 模型、业务模型、DO 和 Entity。
- 明确哪些文件可以扩展，哪些 `DO NOT EDIT` 文件不能手改。

## 和 Gin 的对应关系

Gin 没有固定 DAO 生成方案，项目可能自己写：

```go
type ProductRepository interface {
    FindByID(ctx context.Context, id int64) (*Product, error)
}
```

GoFrame CLI 可以从真实数据库表读取字段和类型，生成：

```text
数据库表
├── dao          查询入口和字段名
├── do           写入和条件数据对象
└── entity       一整行数据库记录
```

生成代码减少手写字段名和结构体错误，但它不会替你完成业务逻辑。

## 生成前提

先完成第 11 课：

- MySQL 正常运行。
- `goframe_mall` 已导入五张表。
- MySQL 驱动已经安装。

修改 `mall/hack/config.yaml`：

```yaml
gfcli:
  gen:
    dao:
      - link: "mysql:root:12345678@tcp(127.0.0.1:3306)/goframe_mall"
        tables: "users,categories,products,orders,order_items"
        descriptionTag: true
```

这里是 CLI 配置，不是应用运行时的 `manifest/config/config.yaml`。

- `link`：代码生成时连接哪个数据库。
- `tables`：只生成哪些表，逗号分隔。
- `descriptionTag`：把数据库字段注释加入生成结果。

## 常用 `gf gen dao` 命令

查看帮助：

```bash
gf gen dao -h
```

按 `hack/config.yaml` 生成：

```bash
cd mall
gf gen dao
```

临时只生成指定表：

```bash
gf gen dao -t products,categories
```

临时通过命令行指定连接：

```bash
gf gen dao \
  -l 'mysql:root:12345678@tcp(127.0.0.1:3306)/goframe_mall' \
  -t products
```

常用选项：

```text
-t,  --tables       只生成指定表，多个表用逗号分隔
-x,  --tablesEx     排除指定表
-g,  --group        生成 DAO 使用的数据库配置分组，默认 default
-p,  --path         生成目录根路径，默认 internal
-a,  --clear        清理数据库中已经不存在的表对应的生成文件
-v,  --overwriteDao 覆盖 DAO 外层文件，使用前必须确认自定义代码不会丢失
```

日常优先把参数写入 `hack/config.yaml`，避免团队成员生成出不同结果。

## 生成后的目录

执行后大致得到：

```text
mall/internal
├── dao
│   ├── products.go
│   ├── categories.go
│   ├── orders.go
│   ├── order_items.go
│   ├── users.go
│   └── internal
│       ├── products.go
│       └── ...
└── model
    ├── do
    │   ├── products.go
    │   └── ...
    └── entity
        ├── products.go
        └── ...
```

实际文件名由表名和 CLI 的 `fileNameCase` 决定。

## DAO：数据库访问入口

外层 `internal/dao/products.go` 大致是：

```go
type productsDao struct {
    *internal.ProductsDao
}

var Products = productsDao{internal.NewProductsDao()}
```

变量：

- `Products`：全局可复用的商品 DAO 对象。
- `Products.Ctx(ctx)`：创建绑定上下文的商品 Model。
- `Products.Columns()`：获取生成的字段名集合。
- `Products.Table()`：获取真实表名。

使用：

```go
columns := dao.Products.Columns()

model := dao.Products.Ctx(ctx).
    Where(columns.Status, 1)
```

相比手写：

```go
Where("status", 1)
```

生成字段可以减少拼错字段名。

## `dao/internal`：DAO 基础实现

`internal/dao/internal/products.go` 通常包含：

```go
type ProductsDao struct {
    table   string
    group   string
    columns ProductsColumns
}

func (dao *ProductsDao) Ctx(ctx context.Context) *gdb.Model
func (dao *ProductsDao) Columns() ProductsColumns
func (dao *ProductsDao) Transaction(
    ctx context.Context,
    f func(context.Context, gdb.TX) error,
) error
```

这些文件带有：

```text
DO NOT EDIT
```

重新执行 `gf gen dao` 时会更新，不能手改。

## Entity：数据库一整行

`internal/model/entity/products.go` 大致是：

```go
type Products struct {
    Id         uint64      `json:"id" orm:"id"`
    CategoryId uint64      `json:"categoryId" orm:"category_id"`
    Name       string      `json:"name" orm:"name"`
    PriceCent  uint64      `json:"priceCent" orm:"price_cent"`
    Stock      uint        `json:"stock" orm:"stock"`
    Status     uint        `json:"status" orm:"status"`
    CreatedAt  *gtime.Time `json:"createdAt" orm:"created_at"`
    UpdatedAt  *gtime.Time `json:"updatedAt" orm:"updated_at"`
}
```

Entity 特点：

- 字段类型和数据库表一致。
- 表的一行是什么样，它就是什么样。
- 适合接收完整查询结果。
- 带 `DO NOT EDIT`，表结构变化后重新生成。

示例：

```go
var product entity.Products

err := dao.Products.Ctx(ctx).
    Where(dao.Products.Columns().Id, id).
    Scan(&product)
```

## DO：数据库操作对象

`internal/model/do/products.go` 大致是：

```go
type Products struct {
    g.Meta     `orm:"table:products, do:true"`
    Id         any
    CategoryId any
    Name       any
    PriceCent  any
    Stock      any
    Status     any
    CreatedAt  *gtime.Time
    UpdatedAt  *gtime.Time
}
```

DO 字段很多是 `any`，因为它用于构造部分更新和条件：

```go
data := do.Products{
    Name:      "新名称",
    PriceCent: 6990,
}

_, err := dao.Products.Ctx(ctx).
    Where(dao.Products.Columns().Id, 1).
    Data(data).
    Update()
```

没有设置的 DO 字段保持 `nil`，ORM 可以忽略它们；而 `Stock: 0` 是明确要求把库存更新为 0。

DO 同样是生成文件，不手改。

## 四种模型的区别

| 类型 | 所在目录 | 服务对象 | 主要作用 |
| --- | --- | --- | --- |
| API 模型 | `api/product/v1` | HTTP 客户端 | path/query/body、校验、JSON 文档 |
| 业务模型 | `internal/model` | Logic/Service | 表达用例输入输出，不依赖 HTTP |
| DO | `internal/model/do` | ORM 写入/条件 | INSERT、UPDATE、WHERE 数据 |
| Entity | `internal/model/entity` | 数据库完整记录 | 接收一行完整查询结果 |

不要直接把 Entity 当 API 响应长期使用。否则数据库新增一个敏感字段时，可能意外返回给客户端。

## 可运行样例：使用生成 DAO 查询商品

把第 12 课商品详情中的：

```go
g.DB().Model("products").Ctx(ctx)
```

改成：

```go
var product entity.Products
columns := dao.Products.Columns()

err := dao.Products.Ctx(ctx).
    Where(columns.Id, req.ID).
    Scan(&product)
if err != nil {
    return nil, gerror.WrapCode(
        gcode.CodeDbOperationError,
        err,
        "查询商品失败",
    )
}
if product.Id == 0 {
    return nil, gerror.NewCode(consts.CodeProductNotFound)
}

return &v1.DetailRes{
    Product: v1.ProductItem{
        ID:         int64(product.Id),
        CategoryID: int64(product.CategoryId),
        Name:       product.Name,
        PriceCent:  int64(product.PriceCent),
        Stock:      int(product.Stock),
        Status:     int(product.Status),
    },
}, nil
```

这里故意显式转换 Entity 到 API 模型，让数据库结构和对外结构保持边界。

## 什么时候重新生成

数据库表结构变化后执行：

```bash
gf gen dao
```

例如：

- 新增字段。
- 修改字段类型。
- 新增或删除表。
- 修改数据库字段注释。

生成前后都要查看 Git diff，确认没有误连其他数据库或覆盖自定义代码。

## 本课练习

1. 修改 `mall/hack/config.yaml`，连接 `goframe_mall`。
2. 为五张表执行 `gf gen dao`。
3. 找到 `Products.Columns().PriceCent` 对应的实际字符串。
4. 把第 12 课的商品详情和商品更新改成使用 `dao.Products`。
5. 商品更新的数据改用 `do.Products`。
6. 查询结果先接收到 `entity.Products`，再转换成 API 的 `ProductItem`。

更新示意：

```go
columns := dao.Products.Columns()

_, err := dao.Products.Ctx(ctx).
    Where(columns.Id, req.ID).
    Data(do.Products{
        Name:      req.Name,
        PriceCent: req.PriceCent,
        Stock:     req.Stock,
        Status:    req.Status,
    }).
    Update()
```

## 验收条件

- 五张表都生成 `dao`、`do`、`entity` 文件。
- 没有手改任何标有 `DO NOT EDIT` 的文件。
- 商品详情不再直接写 `Model("products")`。
- 更新使用 `do.Products`，条件使用 `Products.Columns()`。
- API 响应由 Entity 显式转换而来，没有直接返回 Entity。
- 再执行一次 `gf gen dao` 后项目仍能编译。
- `go test ./...` 和 `go vet ./...` 通过。

完成后把生成目录树、详情查询和更新代码发给我，我先评审再给参考答案。
