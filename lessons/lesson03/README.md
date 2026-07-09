# 第 3 课：规范路由与结构化参数

## 本课目标

把第 2 课“在处理器中逐个取参数”的写法，改成 GoFrame 推荐的规范路由：

- 请求结构体同时声明路径、HTTP 方法和输入字段。
- GoFrame 自动把请求参数转换到请求结构体。
- Controller 使用统一的方法签名并返回响应结构体。

本课只学习结构化接收；字段校验留到第 4 课。

## 与 Gin 的关键区别

Gin 常见写法：

```go
func Create(context *gin.Context) {
    var request CreateProductReq
    if err := context.ShouldBindJSON(&request); err != nil {
        // 手动处理绑定错误
    }
    context.JSON(200, response)
}
```

GoFrame 规范路由：

```go
func (controller *ProductController) Create(
    ctx context.Context,
    req *CreateProductReq,
) (res *CreateProductRes, err error)
```

GoFrame 根据 `Req` 中的元数据注册路由、自动绑定参数，并把返回的 `Res/error`
交给响应中间件。这是后续自动校验和生成 OpenAPI 的基础。

## 请求结构体

```go
type CreateProductReq struct {
    g.Meta    `path:"/products" method:"post" tags:"Product" summary:"Create product"`
    Name      string `json:"name"`
    PriceCent int64  `json:"priceCent"`
}
```

### `type CreateProductReq struct`

- `CreateProductReq` 是新定义的结构体类型，不是变量。
- `Req` 后缀是 GoFrame 工程中接口输入模型的命名约定。
- 每次请求都会得到一个新的 `*CreateProductReq`，不会在请求之间共享数据。

### `g.Meta`

`g.Meta` 是 `gmeta.Meta` 的类型别名，本身是一个空结构体：

```go
type Meta struct{}
```

它不保存业务数据。GoFrame 通过它的 struct tag 读取接口元数据：

- `path:"/products"`：接口在当前路由组内的路径。
- `method:"post"`：只允许 POST。
- `tags:"Product"`：OpenAPI 中的接口分组。
- `summary:"Create product"`：OpenAPI 中的接口摘要。

因为外层路由组前缀是 `/api`，最终路径是 `POST /api/products`。

### 字段及 JSON 标签

- `Name` 的 Go 类型为 `string`，JSON 字段名为 `name`。
- `PriceCent` 的 Go 类型为 `int64`，JSON 字段名为 `priceCent`。
- JSON 数字会自动转换为 `int64`；当前暂不判断缺失、负数或非法字符串。

## 响应结构体

```go
type CreateProductRes struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    PriceCent int64  `json:"priceCent"`
}
```

- `Res` 后缀表示接口输出模型。
- 使用明确结构体比 `g.Map` 更容易检查类型、复用并生成接口文档。
- 返回 JSON 时使用字段的 `json` 标签，而不是 Go 字段名。

## Controller 与方法签名

```go
type ProductController struct{}
```

- 这是一个空结构体类型，用于组织商品相关接口方法。
- 当前没有依赖或状态，所以不需要字段。

```go
func (controller *ProductController) Create(
    ctx context.Context,
    req *CreateProductReq,
) (res *CreateProductRes, err error)
```

逐项说明：

- `(controller *ProductController)` 是指针接收者。
- `Create` 是方法名；路由本身来自 `CreateProductReq` 的 `g.Meta`，不是方法名。
- `ctx context.Context` 是请求级上下文，用于传递取消信号、超时和链路信息。
- `req *CreateProductReq` 是 GoFrame 自动创建并填充的请求对象。
- `res *CreateProductRes` 是具名响应返回值。
- `err error` 是具名错误返回值；`error` 是 Go 标准库内置接口。
- 本课成功路径没有错误，因此 `err` 保持接口的零值 `nil`。

```go
res = &CreateProductRes{
    ID:        1001,
    Name:      req.Name,
    PriceCent: req.PriceCent,
}
return
```

- `&CreateProductRes{...}` 创建结构体并取得指针。
- `res` 被赋值后，裸 `return` 返回当前的 `res` 和 `err`。
- ID 暂时固定为 `1001`；ORM 课程再由数据库生成。

## 注册 Controller

```go
server.Group("/api", func(group *ghttp.RouterGroup) {
    group.Middleware(ghttp.MiddlewareHandlerResponse)
    group.Bind(&ProductController{})
})
```

### `RouterGroup.Middleware`

```go
func (g *RouterGroup) Middleware(
    handlers ...ghttp.HandlerFunc,
) *RouterGroup
```

- 参数是一个或多个中间件函数。
- `ghttp.MiddlewareHandlerResponse` 接收规范路由产生的 `res/error` 并写入 JSON。
- 它默认包装成 `code`、`message`、`data`；第 5 课会拆解它的执行过程。

### `RouterGroup.Bind`

```go
func (g *RouterGroup) Bind(handlerOrObject ...any) *RouterGroup
```

- 接收一个或多个函数或 Controller 对象。
- `&ProductController{}` 的类型是 `*ProductController`。
- GoFrame 反射 Controller 的公开方法，读取其 `Req` 类型中的 `g.Meta` 并注册路由。
- 返回路由组指针，本课不需要接收。

## 运行样例

```bash
go run ./lessons/lesson03
```

请求：

```bash
curl -i -X POST http://127.0.0.1:8002/api/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Mechanical Keyboard","priceCent":29900}'
```

响应中的 `data` 应类似：

```json
{
  "id": 1001,
  "name": "Mechanical Keyboard",
  "priceCent": 29900
}
```

## 课后题：结构化商品查询

在现有文件中增加以下接口：

```text
GET /api/products/:id
```

要求：

1. 定义 `GetProductReq`，嵌入带有
   `path:"/products/:id"` 和 `method:"get"` 的 `g.Meta`。
2. `GetProductReq` 包含 `ID int64`，JSON 标签为 `id`。
3. 定义 `GetProductRes`，包含 `ID int64` 和 `Name string`。
4. 给 `ProductController` 增加 `Get` 方法，签名遵守
   `(context.Context, *GetProductReq) (*GetProductRes, error)`。
5. 返回请求中的 ID，商品名固定为 `"GoFrame Book"`。
6. 不再调用 `group.GET`，现有的 `group.Bind` 应自动注册新方法。
7. 不修改创建商品接口的行为。

验收命令：

```bash
curl -i http://127.0.0.1:8002/api/products/42
```

通过条件：

- HTTP 状态码为 `200`，响应类型为 JSON。
- 外层响应 `code` 为 `0`。
- `data.id` 为 `42`，`data.name` 为 `"GoFrame Book"`。
- 启动日志的路由表同时出现 POST 创建接口和 GET 查询接口。
