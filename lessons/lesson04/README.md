# 第 4 课：校验、转换与动态值

## 本课目标

这节课解决三个问题：

- 请求参数不合法时，如何让 GoFrame 自动拦截。
- 字符串、数字、布尔值之间如何安全转换。
- `any` 这种动态值如何临时包装后再取具体类型。

这节不会把 `gvalid`、`gconv`、`gvar` 当手册背。我们只讲样例里真正用到的 API。

## 与 Gin 的对应关系

Gin 常见写法是：

```go
var req CreateProductReq
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, ...)
    return
}

if req.PriceCent <= 0 {
    c.JSON(400, ...)
    return
}
```

GoFrame 标准路由更偏声明式：

```go
type CreateProductReq struct {
    Name      string `json:"name" v:"required|length:2,60"`
    PriceCent int64  `json:"priceCent" v:"min:1"`
    Stock     int    `json:"stock" v:"min:0"`
}
```

请求进入 Controller 前，GoFrame 会根据 `v` 标签做参数校验。校验失败时，方法不会继续执行，错误会交给 `MiddlewareHandlerResponse` 输出。

## 本课代码

运行：

```bash
go run ./lessons/lesson04
```

### 1. 自动校验商品创建

```go
type CreateProductReq struct {
    g.Meta `path:"/products" method:"post" tags:"Product" summary:"Create product with validation"`

    Name      string `json:"name" v:"required|length:2,60#商品名不能为空|商品名长度必须在2到60个字符之间"`
    PriceCent int64  `json:"priceCent" v:"min:1#价格必须大于0"`
    Stock     int    `json:"stock" v:"min:0#库存不能小于0"`
}
```

逐项解释：

- `g.Meta`：声明接口元信息。这里最终路由是 `POST /api/products`。
- `Name string`：商品名。
- `json:"name"`：请求 JSON 中的字段名。
- `v:"required|length:2,60#..."`：校验规则。
  - `required`：不能为空。
  - `length:2,60`：长度必须在 2 到 60 之间。
  - `#` 后面是自定义错误消息；多个规则对应多个消息，用 `|` 分隔。
- `PriceCent int64`：金额，单位是“分”。
- `v:"min:1"`：至少为 1，也就是价格必须大于 0。
- `Stock int`：库存数量。
- `v:"min:0"`：库存不能小于 0。

Controller 方法：

```go
func (controller *ProductController) Create(
    ctx context.Context,
    req *CreateProductReq,
) (res *CreateProductRes, err error)
```

- `ctx context.Context`：请求上下文。
- `req *CreateProductReq`：已经完成参数绑定和校验的请求对象。
- `res *CreateProductRes`：返回给客户端的数据。
- `err error`：业务错误。校验错误在进入方法前已经处理。

成功请求：

```bash
curl -i -X POST http://127.0.0.1:8003/api/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"GoFrame Book","priceCent":9900,"stock":20}'
```

错误请求：

```bash
curl -i -X POST http://127.0.0.1:8003/api/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"","priceCent":0,"stock":-1}'
```

### 2. `gconv`：类型转换

样例接口：

```text
GET /api/tools/check-price?price=9900
```

代码：

```go
priceCent := gconv.Int64(req.Price)
```

`gconv.Int64` 的作用是把输入值转换成 `int64`：

- 参数：`req.Price`，这里是 `string`。
- 返回值：`int64`。
- 如果传入 `"9900"`，得到 `9900`。
- 如果传入 `"abc"`，转换结果是 `0`，所以后面还要配合校验。

### 3. `gvalid`：手动校验

代码：

```go
if err = gvalid.New().
    Data(priceCent).
    Rules("min:1").
    Messages("价格必须大于0").
    Run(ctx); err != nil {
    return nil, err
}
```

逐项解释：

- `gvalid.New()`：创建一个新的校验器，返回 `*gvalid.Validator`。
- `.Data(priceCent)`：指定要校验的数据。
- `.Rules("min:1")`：指定规则，这里表示最小值为 1。
- `.Messages("价格必须大于0")`：指定错误消息。
- `.Run(ctx)`：执行校验。
- 返回值类型是 `gvalid.Error`，它实现了 Go 的 `error` 接口，所以可以直接赋值给 `err error`。

请求：

```bash
curl -i 'http://127.0.0.1:8003/api/tools/check-price?price=9900'
```

错误请求：

```bash
curl -i 'http://127.0.0.1:8003/api/tools/check-price?price=abc'
```

### 4. `gvar.Var`：动态值包装

样例接口：

```text
POST /api/tools/dynamic
```

代码：

```go
value := gvar.New(req.Value)
```

`gvar.New`：

- 参数：任意类型的值，类型是 `any`。
- 返回值：`*gvar.Var`。
- 作用：把动态值包装起来，后面可以按不同类型读取。

本课用到的方法：

```go
value.String()
value.Int()
value.Bool()
```

它们分别尝试把底层值转成：

- `string`
- `int`
- `bool`

请求：

```bash
curl -i -X POST http://127.0.0.1:8003/api/tools/dynamic \
  -H 'Content-Type: application/json' \
  -d '{"value":"123"}'
```

你会看到 `"123"` 同时可以被读取为字符串和整数。

## 这节课的重点

参数来源和校验责任可以这样分：

```text
Req 结构体字段：声明接口要什么参数
json/in 标签：声明参数叫什么、从哪里来
v 标签：声明参数是否合法
gconv：把一个值转成目标类型
gvalid：手动执行校验
gvar.Var：临时包装动态值，再按需要取类型
```

## 课后题：完善商品校验

在第 4 课代码中给商品创建接口新增一个字段：

```go
CategoryID int64 `json:"categoryId"`
```

要求：

1. `CategoryID` 必须大于 0。
2. 成功响应里也返回 `categoryId`。
3. 商品名必须是 2 到 60 个字符。
4. 价格必须大于 0。
5. 库存不能小于 0。
6. 不要在 Controller 方法里手写 `if req.CategoryID <= 0`，用 `v` 标签完成。

验收命令：

```bash
curl -i -X POST http://127.0.0.1:8003/api/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"GoFrame Book","categoryId":10,"priceCent":9900,"stock":20}'
```

通过条件：

- HTTP 状态码为 `200`。
- 外层 `code` 为 `0`。
- `data.categoryId` 为 `10`。

非法请求：

```bash
curl -i -X POST http://127.0.0.1:8003/api/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"A","categoryId":0,"priceCent":0,"stock":-1}'
```

通过条件：

- 返回统一 JSON。
- 外层 `code` 不是 `0`。
- `message` 中能看出参数校验失败原因。
