# 第 22 课：上线前收尾与综合题

## 本课目标

这一课把“本机能跑”整理成“可以交付和排查”：

- 存活检查与就绪检查。
- 结构化日志、TraceID 和关键业务字段。
- GoFrame HTTP Server 的优雅关闭。
- 配置与密钥边界。
- `gf build` 构建二进制。
- Docker Compose 启动商城、MySQL、Redis 和本地支付 Mock。
- 最终综合题与完整验收。

本课不会展开 Kubernetes、服务注册、消息队列或前端页面。

## 本课代码已经实现

代码入口：

```text
lessons/lesson22/app/main.go
```

本课 Docker 文件也已经对齐这个入口：

```text
lessons/lesson22/Dockerfile
lessons/lesson22/docker-compose.yaml
lessons/lesson22/config.docker.yaml
lessons/lesson22/.env.example
```

先本地运行：

```bash
go run lessons/lesson22/app/main.go
```

默认监听：

```text
http://127.0.0.1:8022
```

本地没有 MySQL/Redis 配置时，`/health/live` 会成功，`/health/ready` 会失败，这是正常的。

可以直接测试：

```bash
curl http://127.0.0.1:8022/health/live
curl http://127.0.0.1:8022/health/ready
curl http://127.0.0.1:8022/system/config-safe
curl -X POST http://127.0.0.1:8022/system/log-demo \
  -H 'Content-Type: application/json' \
  -d '{"orderId":1001,"userId":12,"idempotencyKey":"demo-key"}'
curl 'http://127.0.0.1:8022/system/slow?seconds=5'
curl -X POST 'http://127.0.0.1:8022/system/shutdown-after?seconds=5'
```

本课实现的接口：

| 接口 | 作用 |
| --- | --- |
| `GET /health/live` | 存活检查，只证明进程活着 |
| `GET /health/ready` | 就绪检查，检查 MySQL 和 Redis |
| `GET /system/config-safe` | 查看安全配置状态，不返回密钥原文 |
| `POST /system/log-demo` | 输出一条带业务字段和 TraceID 的结构化日志 |
| `GET /system/slow?seconds=5` | 慢请求，用来观察优雅关闭 |
| `POST /system/shutdown-after?seconds=5` | 5 秒后关闭服务，教学演示用 |

统一响应格式：

```go
type ApiResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data"`
    TraceID string `json:"traceId"`
    CostMs  int64  `json:"costMs"`
}
```

## 和 Gin 的对应关系

Gin 项目通常自己组合 `http.Server`、信号监听、日志库、配置库和健康检查。GoFrame 已经提供 Server、配置、日志、Trace、数据库和 Redis 组件，但业务是否可观测、错误是否可控，仍由项目代码决定。

上线能力不是多一个“deploy 包”，而是贯穿每层：

```text
请求进入 → TraceID → 结构化日志 → Controller → Logic → MySQL/Redis/支付
                                            ↓
                              健康检查、超时、错误码、优雅退出
```

## 存活与就绪不是一回事

建议提供两个接口：

| 接口 | 回答的问题 | 是否检查依赖 |
| --- | --- | --- |
| `/health/live` | Go 进程是否活着 | 不检查 |
| `/health/ready` | 当前实例能否接业务流量 | 检查 MySQL、Redis |

如果 Redis 短暂故障，进程仍然是活的，不应该因为存活检查失败而反复重启；但就绪检查应失败，避免继续接收新流量。

## 健康检查 API 类型

```go
type LiveReq struct {
	g.Meta `path:"/health/live" method:"get" tags:"System" summary:"存活检查"`
}

type LiveRes struct {
	Status string `json:"status" dc:"进程状态"`
}

type ReadyReq struct {
	g.Meta `path:"/health/ready" method:"get" tags:"System" summary:"就绪检查"`
}

type ReadyRes struct {
	Status string `json:"status"`
	MySQL  string `json:"mysql"`
	Redis  string `json:"redis"`
}
```

- `LiveReq`、`ReadyReq`：结构化路由请求类型。
- `g.Meta`：声明路径、方法、OpenAPI 标签和摘要。
- `dc`：字段描述，会出现在 OpenAPI 中。
- `LiveRes`、`ReadyRes`：业务响应数据；统一响应中间件仍会在外层包装 `code/message/data/traceId`。

本课实际 Controller：

```go
func (c *SystemController) Live(
	ctx context.Context,
	_ *LiveReq,
) (*LiveRes, error) {
	return &LiveRes{Status: "ok"}, nil
}

