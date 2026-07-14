# 第 6 课：OpenAPI 与 Swagger

## 本课目标

这节课学会让 GoFrame 自动生成接口文档：

- `SetOpenApiPath`：暴露 OpenAPI JSON。
- `SetSwaggerPath`：暴露 Swagger UI 页面。
- `g.Meta`：给接口写路径、方法、分组、摘要。
- 字段标签：给参数和响应字段写来源、默认值、校验和描述。

## 与 Gin 的对应关系

Gin 通常需要额外接入 swag、注释扫描或手写文档。

GoFrame 标准路由会从 `Req/Res` 结构体里读取元数据：

```go
type CreateProductReq struct {
    g.Meta `path:"/products" method:"post" tags:"Product" summary:"Create product"`

    Name string `json:"name" v:"required" dc:"商品名称"`
}
```

也就是说，接口代码本身就是文档来源。

## 开启 OpenAPI 和 Swagger

```go
server.SetOpenApiPath("/api.json")
server.SetSwaggerPath("/swagger")
```

逐项解释：

- `server`：类型是 `*ghttp.Server`。
- `SetOpenApiPath(path string)`：设置 OpenAPI JSON 的访问路径。
- `path string`：字符串参数，例如 `"/api.json"`。
- `SetSwaggerPath(path string)`：设置 Swagger UI 的访问路径。
- `"/swagger"`：浏览器访问 `http://127.0.0.1:8005/swagger/`。

运行：

```bash
go run ./lessons/lesson06
```

访问文档：

```bash
curl -s http://127.0.0.1:8005/api.json
```

浏览器打开：

```text
http://127.0.0.1:8005/swagger/
```

## 接口元数据

```go
type ListProductsReq struct {
    g.Meta `path:"/products" method:"get" tags:"Product" summary:"List products"`
}
```

- `path:"/products"`：当前路由组下的接口路径。
- `method:"get"`：HTTP 方法。
- `tags:"Product"`：Swagger 里的接口分组。
- `summary:"List products"`：接口摘要。

外层路由组是：

```go
server.Group("/api", ...)
```

所以最终接口是：

```text
GET /api/products
```

## 字段标签

```go
Page int `json:"page" in:"query" d:"1" v:"min:1#页码必须大于0" dc:"页码"`
```

逐项解释：

- `json:"page"`：字段名。
- `in:"query"`：参数来自 query，例如 `?page=1`。
- `d:"1"`：默认值是 `1`。
- `v:"min:1#页码必须大于0"`：校验规则和错误消息。
- `dc:"页码"`：字段描述，会进入 OpenAPI/Swagger。

常见来源：

```text
in:"path"    动态路由参数
in:"query"   query 参数
in:"header"  请求头
```

POST JSON body 字段通常不需要写 `in:"body"`，保持 `json` 标签即可。

## 本课样例接口

```text
GET  /api/products
POST /api/products
GET  /api.json
GET  /swagger/
```

商品列表：

```bash
curl -s 'http://127.0.0.1:8005/api/products?page=1&size=10'
```

新增商品：

```bash
curl -s -X POST http://127.0.0.1:8005/api/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"GoFrame Book","categoryId":1,"priceCent":9900,"stock":20}'
```

## 课后题：让商品详情出现在 Swagger 中

在第 6 课代码中新增接口：

```text
GET /api/products/:id
```

要求：

1. 定义 `GetProductReq`，使用 `g.Meta` 声明 `path:"/products/:id"` 和 `method:"get"`。
2. `ID int64` 使用 `json:"id" in:"path" v:"min:1#商品ID必须大于0" dc:"商品ID"`。
3. 定义 `GetProductRes`，返回 `ID`、`Name`、`PriceCent`、`Stock`。
4. 给 `ProductController` 新增 `Get` 方法。
5. 不手写 `group.GET`，仍然依靠 `group.Bind(&ProductController{})` 自动注册。
6. Swagger 页面和 `/api.json` 都能看到这个接口。

验收命令：

```bash
curl -s http://127.0.0.1:8005/api/products/1001
curl -s http://127.0.0.1:8005/api.json | grep '/api/products/{id}'
```

通过条件：

- 商品详情接口返回 `code:0`。
- 响应 `data.id` 为 `1001`。
- `/api.json` 中能看到 `/api/products/{id}`。
