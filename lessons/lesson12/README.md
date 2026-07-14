# 第 12 课：ORM 增删改查、筛选与分页

## 本课目标

这一课直接操作 `products` 表，完成商品 CRUD，并理解：

- `Data`：设置要写入的数据。
- `Where`：设置查询或更新条件。
- `Fields`：指定查询字段。
- `Insert`、`Update`、`Delete`：增改删。
- `Scan`、`One`、`All`：读取查询结果。
- `Count`、`Limit`、`Offset`：统计与分页。

第 13 课才使用自动生成的 DAO，本课故意先用：

```go
g.DB().Model("products")
```

这样你能看懂 DAO 最终封装了什么。

## 和 Gin 的对应关系

Gin 不负责数据库操作。Gin 配合 `database/sql` 时可能写：

```go
rows, err := db.QueryContext(
    c.Request.Context(),
    "SELECT id,name,price_cent FROM products WHERE status=? LIMIT ? OFFSET ?",
    1,
    size,
    offset,
)
```

GoFrame ORM 写成链式调用：

```go
err := g.DB().Model("products").Ctx(ctx).
    Fields("id", "name", "price_cent").
    Where("status", 1).
    Limit(size).
    Offset(offset).
    Scan(&list)
```

Gin 的 `c.Request.Context()` 对应当前 Controller 的 `ctx`。

## 核心类型

### `*gdb.Model`

```go
model := g.DB().Model("products").Ctx(ctx)
```

`model` 的类型是：

```go
*gdb.Model
```

它保存的是“准备构造的 SQL”，不是一行商品，也不是数据库连接本身。

链式方法一般返回新的 `*gdb.Model`，因此可以继续拼接：

```go
model = model.Where("status", 1)
model = model.WhereLike("name", "%GoFrame%")
```

### `g.Map`

```go
data := g.Map{
    "name":       "GoFrame 实战手册",
    "price_cent": 5990,
}
```

它可以理解为：

```go
map[string]any
```

键是数据库字段名，值是准备写入的内容。

### `gdb.Record` 和 `gdb.Result`

动态查询结果类型：

```go
type Record map[string]*gvar.Var
type Result []Record
```

- `Record`：一行数据。
- `Result`：多行数据。
- 每个字段值是 `*gvar.Var`。

业务代码更推荐使用 `Scan` 转成明确的结构体。

## API 方法与参数

### `Data`

```go
func (m *Model) Data(data ...any) *Model
```

作用：设置 INSERT 或 UPDATE 的数据。

```go
model.Data(g.Map{"name": "键盘", "stock": 20})
```

它只负责保存数据，后面调用 `Insert()` 或 `Update()` 才真正执行 SQL。

### `Where`

```go
func (m *Model) Where(where any, args ...any) *Model
```

常用写法：

```go
model.Where("id", 1)
model.Where("status", 1)
model.Where("price_cent >= ?", 1000)
```

- `where`：字段名、条件字符串、map 或结构体。
- `args`：占位符对应的值。

不要把用户输入直接拼接到 SQL 字符串，应使用参数绑定。

### `Fields`

```go
func (m *Model) Fields(fieldNamesOrMapStruct ...any) *Model
```

```go
model.Fields("id", "name", "price_cent", "stock")
```

只查询接口真正需要的字段，避免无意义的 `SELECT *`。

### `Insert`

```go
func (m *Model) Insert(data ...any) (sql.Result, error)
```

两种等价写法：

```go
result, err := model.Data(data).Insert()
result, err := model.Insert(data)
```

`sql.Result` 常用方法：

```go
LastInsertId() (int64, error)
RowsAffected() (int64, error)
```

### `Update`

```go
func (m *Model) Update(dataAndWhere ...any) (sql.Result, error)
```

推荐把条件和数据分开写清楚：

```go
result, err := model.
    Where("id", id).
    Data(data).
    Update()
```

没有 `Where` 的 Update 可能修改整张表，必须特别小心。

### `Delete`

```go
func (m *Model) Delete(where ...any) (sql.Result, error)
```

```go
result, err := model.Where("id", id).Delete()
```

本课是硬删除。生产商城通常会增加 `deleted_at` 做软删除，后面再扩展。

### `One`、`All`、`Scan`

```go
func (m *Model) One(where ...any) (gdb.Record, error)
func (m *Model) All(where ...any) (gdb.Result, error)
func (m *Model) Scan(pointer any, where ...any) error
```

动态读取：

```go
record, err := model.Where("id", 1).One()
name := record["name"].String()

records, err := model.Where("status", 1).All()
```

结构化读取：

```go
var product Product
err := model.Where("id", 1).Scan(&product)

var products []Product
err := model.Where("status", 1).Scan(&products)
```