func (c *SystemController) Ready(
	ctx context.Context,
	_ *ReadyReq,
) (*ReadyRes, error) {
	res := &ReadyRes{
		Status: "ok",
		MySQL:  "ok",
		Redis:  "ok",
	}

	if err := g.DB().PingMaster(); err != nil {
		res.Status = "not_ready"
		res.MySQL = "error"
		g.Log().Error(ctx, "readiness mysql failed", g.Map{"error": err.Error()})
		setHTTPStatus(ctx, http.StatusServiceUnavailable)
		return res, nil
	}
	if _, err := g.Redis().Do(ctx, "PING"); err != nil {
		res.Status = "not_ready"
		res.Redis = "error"
		g.Log().Error(ctx, "readiness redis failed", g.Map{"error": err.Error()})
		setHTTPStatus(ctx, http.StatusServiceUnavailable)
		return res, nil
	}
	return res, nil
}
```

### 用到的函数

```go
g.DB(name ...string) gdb.DB
```

返回指定分组的数据库对象；不传名称时使用 `default`。

```go
PingMaster() error
```

检查主数据库连接是否可用，成功返回 `nil`。

```go
g.Redis(name ...string) *gredis.Redis
```

返回 Redis 客户端；`Do(ctx, "PING")` 发送原始 Redis 命令。

健康接口不要把数据库密码、内部地址或底层错误原文返回给外部。详细错误写日志，对外只返回稳定状态。

## 结构化日志

不利于检索的日志：

```go
g.Log().Infof(ctx, "order %d failed for user %d", orderID, userID)
```

更适合机器检索的字段：

```go
g.Log().Error(ctx, "create order failed", g.Map{
	"orderId":       orderID,
	"userId":        userID,
	"idempotencyKey": idempotencyKey,
	"error":         err,
})
```

- 第一个参数 `ctx`：让日志自动关联当前 Trace。
- 第二个参数：稳定事件名，不要每次拼成不同句子。
- `g.Map`：`map[string]any` 的便捷类型，保存可检索字段。

建议记录：

- 请求：方法、路径、耗时、HTTP 状态、TraceID。
- 订单：订单号、用户 ID、幂等键、状态变化。
- 外部调用：服务名、耗时、结果、重试次数。
- 定时任务：任务名、扫描数、成功数、失败数。

不要记录：

- 明文密码。
- JWT、Session ID、Cookie。
- 完整银行卡号、身份证号。
- 配置中的密钥。

## TraceID

接口层创建 Span：

```go
ctx, span := gtrace.NewSpan(ctx, "api.order.create")
defer span.End()

traceID := gtrace.GetTraceID(ctx)
```

- `NewSpan`：基于原 `ctx` 创建新的链路片段，并返回新上下文和 Span。
- `span.End()`：函数结束时关闭该片段。
- `GetTraceID`：读取链路 ID；它不负责生成。
- 当原上下文没有有效链路时，`NewSpan` 会建立链路，因此接口入口需要创建 Span 或由已经配置好的追踪中间件创建。

之后必须继续传新 `ctx`：

```go
service.Order().Create(ctx, input)
g.Log().Info(ctx, "create order")
g.DB().Model(...).Ctx(ctx)
g.Client().Post(ctx, url, data)
```

如果中途换成 `context.Background()`，Trace 就断了。

## 优雅关闭

`ghttp.Server` 的两个相关方法：

```go
Run()
Shutdown() error
```

- `Run`：启动 HTTP 服务并阻塞当前 goroutine。
- `Shutdown`：停止接收新请求，等待正在处理的请求按框架策略收尾。

GoFrame Server 已包含系统信号和优雅退出处理，正常用 `Run()` 启动即可。在终端按 `Ctrl+C` 时，观察日志并确认正在执行的请求不会被粗暴截断。

本课为了让你不用手动按 `Ctrl+C`，额外写了一个教学接口：

```go
type ShutdownAfterReq struct {
    g.Meta `path:"/system/shutdown-after" method:"post" tags:"System" summary:"延迟关闭服务，教学演示用"`

    Seconds int `json:"seconds" in:"query" d:"5"`
}

