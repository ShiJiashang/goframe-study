# 第 2 课：路由、请求与响应

## 本课目标

掌握路由组、HTTP 方法、路径参数、查询参数以及表单/JSON 请求体的读取方式。
本课仍使用手动取参；第 3 课会用结构体自动接收参数。

## 与 Gin 的对应关系

| Gin | GoFrame | 作用 |
| --- | --- | --- |
| `engine.Group("/api")` | `server.Group("/api", callback)` | 创建具有公共前缀的路由组 |
| `group.GET(...)` | `group.GET(...)` | 注册只接受 GET 的路由 |
| `group.POST(...)` | `group.POST(...)` | 注册只接受 POST 的路由 |
| `context.Param("id")` | `request.GetRouter("id")` | 读取路径参数 |
| `context.DefaultQuery(...)` | `request.GetQuery(key, defaultValue)` | 读取 query 参数及默认值 |
| `context.PostForm(...)` | `request.GetForm(...)` | 读取表单参数 |
| `context.GetRawData()` | `request.GetBody()` | 获取原始请求体字节 |
| `context.JSON(...)` | `request.Response.WriteJson(...)` | 返回 JSON |

两者表面很像，但 GoFrame 的取参方法通常返回 `*gvar.Var`，可以继续调用
`String`、`Int`、`Int64` 等方法完成类型转换。

## 样例结构

样例监听 `8001` 端口，提供两个接口：

```text
GET  /api/products/:id
POST /api/products/preview
```

### `Server.Group`

```go
func (s *Server) Group(
    prefix string,
    groups ...func(group *ghttp.RouterGroup),
) *ghttp.RouterGroup
```

- 接收者 `s` 的类型是 `*ghttp.Server`。
- `prefix` 是组内所有路由共享的前缀，本课是 `/api`。
- `groups` 是可变参数，每个值都是一个接收 `*ghttp.RouterGroup` 的函数。
- 返回值是创建的路由组；本课在回调中完成注册，因此没有保存返回值。

```go
server.Group("/api", func(group *ghttp.RouterGroup) {
    // 注册组内路由
})
```

- `group` 是回调函数的局部变量，类型为 `*ghttp.RouterGroup`。
- 它只负责组织和注册路由，不代表某一次 HTTP 请求。
- Gin 常把路由组保存在变量中；GoFrame 两种写法都支持，本课使用回调形式。

### `RouterGroup.GET` 与 `POST`

```go
func (g *RouterGroup) GET(
    pattern string,
    object any,
    params ...any,
) *RouterGroup
```

`POST` 的签名与它相同，只是限定的 HTTP 方法不同。

- `pattern` 是组内路径，它会和 `/api` 前缀拼接。
- `object` 是处理器，本课传入 `func(*ghttp.Request)`。
- `params` 是可选的附加注册参数，本课不使用。
- 返回一个路由组指针，便于链式或继续注册；本课不接收返回值。

使用错误的 HTTP 方法访问时，不会执行对应处理器。

## GET 接口：路径参数与查询参数

```go
group.GET("/products/:id", func(request *ghttp.Request) {
    productID := request.GetRouter("id").Int()
    currency := request.GetQuery("currency", "CNY").String()
    // ...
})
```

最终路径为 `/api/products/:id`，其中 `:id` 是路径参数占位符。

### `GetRouter`

```go
func (r *Request) GetRouter(key string, def ...any) *gvar.Var
```

- `key` 是路径参数名，不包含冒号。
- `def` 是找不到参数时可选的默认值。
- 返回 `*gvar.Var`，它是 GoFrame 对动态值的包装，不是最终的 `int`。
- `.Int()` 把包装值转换成 Go 的 `int`，所以 `productID` 的类型是 `int`。

### `GetQuery`

```go
func (r *Request) GetQuery(key string, def ...any) *gvar.Var
```

- 只从 URL 的 query string 中取值。
- 本课为 `currency` 指定默认值 `"CNY"`。
- `.String()` 返回 Go 的 `string`，所以 `currency` 的类型是 `string`。

请求示例：

```bash
curl 'http://127.0.0.1:8001/api/products/42?currency=USD'
```

## POST 接口：统一读取表单或 JSON

```go
name := request.Get("name").String()
priceCent := request.Get("priceCent").Int64()
```

### `Request.Get`

```go
func (r *Request) Get(key string, def ...any) *gvar.Var
```

- `Get` 是 `GetRequest` 的简写。
- 它可以综合读取路由、query、JSON body、表单以及自定义参数。
- 如果同名参数来自多个位置，GoFrame 有自己的覆盖优先级；业务代码不应故意
  从多个位置发送同一个字段。
- `.Int64()` 返回 `int64`。金额字段使用 `int64` 表示“分”，避免浮点误差。

JSON 请求：

```bash
curl -i -X POST http://127.0.0.1:8001/api/products/preview \
  -H 'Content-Type: application/json' \
  -d '{"name":"Mechanical Keyboard","priceCent":29900}'
```

表单请求也能被同一段代码读取：

```bash
curl -i -X POST http://127.0.0.1:8001/api/products/preview \
  -d 'name=Mechanical Keyboard&priceCent=29900'
```

需要限定来源时应使用以下方法，而不是 `Get`：

```go
request.GetQuery("key") // 只读 URL query
request.GetRouter("key") // 只读路径参数
request.GetForm("key") // 只读表单
request.GetBody() // 返回原始请求体，类型为 []byte
```

### `g.Map`

```go
g.Map{
    "name":      name,
    "priceCent": priceCent,
}
```

- `g.Map` 是 GoFrame 提供的 `map[string]any` 便捷类型。
- key 的类型固定为 `string`，value 可以是任意类型。
- 它适合短小示例；正式接口会从第 3 课开始使用明确的响应结构体。

## 运行样例

```bash
go run ./lessons/lesson02
```

分别执行：

```bash
curl -i 'http://127.0.0.1:8001/api/products/42?currency=USD'

curl -i -X POST http://127.0.0.1:8001/api/products/preview \
  -H 'Content-Type: application/json' \
  -d '{"name":"Mechanical Keyboard","priceCent":29900}'
```

## 课后题：价格计算接口

在 `/api` 路由组中增加：

```text
POST /api/products/calculate
```

它接收以下 JSON：

```json
{
  "priceCent": 29900,
  "quantity": 2
}
```

返回：

```json
{
  "priceCent": 29900,
  "quantity": 2,
  "totalCent": 59800
}
```

要求：

1. 必须使用现有的 `/api` 路由组及 `group.POST`。
2. 使用 `request.Get` 分别读取两个字段。
3. `priceCent` 使用 `Int64()`，`quantity` 使用 `Int()`。
4. `totalCent` 必须由服务器计算，不能从请求直接读取。
5. 使用 `WriteJson` 返回 `g.Map`，不手写 JSON 字符串。
6. 不修改已有两个接口的行为。

本课暂不处理缺少参数、负数或溢出，这些属于第 4 课的数据校验。

验收命令：

```bash
curl -i -X POST http://127.0.0.1:8001/api/products/calculate \
  -H 'Content-Type: application/json' \
  -d '{"priceCent":29900,"quantity":2}'
```

通过条件：

- HTTP 状态码为 `200`。
- `Content-Type` 包含 `application/json`。
- `priceCent`、`quantity` 与输入一致。
- `totalCent` 等于 `59800`。
