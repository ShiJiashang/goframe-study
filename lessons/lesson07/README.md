# 第 7 课：CLI 与工程分层

## 本课目标

第 1–6 课是小样例。从这节开始，我们进入真正的商城 API 项目：

```text
mall/
```

这节课重点不是写复杂业务，而是看懂 GoFrame 官方项目骨架：

- `gf init` 如何创建项目。
- `main.go` 如何启动命令。
- `gcmd.Command` 是什么。
- `api`、`internal/controller`、`internal/logic`、`internal/service`、`internal/model`、`internal/dao` 各自负责什么。
- `gf gen ctrl` 和 `gf gen service` 分别生成什么。

## 与 Gin 的对应关系

Gin 小项目常见结构：

```text
main.go
router.go
handler/
service/
model/
```

GoFrame 官方结构更强调分层和代码生成：

```text
api/                  接口定义：Req/Res
internal/controller/  控制器：接收请求，调用 service
internal/logic/       业务实现：真正写业务逻辑
internal/service/     服务接口：由 gf gen service 生成
internal/model/       业务输入输出模型
internal/dao/         数据库访问对象：后面由 gf gen dao 生成
internal/cmd/         启动命令
manifest/             配置、Docker、部署相关文件
```

你可以先这样类比：

```text
Gin router 注册路由       -> GoFrame api + controller + Bind
Gin handler               -> GoFrame controller
Gin service               -> GoFrame service 接口 + logic 实现
Gin model                 -> GoFrame model/entity/do
main.go 直接启动 HTTP     -> GoFrame main.go 调用 gcmd.Command
```

## 本课生成的项目

本课使用命令创建了商城项目骨架：

```bash
gf init mall -g goframe-study/mall
```

项目目录：

```text
mall/
```

进入项目：

```bash
cd mall
```

运行：

```bash
go run main.go
```

访问：

```bash
curl -i http://127.0.0.1:8000/hello
```

如果端口被占用，可以先停掉之前课程的服务。

## `main.go`

```go
package main

import (
    "github.com/gogf/gf/v2/os/gctx"

    "goframe-study/mall/internal/cmd"
)

func main() {
    cmd.Main.Run(gctx.GetInitCtx())
}
```

逐项解释：

- `package main`：表示这是可执行程序入口。
- `gctx.GetInitCtx()`：获取初始化上下文，类型是 `context.Context`。
- `cmd.Main`：定义在 `internal/cmd/cmd.go` 里的启动命令。
- `Run(ctx)`：执行命令。

Gin 项目里你可能直接在 `main.go` 写：

```go
r := gin.Default()
r.Run(":8000")
```

GoFrame 官方骨架把启动逻辑放进 `internal/cmd`，这样后面可以扩展多个命令，比如：

```text
server
migrate
worker
cron
```

## `gcmd.Command`

`mall/internal/cmd/cmd.go` 里核心结构是：

```go
var Main = gcmd.Command{
    Name:  "main",
    Usage: "main",
    Brief: "start http server",
    Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
        s := g.Server()
        s.Group("/", func(group *ghttp.RouterGroup) {
            group.Middleware(ghttp.MiddlewareHandlerResponse)
            group.Bind(
                hello.NewV1(),
            )
        })
        s.Run()
        return nil
    },
}
```

字段解释：

- `Name`：命令名称。
- `Usage`：命令用法说明。
- `Brief`：命令简短描述。
- `Func`：命令真正执行的函数。
- `ctx context.Context`：命令上下文。
- `parser *gcmd.Parser`：命令行参数解析器。
- `err error`：命令执行失败时返回错误。

`Func` 里面的 HTTP 启动逻辑和前几课是一样的：

```go
s := g.Server()
```

创建服务器。

```go
group.Middleware(ghttp.MiddlewareHandlerResponse)
```

使用默认统一响应中间件。

```go
group.Bind(hello.NewV1())
```

绑定生成的 Controller。

```go
s.Run()
```

启动 HTTP 服务。

## `api` 与 `controller`

接口定义在：

```text
mall/api/hello/v1/hello.go
```

里面有：

```go
type HelloReq struct {
    g.Meta `path:"/hello" tags:"Hello" method:"get" summary:"You first hello api"`
}

type HelloRes struct {
    g.Meta `mime:"text/html" example:"string"`
}
```