func (c *SystemController) ShutdownAfter(ctx context.Context, req *ShutdownAfterReq) (*ShutdownAfterRes, error) {
    if req.Seconds <= 0 {
        req.Seconds = 5
    }
    if req.Seconds > 30 {
        return nil, gerror.NewCode(gcode.CodeValidationFailed, "seconds must be <= 30")
    }

    seconds := req.Seconds
    g.Log().Warning(ctx, "shutdown scheduled", g.Map{"afterSeconds": seconds})

    go func() {
        time.Sleep(time.Duration(seconds) * time.Second)
        bgCtx := context.Background()
        g.Log().Warning(bgCtx, "shutdown starting", g.Map{"afterSeconds": seconds})
        if err := c.server.Shutdown(); err != nil {
            g.Log().Error(bgCtx, "shutdown failed", g.Map{"error": err.Error()})
            return
        }
        g.Log().Info(bgCtx, "shutdown completed")
    }()

    return &ShutdownAfterRes{
        Scheduled:     true,
        AfterSeconds: seconds,
    }, nil
}
```

测试方式：

```bash
go run lessons/lesson22/app/main.go
curl -X POST 'http://127.0.0.1:8022/system/shutdown-after?seconds=5'
```

请求会先返回成功，服务会在 5 秒后调用：

```go
c.server.Shutdown()
```

注意：这个接口只用于学习演示，真实生产不要暴露“远程关闭服务”的 HTTP 接口。

自己的后台资源仍需自己管理：

- cron/worker 收到取消信号后不要再领取新任务。
- 长任务应持续检查 `ctx.Done()`。
- 日志缓冲、消息连接等自建资源要在退出前关闭。
- 不要在业务代码里直接 `os.Exit`，它会跳过 `defer`。

## 配置和密钥

仓库可以提交：

- 配置字段结构。
- 本地开发默认值。
- `.env.example`。

仓库不能提交：

- 生产数据库密码。
- JWT 真实密钥。
- Redis 密码。
- 支付平台私钥或证书。

本课 `config.docker.yaml` 只包含本地 Compose 凭证。Compose 用环境变量覆盖 JWT 配置：

```yaml
environment:
  AUTH_JWTSECRET: ${JWT_SECRET:-local-compose-only-change-me}
```

GoFrame 配置键 `auth.jwtSecret` 可由对应环境变量 `AUTH_JWTSECRET` 覆盖。部署前应设置足够长的随机密钥：

```bash
cd lessons/lesson22
cp .env.example .env
# 编辑 .env，不要把它提交到 Git
```

代码继续从配置组件读取，不要直接散落 `os.Getenv`：

```go
value, err := g.Cfg().GetEffective(ctx, "auth.jwtSecret")
if err != nil {
	return gerror.Wrap(err, "读取 JWT 密钥失败")
}
secret := value.Bytes()
```

这里必须用 `GetEffective` 才会按优先级读取环境变量；普通的 `Get/MustGet` 只读取配置内容。

## `gf build`

先确认命令可用：

```bash
gf version
gf build -h
```

本课代码在仓库根模块下，所以在仓库根目录构建：

```bash
gf build lessons/lesson22/app/main.go -n lesson22 -p ./bin
```

参数：

- `lessons/lesson22/app/main.go`：构建入口文件。
- `-n lesson22`：输出二进制名。
- `-p ./bin`：输出目录。

交叉构建 Linux amd64：

```bash
gf build lessons/lesson22/app/main.go -n lesson22 -a amd64 -s linux -p ./bin
```

- `-a` / `--arch`：目标 CPU 架构。
- `-s` / `--system`：目标操作系统。
- `-v` / `--version`：写入构建版本。
- `-e` / `--extra`：额外传给 `go build` 的参数。

`gf build` 是 `go build` 的便捷封装，最终产物仍是 Go 二进制。

## Dockerfile

本课提供 `Dockerfile`，分为两阶段：

```text
golang 镜像：下载依赖并编译 /out/lesson22
        ↓
alpine 镜像：只复制二进制并以非 root 用户运行
```

关键语句：

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/lesson22 ./lessons/lesson22/app
```

