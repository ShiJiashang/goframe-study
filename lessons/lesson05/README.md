# 第 5 课：中间件与统一响应

## 本课目标

这节课开始看 GoFrame 请求链路：

- 中间件函数长什么样。
- `r.Middleware.Next()` 到底在干什么。
- CORS 如何接入。
- 如何把标准路由返回的 `res/error` 包装成统一响应。

本课仍然只讲代码里用到的 API。

## 与 Gin 的对应关系

Gin 中间件常见写法：

```go
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        cost := time.Since(start)
        log.Println(cost)
    }
}
```

GoFrame 中间件写法：

```go
func AccessLogMiddleware(r *ghttp.Request) {
    start := time.Now()
    r.Middleware.Next()
    duration := time.Since(start)
}
```

对应关系：

```text
Gin 的 *gin.Context       -> GoFrame 的 *ghttp.Request
Gin 的 c.Next()           -> GoFrame 的 r.Middleware.Next()
Gin 的 c.JSON(...)        -> GoFrame 的 r.Response.WriteJson(...)
Gin 的 c.GetHeader(...)   -> GoFrame 的 r.Header.Get(...)
```

## 样例接口

运行：

```bash
go run ./lessons/lesson05
```

请求成功接口：

```bash
curl -i http://127.0.0.1:8004/api/products/42
```

请求错误接口：

```bash
curl -i http://127.0.0.1:8004/api/products/404
```

CORS 验证：

```bash
curl -i http://127.0.0.1:8004/api/products/42 \
  -H 'Origin: http://localhost:3000'
```

响应体统一为：

```json
{
  "code": 0,
  "message": "OK",
  "data": {},
  "traceId": "..."
}
```

## `ApiResponse`

```go
type ApiResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data"`
    TraceID string `json:"traceId"`
}
```

- `Code`：业务码。成功通常是 `0`。
- `Message`：给调用方看的消息。
- `Data any`：真实业务数据，类型不固定，所以用 `any`。
- `TraceID`：链路 ID，排查日志时很有用。

`any` 是 Go 里的空接口别名：

```go
type any = interface{}
```

它表示“可以接任意类型的值”。

## 商品接口

```go
type GetProductReq struct {
    g.Meta `path:"/products/:id" method:"get" tags:"Product" summary:"Get product"`

    ID int64 `json:"id" in:"path"`
}
```

- `path:"/products/:id"`：声明动态路由。
- `method:"get"`：声明 HTTP 方法。
- `ID int64`：接收商品 ID。
- `in:"path"`：明确 ID 来自路径参数。

Controller：

```go
func (controller *ProductController) Get(
    ctx context.Context,
    req *GetProductReq,
) (res *GetProductRes, err error)
```

如果 ID 是 `404`，样例故意返回错误：

```go
return nil, gerror.NewCode(gcode.CodeNotFound, "商品不存在")
```

- `gerror.NewCode`：创建带错误码的 GoFrame 错误。
- 第一个参数是 `gcode.Code`。
- 第二个参数是错误消息。
- 返回值实现了标准 `error` 接口。

## 访问日志中间件

```go
func AccessLogMiddleware(r *ghttp.Request) {
    start := time.Now()

    r.Middleware.Next()

    duration := time.Since(start)
    g.Log().Infof(r.Context(), "%s %s cost=%s", r.Method, r.URL.Path, duration)
}
```

逐项解释：

- `func AccessLogMiddleware(r *ghttp.Request)`：GoFrame 中间件函数签名。
- `r *ghttp.Request`：当前 HTTP 请求对象。
- `time.Now()`：记录开始时间，返回 `time.Time`。
- `r.Middleware.Next()`：继续执行后面的中间件和最终 Controller。
- `time.Since(start)`：计算从 `start` 到现在经过了多久，返回 `time.Duration`。
- `g.Log().Infof`：格式化输出 info 级别日志。
- `r.Method`：HTTP 方法，比如 `GET`。
- `r.URL.Path`：请求路径，比如 `/api/products/42`。

这一段最关键的是：

```text
Next 前面：请求进入时做事
Next 后面：业务执行完后做事
```

## 统一响应中间件

```go
func ResponseMiddleware(r *ghttp.Request) {
    r.Middleware.Next()

    err := r.GetError()
    res := r.GetHandlerResponse()
}
```

`r.Middleware.Next()` 执行完之后，标准路由方法的返回值会被 GoFrame 保存起来：

- `r.GetHandlerResponse()`：拿 Controller 返回的 `res`。
- `r.GetError()`：拿 Controller 返回的 `err`。

然后我们自己组装响应：

```go
r.Response.WriteJson(ApiResponse{
    Code:    code.Code(),
    Message: message,
    Data:    res,
    TraceID: gtrace.GetTraceID(r.Context()),
})
```

涉及到的 API：

- `r.Response.WriteJson(value)`：把 `value` 序列化成 JSON 写入响应。
- `gerror.Code(err)`：从 GoFrame 错误里取错误码。
- `gcode.CodeOK`：GoFrame 内置成功码。
- `gcode.CodeInternalError`：GoFrame 内置内部错误码。
- `gtrace.GetTraceID(ctx)`：从上下文里取 trace ID。

这节我们自己写了 `ResponseMiddleware`，所以没有再使用：

```go
ghttp.MiddlewareHandlerResponse
```

它是 GoFrame 默认统一响应中间件。我们这节是在学习它背后的核心思路。

## CORS

注册中间件：

```go
group.Middleware(
    ghttp.MiddlewareCORS,
    AccessLogMiddleware,
    ResponseMiddleware,
)
```

`ghttp.MiddlewareCORS` 是 GoFrame 内置 CORS 中间件，内部会调用：

```go
r.Response.CORSDefault()
```

它会设置跨域相关响应头。浏览器前端请求 API 时经常需要它。

## 中间件执行顺序

注册顺序：

```text
MiddlewareCORS -> AccessLogMiddleware -> ResponseMiddleware -> AuthMiddleware -> Controller
```

执行时大概是：

```text
CORS 进入
  AccessLog 进入，记录 start
    Response 进入
      Auth 进入，检查登录态
        Controller 执行业务
    Response 读取 res/error，写 JSON
  AccessLog 计算耗时，写日志