- `HelloReq`：请求结构体。
- `HelloRes`：响应结构体。
- `g.Meta`：接口元数据。
- `path:"/hello"`：路由路径。
- `method:"get"`：HTTP 方法。

Controller 在：

```text
mall/internal/controller/hello/
```

其中：

```text
hello_new.go       由 gf 生成和维护，不要手改
hello_v1_hello.go 业务实现文件，可以写逻辑
```

`hello_new.go` 里有：

```go
func NewV1() hello.IHelloV1 {
    return &ControllerV1{}
}
```

它的作用是创建 v1 版本的 Hello Controller。

## `gf gen ctrl`

命令：

```bash
gf gen ctrl
```

作用：

- 扫描 `api/` 目录。
- 根据 `Req/Res` 生成 Controller 接口和骨架。
- 默认输出到 `internal/controller`。

常用参数：

```bash
gf gen ctrl -s api -d internal/controller
```

- `-s` / `--srcFolder`：接口定义目录，默认 `api`。
- `-d` / `--dstFolder`：生成 Controller 的目录，默认 `internal/controller`。

注意：

```text
带 “DO NOT EDIT” 的文件不要手改。
只生成一次或明确标注可写的实现文件，才写业务逻辑。
```

## `gf gen service`

命令：

```bash
gf gen service
```

作用：

- 扫描 `internal/logic`。
- 根据 logic 中的结构体生成 `internal/service` 接口。
- Controller 后面通过 service 调用业务逻辑。

默认约定：

```text
internal/logic/product/product.go  业务实现
internal/service/product.go        生成的服务接口
```

这节先知道它的职责。第 14 课会重点练 Controller、Service、Logic 的协作。

## 常用 `gf` 指令速查

下面这些指令是后面商城项目会反复用到的。你不用现在全背，但要知道“什么时候用哪个”。

本课统一使用 `gf` 命令，不再写绝对路径。如果终端提示 `gf: command not found`，先把 Go bin 目录加入 PATH。

### 查看版本

```bash
gf -v
```

作用：

- 查看当前安装的 GoFrame CLI 版本。
- 查看当前项目使用的 GoFrame 依赖版本。
- 排查“CLI 版本”和“项目依赖版本”不一致的问题。

常见输出里重点看：

```text
CLI Detail
GF Version(go.mod)
```

### 查看帮助

格式：

```bash
gf 命令 -h
```

例子：

```bash
gf init -h
gf gen ctrl -h
gf gen service -h
```

作用：

- 查看命令支持哪些参数。
- 忘记格式时先看帮助，不要靠猜。

### `gf init`：创建项目

格式：

```bash
gf init 项目名
```

本课实际使用：

```bash
gf init mall -g goframe-study/mall
```

参数解释：

- `mall`：要创建的项目目录。
- `-g` / `--module`：指定 Go module 名称。
- `goframe-study/mall`：本项目的 module 名。

作用：

- 创建 GoFrame 官方推荐项目骨架。
- 生成 `api`、`internal`、`manifest`、`hack` 等目录。
- 生成默认 `/hello` 接口。

什么时候用：

```text
新建一个 GoFrame 项目时使用。
```

### `gf run`：开发运行项目

格式：

```bash
gf run main.go
```

在本项目中：

```bash
cd mall
gf run main.go
```

作用：

- 运行 GoFrame 项目。
- 开发阶段比直接 `go run main.go` 更贴近 GoFrame CLI 工作流。

当前你也可以用：

```bash
go run main.go
```

两者都能启动服务。课程里如果只是验证接口，`go run main.go` 也可以。

### `gf gen ctrl`：根据 API 定义生成 Controller

格式：

```bash
gf gen ctrl
```

完整常见格式：

```bash
gf gen ctrl -s api -d internal/controller
```

参数解释：

- `-s` / `--srcFolder`：接口定义目录，默认是 `api`。
- `-d` / `--dstFolder`：Controller 生成目录，默认是 `internal/controller`。

作用：

- 扫描 `api` 目录下的 `Req/Res`。
- 生成 Controller 接口文件。
- 生成 Controller 构造函数。
- 生成待实现的 Controller 方法文件。

什么时候用：