- `CGO_ENABLED=0`：生成不依赖系统 C 库的二进制。
- `GOOS=linux`：目标系统是容器 Linux。
- `-trimpath`：去除本机构建路径。
- `-ldflags="-s -w"`：去掉部分调试符号，缩小学习镜像。

## Docker Compose

本课文件包含四个服务：

| 服务 | 作用 | 容器内访问名 |
| --- | --- | --- |
| `app` | 商城 API | `app:8000` |
| `mysql` | MySQL 8 | `mysql:3306` |
| `redis` | Redis 7 | `redis:6379` |
| `mock-payment` | 本地支付服务 | `mock-payment:9001` |

容器之间不能用 `127.0.0.1` 找另一个容器，所以 Docker 配置中的数据库地址是 `mysql`，Redis 地址是 `redis`。

这份 Compose 是第 22 课的运行模板。`app` 容器启动后会读取 `config.docker.yaml`，通过容器名访问 `mysql:3306`、`redis:6379` 和 `mock-payment:9001`。

启动：

```bash
cd lessons/lesson22
docker compose version
docker compose up --build
```

- `docker compose version`：确认 Compose 插件已经安装；若提示 `unknown command`，先在 Docker Desktop 中启用/安装 Compose 插件。
- `up`：创建并启动服务。
- `--build`：启动前重新构建第 22 课应用镜像。

查看状态：

```bash
docker compose ps
docker compose logs -f app
```

测试：

```bash
curl http://127.0.0.1:8000/health/live
curl http://127.0.0.1:8000/health/ready
```

停止但保留数据：

```bash
docker compose down
```

删除数据库与 Redis volume，重新初始化：

```bash
docker compose down -v
```

`-v` 会永久删除该 Compose 项目的本地数据，只在确认不需要数据时执行。

注意：`lesson11/schema.sql` 只会在 MySQL 数据卷第一次创建时自动执行。以后改表要用迁移脚本，不能指望每次重启重放初始化 SQL。

## 上线前检查单

- 配置：生产密钥来自部署环境，不在 Git 中。
- 数据库：迁移已备份、可回滚，连接池有上限。
- Redis：缓存失效、Session、撤销列表和幂等键都有清晰 TTL。
- HTTP：外部调用都有超时，响应体会关闭。
- 鉴权：错误信息不泄露密码、token、内部堆栈。
- 日志：含 TraceID 和关键业务字段，不含敏感数据。
- 任务：多实例执行策略明确。
- 进程：能通过存活/就绪检查并优雅退出。
- 测试：单元、集成、竞态检查、vet 全部通过。

## 最终综合题：独立完成轻量商城 API

不要先看完整答案。按照以下顺序完成：

1. 用户、分类、商品、订单、订单明细五张表。
2. 商品新增、查询、修改、删除、筛选、分页。
3. 创建、查询、取消订单；金额统一用整数“分”。
4. 事务扣库存，20 个并发请求下不超卖。
5. Redis Cache Aside 缓存商品详情，更新后失效。
6. Session 登录、退出和后台保护。
7. JWT Access/Refresh、退出撤销和管理员权限。
8. 支付客户端的成功、超时和异常处理；支付回调幂等。
9. 定时取消超时订单并归还库存；创建订单幂等。
10. 统一响应、TraceID、OpenAPI/Swagger、健康检查。
11. 核心自动测试与 Docker Compose。

## 最终验收条件

- `go test ./...`、`go test -race ./...`、`go vet ./...` 通过。
- 非法参数得到稳定的 `code/message/data/traceId` 响应。
- 并发下不超卖，事务失败能够完整回滚。
- 商品更新后缓存立即失效，下次读取重新建立缓存。
- Session 登录与退出可验证。
- JWT Access、Refresh、退出撤销和管理员权限可验证。
- 重复创建订单或重复支付回调不会重复扣库存或重复更新。
- MySQL、Redis、支付服务异常时有明确日志和可控响应。
- `/health/live` 与 `/health/ready` 行为符合各自职责。
- `gf build` 能生成商城二进制。
- `docker compose up --build` 能启动应用、MySQL、Redis 和 Mock 支付。
- 关闭应用时正在处理的请求能收尾，日志中没有敏感配置。

最终提交时请提供：项目目录树、接口清单、测试输出、并发测试结果、Compose 状态和你认为风险最高的三段代码。我会逐项验收并给修改提示，验收完成后再给最终参考方案。