`Scan` 的参数必须是指针，否则 ORM 无法把结果写进去。

### `Count`、`Limit`、`Offset`

```go
func (m *Model) Count(where ...any) (int, error)
func (m *Model) Limit(limit ...int) *Model
func (m *Model) Offset(offset int) *Model
```

分页：

```go
offset := (page - 1) * size

total, err := model.Count()
err = model.Limit(size).Offset(offset).Scan(&list)
```

- `Count()`：返回过滤条件下的总数，不是当前页数量。
- `Limit(size)`：一页最多读取多少条。
- `Offset(offset)`：跳过多少条。

## 可运行样例：商品 CRUD

在 `mall/api/product/v1/product.go` 中定义这些接口。原来的 `ListReq/ListRes` 可以按下面结构调整：

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

type ProductItem struct {
    ID         int64  `json:"id"`
    CategoryID int64  `json:"categoryId"`
    Name       string `json:"name"`
    PriceCent  int64  `json:"priceCent"`
    Stock      int    `json:"stock"`
    Status     int    `json:"status"`
}

type CreateReq struct {
    g.Meta     `path:"/products" method:"post" tags:"Product" summary:"新增商品"`
    CategoryID int64  `json:"categoryId" v:"required|min:1#分类不能为空|分类不正确"`
    Name       string `json:"name" v:"required|length:2,128#商品名不能为空|商品名长度为2到128"`
    PriceCent  int64  `json:"priceCent" v:"min:1#价格必须大于0"`
    Stock      int    `json:"stock" v:"min:0#库存不能小于0"`
}

type CreateRes struct {
    ID int64 `json:"id"`
}

type DetailReq struct {
    g.Meta `path:"/products/{id}" method:"get" tags:"Product" summary:"商品详情"`
    ID     int64 `json:"id" in:"path" v:"min:1#商品ID不正确"`
}

type DetailRes struct {
    Product ProductItem `json:"product"`
}

type ListReq struct {
    g.Meta `path:"/products" method:"get" tags:"Product" summary:"商品列表"`
    Page   int `json:"page" in:"query" d:"1" v:"min:1#页码必须大于0"`
    Size   int `json:"size" in:"query" d:"10" v:"between:1,100#每页数量必须在1到100之间"`
}

type ListRes struct {
    List  []ProductItem `json:"list"`
    Total int           `json:"total"`
}

type UpdateReq struct {
    g.Meta     `path:"/products/{id}" method:"put" tags:"Product" summary:"更新商品"`
    ID         int64  `json:"id" in:"path" v:"min:1#商品ID不正确"`
    CategoryID int64  `json:"categoryId" v:"min:1#分类不正确"`
    Name       string `json:"name" v:"required|length:2,128#商品名不能为空|商品名长度为2到128"`
    PriceCent  int64  `json:"priceCent" v:"min:1#价格必须大于0"`
    Stock      int    `json:"stock" v:"min:0#库存不能小于0"`
    Status     int    `json:"status" v:"in:0,1#状态只能是0或1"`
}

type UpdateRes struct{}

type DeleteReq struct {
    g.Meta `path:"/products/{id}" method:"delete" tags:"Product" summary:"删除商品"`
    ID     int64 `json:"id" in:"path" v:"min:1#商品ID不正确"`
}