```text
新增或修改 api/ 下的接口定义后使用。
```

注意：

```text
标有 DO NOT EDIT 的文件不要手改。
业务逻辑写在非 DO NOT EDIT 的实现文件里。
```

### `gf gen service`：根据 Logic 生成 Service 接口

格式：

```bash
gf gen service
```

完整常见格式：

```bash
gf gen service -s internal/logic -d internal/service
```

参数解释：

- `-s` / `--srcFolder`：Logic 源码目录，默认是 `internal/logic`。
- `-d` / `--dstFolder`：Service 接口生成目录，默认是 `internal/service`。

作用：

- 扫描 `internal/logic`。
- 根据符合规则的结构体生成 `internal/service` 接口。
- 让 Controller 依赖接口，而不是直接依赖具体 Logic。

什么时候用：

```text
新增或修改 logic 服务方法后使用。
```

### `gf gen dao`：根据数据库表生成 DAO

格式示例：

```bash
gf gen dao
```

作用：

- 读取数据库表结构。
- 生成 `internal/dao`、`internal/model/do`、`internal/model/entity`。

什么时候用：

```text
第 13 课接入 MySQL 后使用。
```

注意：

```text
DAO 生成代码大多不要手改。
数据库字段变化后重新生成。
```

这节课先知道它负责“数据库访问层代码生成”，暂时不用执行。

### `gf build`：构建二进制

格式：

```bash
gf build
```

作用：

- 编译 GoFrame 项目。
- 输出可部署的二进制文件。

什么时候用：

```text
第 22 课上线前收尾时使用。
```

开发阶段主要还是：

```bash
go test ./...
go vet ./...
go run main.go
```

## 常用 Go 指令

虽然这节是 GoFrame CLI，但 Go 项目日常也离不开这些命令。

### 运行测试

```bash
go test ./...
```

作用：

- 运行当前 module 下所有包的测试。
- `./...` 表示递归所有子目录。

### 静态检查

```bash
go vet ./...
```

作用：

- 检查可疑代码。
- 比如格式化字符串参数不匹配、不可达代码等。

### 整理依赖

```bash
go mod tidy
```

作用：

- 删除没用到的依赖。
- 补齐代码实际用到但 `go.mod` 缺失的依赖。

注意：

```text
它会修改 go.mod 和 go.sum。
只有在确实需要整理依赖时再运行。
```

## 当前推荐工作流

从第 7 课开始，我们以后大体按这个节奏写商城项目：

```text
1. 在 api/ 定义接口 Req/Res
2. 运行 gf gen ctrl
3. 在 internal/controller/ 填控制器逻辑
4. 复杂业务下沉到 internal/logic
5. 运行 gf gen service 生成服务接口
6. Controller 调用 service
7. 后面接 MySQL 时运行 gf gen dao
```

## 课后题：新增商品 API 定义

在 `mall/` 项目里新增商品接口定义，不要求你现在写完整业务。

目标接口：

```text
GET /product/list
```

要求：

1. 新建文件：

```text
mall/api/product/v1/product.go
```

2. 定义 `ListReq`：

```go
type ListReq struct {
    g.Meta `path:"/product/list" method:"get" tags:"Product" summary:"商品列表"`

    Page int `json:"page" in:"query" d:"1" v:"min:1#页码必须大于0" dc:"页码"`
    Size int `json:"size" in:"query" d:"10" v:"between:1,100#每页数量必须在1到100之间" dc:"每页数量"`
}
```

3. 定义 `ListRes`：

```go
type ListRes struct {
    List  []ProductItem `json:"list" dc:"商品列表"`
    Total int           `json:"total" dc:"总数量"`
}
```

4. 定义 `ProductItem`：

```go
type ProductItem struct {
    ID        int64  `json:"id" dc:"商品ID"`
    Name      string `json:"name" dc:"商品名称"`
    PriceCent int64  `json:"priceCent" dc:"价格，单位：分"`
}
```

5. 在 `mall/` 目录运行：

```bash
gf gen ctrl
```

验收条件：

- 生成 `mall/api/product/product.go`。
- 生成 `mall/internal/controller/product/` 相关文件。
- 不手改任何标注 `DO NOT EDIT` 的文件。
- `cd mall && go test ./...` 通过。
