# 第 1 课：环境与第一个服务器

## 本课目标

运行一个监听 `8000` 端口的 GoFrame HTTP 服务，理解样例中每一个包、函数、
参数、类型和变量的作用。

## 与 Gin 的对应关系

| Gin | GoFrame | 作用 |
| --- | --- | --- |
| `gin.Default()` | `g.Server()` | 获取服务器对象 |
| `engine.GET(...)` | `server.BindHandler(...)` | 注册路由处理函数 |
| `*gin.Context` | `*ghttp.Request` | 表示当前 HTTP 请求及其上下文 |
| `context.String(...)` | `request.Response.Write(...)` | 写入 HTTP 响应 |
| `engine.Run(":8000")` | `server.SetPort(8000)` + `server.Run()` | 配置并启动服务 |

目前只需要建立概念映射，不必认为两套 API 的所有行为都完全相同。

## 代码逐项解释

样例位于 [`main.go`](main.go)。

### 包

```go
import (
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/ghttp"
)
```

- `frame/g` 是 GoFrame 常用组件的便捷入口。本课使用它取得服务器对象。
- `net/ghttp` 是 HTTP 服务组件。本课显式使用它的 `Request` 类型。
- `g` 和 `ghttp` 是包名，不是变量。

### `g.Server`

```go
func Server(name ...any) *ghttp.Server
```

- `name ...any` 是可变参数，可以不传；后面学习多服务实例时才会用到名字。
- 返回值 `*ghttp.Server` 是指向服务器对象的指针。
- GoFrame 会按名字复用服务器实例；本课调用 `g.Server()` 获取默认实例。

```go
server := g.Server()
```

- `server` 是局部变量。
- 编译器推断它的类型为 `*ghttp.Server`。
- 使用指针意味着后续配置和启动操作作用于同一个服务器对象。

### `SetPort`

```go
func (s *Server) SetPort(port ...int)
```

- `(s *Server)` 是方法接收者，表示该方法属于 `*ghttp.Server`。
- `port ...int` 接收一个或多个整数端口。
- 本课传入 `8000`，所以服务监听 `8000` 端口。
- 该方法没有返回值，它直接修改服务器配置。

### `BindHandler`

```go
func (s *Server) BindHandler(pattern string, handler any)
```

- `pattern string` 是路由规则，本课为 `"/hello"`。
- `handler any` 接受处理函数或其他 GoFrame 支持的处理器形式。
- 本课传入匿名函数 `func(request *ghttp.Request)`。

```go
func(request *ghttp.Request) {
    request.Response.Write("Hello, GoFrame!")
}
```

- 匿名函数没有函数名，由 `BindHandler` 保存并在请求到达时调用。
- `request` 是当前请求的局部变量，类型为 `*ghttp.Request`。
- `ghttp.Request` 内嵌标准库的 `*http.Request`，并增加了 `Response`、`Session`、
  `Cookie`、路由和中间件等能力。
- `request.Response` 的类型是 `*ghttp.Response`，负责构造当前请求的响应。

### `Response.Write`

```go
func (r *Response) Write(content ...any)
```

- `content ...any` 可接收多个任意类型的值。
- 本课传入字符串 `"Hello, GoFrame!"`。
- 方法将内容写入 GoFrame 的响应缓冲区，没有返回值。

### `Run`

```go
func (s *Server) Run()
```

- 启动 HTTP 服务并阻塞当前 goroutine。
- 正常启动后，`main` 不会继续向下执行。
- 在终端按 `Ctrl+C` 可停止服务。

## 运行样例

在仓库根目录执行：

```bash
go run ./lessons/lesson01
```

另开一个终端验证：

```bash
curl -i http://127.0.0.1:8000/hello
```

响应正文应为：

```text
Hello, GoFrame!
```

## 课后题：实现健康检查

在现有服务器上增加 `GET /ping`。访问时返回以下 JSON：

```json
{"status":"ok"}
```

要求：

1. 继续使用同一个 `server` 变量，不创建第二个服务器。
2. 处理函数参数类型必须是 `*ghttp.Request`。
3. 使用 `request.Response.WriteJson(...)` 输出，不手写 JSON 字符串。
4. 不删除或改变 `/hello` 的行为。

验收命令：

```bash
curl -i http://127.0.0.1:8000/hello
curl -i http://127.0.0.1:8000/ping
```

通过条件：两个接口均为 HTTP 200，`/ping` 响应的
`Content-Type` 包含 `application/json`，JSON 字段 `status` 等于 `ok`。

完成后把代码或两条 `curl` 的输出发给我。我会先评审，不会直接跳到第 2 课。