CORS 结束
```

## 中间件中途出错怎么终止

Gin 中常见写法是：

```go
c.AbortWithStatusJSON(401, gin.H{
    "message": "未登录",
})
return
```

GoFrame 中没有完全同名的 `AbortWithStatusJSON`，但有两种常用处理方式。

### 方式一：`SetError + return`

这是本项目更推荐的方式，因为它能继续复用统一响应中间件。

```go
func AuthMiddleware(r *ghttp.Request) {
    token := r.Header.Get("Authorization")
    if token == "" {
        r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, "未登录"))
        return
    }

    r.SetParam("userId", int64(12))
    r.Middleware.Next()
}
```

逐项解释：

- `r.Header.Get("Authorization")`：读取请求头中的 Token。
- `r.SetError(err)`：把错误保存到当前请求对象中。
- `gerror.NewCode(gcode.CodeNotAuthorized, "未登录")`：创建一个带错误码的错误。
- `return`：当前中间件直接返回。
- 没有调用 `r.Middleware.Next()`：后续中间件和 Controller 不会继续执行。
- `r.SetParam("userId", int64(12))`：类似 Gin 的 `c.Set("userId", 12)`，给后续逻辑保存请求级临时变量。

对应 Gin：

```text
Gin:      c.Set("userId", 12)
GoFrame:  r.SetParam("userId", int64(12))

Gin:      c.Get("userId")
GoFrame:  r.GetParam("userId")
```

Controller 里可以这样拿：

```go
r := ghttp.RequestFromCtx(ctx)
userID := r.GetParam("userId").Int64()
```

`r.GetParam("userId")` 返回的是动态值，所以要继续调用 `.Int64()`、`.String()` 等方法转换成具体类型。

### 为什么 `AuthMiddleware` 要放在 `ResponseMiddleware` 后面

注册顺序是：

```go
group.Middleware(
    ghttp.MiddlewareCORS,
    AccessLogMiddleware,
    ResponseMiddleware,
    AuthMiddleware,
)
```

执行链路是：

```text
ResponseMiddleware 进入
  AuthMiddleware 进入
    未登录：r.SetError(...) + return
  回到 ResponseMiddleware
ResponseMiddleware 调用 r.GetError()，统一写 JSON
```

所以未登录时，Controller 不会执行，但统一响应仍然能输出：

```json
{
  "code": 64,
  "message": "未登录",
  "data": null,
  "traceId": "...",
  "costMs": 0
}
```

核心规则：

```text
谁要统一包装错误，谁就要包在会出错的中间件外面。
```

### 方式二：直接写响应并退出

GoFrame 也支持直接写响应并终止：

```go
func AuthMiddleware(r *ghttp.Request) {
    token := r.Header.Get("Authorization")
    if token == "" {
        r.Response.WriteJsonExit(ApiResponse{
            Code:    401,
            Message: "未登录",
            Data:    nil,
            TraceID: gtrace.GetTraceID(r.Context()),
        })
        return
    }

    r.Middleware.Next()
}
```

`WriteJsonExit` 做两件事：

```text
1. 写 JSON 响应
2. 终止当前请求后续执行
```

这个方式适合简单项目或特殊接口。但在商城项目里，更推荐 `SetError + return`，因为响应格式统一由 `ResponseMiddleware` 管。

## 课后题：鉴权中间件中断请求

第 5 课代码已经加入 `AuthMiddleware`。请你运行并观察它的行为，然后做一个小改造：

```go
r.SetParam("userId", int64(12))
```

要求：

1. 未传 `Authorization` 请求头时，Controller 不能执行。
2. 未登录响应必须仍然是统一 JSON，包含 `code`、`message`、`data`、`traceId`、`costMs`。
3. 传入 `Authorization` 请求头时，请求可以正常到达 Controller。
4. Controller 能通过 `r.GetParam("userId").Int64()` 拿到中间件写入的用户 ID。
5. 不要在 Controller 里读取 `Authorization`，鉴权逻辑属于中间件。

验收命令：

```bash
curl -s http://127.0.0.1:8004/api/products/42
curl -s http://127.0.0.1:8004/api/products/42 -H 'Authorization: Bearer abc'
```

通过条件：

- 第一个响应 `code` 不是 `0`，`message` 是 `"未登录"`。
- 第二个响应 `code` 为 `0`。
- 第二个响应的 `data.name` 能看出 Controller 读取到了 `userId`。