type DeleteRes struct{}
```

生成 Controller：

```bash
cd mall
gf gen ctrl
```

### 新增商品

在生成的 `product_v1_create.go` 中实现：

```go
func (c *ControllerV1) Create(
    ctx context.Context,
    req *v1.CreateReq,
) (res *v1.CreateRes, err error) {
    result, err := g.DB().Model("products").Ctx(ctx).Data(g.Map{
        "category_id": req.CategoryID,
        "name":        req.Name,
        "price_cent":  req.PriceCent,
        "stock":       req.Stock,
        "status":      1,
    }).Insert()
    if err != nil {
        return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "新增商品失败")
    }

    id, err := result.LastInsertId()
    if err != nil {
        return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "获取商品ID失败")
    }
    return &v1.CreateRes{ID: id}, nil
}
```

### 商品详情

```go
func (c *ControllerV1) Detail(
    ctx context.Context,
    req *v1.DetailReq,
) (res *v1.DetailRes, err error) {
    var item v1.ProductItem

    err = g.DB().Model("products").Ctx(ctx).
        Fields("id", "category_id", "name", "price_cent", "stock", "status").
        Where("id", req.ID).
        Scan(&item)
    if err != nil {
        return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "查询商品失败")
    }
    if item.ID == 0 {
        return nil, gerror.NewCode(consts.CodeProductNotFound)
    }
    return &v1.DetailRes{Product: item}, nil
}
```

### 商品列表

```go
func (c *ControllerV1) List(
    ctx context.Context,
    req *v1.ListReq,
) (res *v1.ListRes, err error) {
    model := g.DB().Model("products").Ctx(ctx)

    total, err := model.Count()
    if err != nil {
        return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "统计商品失败")
    }

    var list []v1.ProductItem
    offset := (req.Page - 1) * req.Size
    err = model.
        Fields("id", "category_id", "name", "price_cent", "stock", "status").
        OrderDesc("id").
        Limit(req.Size).
        Offset(offset).
        Scan(&list)
    if err != nil {
        return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "查询商品列表失败")
    }

    return &v1.ListRes{List: list, Total: total}, nil
}
```

### 更新商品

```go
func (c *ControllerV1) Update(
    ctx context.Context,
    req *v1.UpdateReq,
) (res *v1.UpdateRes, err error) {
    result, err := g.DB().Model("products").Ctx(ctx).
        Where("id", req.ID).
        Data(g.Map{
            "category_id": req.CategoryID,
            "name":        req.Name,
            "price_cent":  req.PriceCent,
            "stock":       req.Stock,
            "status":      req.Status,
        }).
        Update()
    if err != nil {
        return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "更新商品失败")
    }

    affected, err := result.RowsAffected()
    if err != nil {
        return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "读取更新结果失败")
    }
    if affected == 0 {
        // MySQL 中“数据完全没变化”也可能是 0，不能仅凭它断定商品不存在。
        g.Log().Infof(ctx, "product update affected=0 id=%d", req.ID)
    }
    return &v1.UpdateRes{}, nil
}
```

### 删除商品

```go
func (c *ControllerV1) Delete(
    ctx context.Context,
    req *v1.DeleteReq,
) (res *v1.DeleteRes, err error) {
    result, err := g.DB().Model("products").Ctx(ctx).
        Where("id", req.ID).
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
    return &v1.DeleteRes{}, nil
}
```

## 运行样例

```bash
gf run main.go
```

```bash
curl -X POST 'http://127.0.0.1:8000/products' \
  -H 'Content-Type: application/json' \
  -d '{"categoryId":1,"name":"Go语言鼠标垫","priceCent":3900,"stock":30}'

curl 'http://127.0.0.1:8000/products?page=1&size=10'

curl 'http://127.0.0.1:8000/products/1'

curl -X PUT 'http://127.0.0.1:8000/products/1' \
  -H 'Content-Type: application/json' \
  -d '{"categoryId":1,"name":"GoFrame 实战手册","priceCent":5990,"stock":15,"status":1}'
```

删除会影响后面课程的种子数据，练习时建议删除刚刚新建的商品 ID，不要删除 ID 1：

```bash
curl -X DELETE 'http://127.0.0.1:8000/products/新建商品ID'
```

## 本课练习：商品条件筛选

给 `ListReq` 增加四个可选 query 参数：

```go
Name       string `json:"name" in:"query"`
CategoryID int64  `json:"categoryId" in:"query"`
MinPrice   int64  `json:"minPrice" in:"query" v:"min:0#最低价格不能小于0"`
MaxPrice   int64  `json:"maxPrice" in:"query" v:"min:0#最高价格不能小于0"`
```

照着做：

1. 先创建基础 `model`。
2. `Name` 不为空时使用 `WhereLike("name", "%"+req.Name+"%")`。
3. `CategoryID > 0` 时增加分类条件。
4. `MinPrice > 0` 时使用 `WhereGTE("price_cent", req.MinPrice)`。
5. `MaxPrice > 0` 时使用 `WhereLTE("price_cent", req.MaxPrice)`。
6. 必须先把所有过滤条件拼到 `model` 上，再分别执行 `Count` 和分页 `Scan`。
7. 如果 `MaxPrice > 0 && MaxPrice < MinPrice`，返回参数错误。

不要这样写：

```go
total, _ := g.DB().Model("products").Count()
```

它会统计整张表，和筛选后的列表数量不一致。

## 验收条件

- 新增、详情、列表、更新、删除均能运行。
- `?name=GoFrame` 只返回名称包含 GoFrame 的商品。
- `?categoryId=1&minPrice=5000&maxPrice=10000` 的 `total` 与列表条件一致。
- `page=2&size=1` 不重复返回第一页数据。
- `maxPrice < minPrice` 返回稳定参数错误。
- 所有 SQL 都携带 `ctx`。
- `go test ./...` 和 `go vet ./...` 通过。

完成后提交 CRUD 请求结果和筛选代码。我先指出问题，再给参考答案。
