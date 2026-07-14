# GoFrame v2.10 实战学习计划：轻量商城 API

## 学习方式

- 共 22 课，每课 45–60 分钟，建议每周 3 课。
- 使用 Go 1.26.3、GoFrame v2.10.2、MySQL 8、Redis 7。
- 前 6 课用小样例熟悉 GoFrame 常用组件。
- 第 7 课开始持续完善“轻量商城 API”项目。
- 每课固定包含：
  1. Gin 与 GoFrame 的对应关系。
  2. 逐一解释样例中出现的包、函数签名、接收者、参数、返回值、类型、接口和关键变量。
  3. 可直接运行的样例代码。
  4. 一道练习及明确验收条件。
  5. 学员提交代码后，先评审和提示，再给参考答案。
- 只讲样例真正用到的 API，不把整个包当手册背诵。

## 第一阶段：Web 基础与标准接口

### 1. 环境与第一个服务器

- 核心包/API：`frame/g`、`net/ghttp`、`g.Server`、`ghttp.Server`、`BindHandler`、`Run`、`Request`、`Response`。
- 样例：实现 `/hello`。
- 练习：实现 `/ping`。

### 2. 路由、请求与响应

- 核心包/API：`RouterGroup`、GET/POST、动态路由参数、query 参数、form 参数、body 参数、JSON 响应。
- 样例：商品预览接口。
- 练习：实现价格计算接口。

### 3. 规范路由与结构化参数

- 核心包/API：`g.Meta`、`XxxReq/XxxRes`、`context.Context`、Controller 方法、`Bind`。
- 样例：结构化新增商品接口。
- 练习：定义结构化商品查询接口。

### 4. 校验、转换与动态值

- 核心包/API：`gvalid`、`gconv`、`gvar.Var`、校验标签、自动转换、校验错误。
- 样例：带校验的商品创建接口、价格转换接口、动态值接口。
- 练习：校验商品名、分类、价格和库存。

### 5. 中间件与统一响应

- 核心包/API：`ghttp.MiddlewareHandlerResponse`、`r.Middleware.Next()`、CORS、自定义响应结构、`GetHandlerResponse`、`GetError`。
- 样例：统一响应、访问日志、CORS。
- 练习：实现耗时统计和统一错误响应。

### 6. OpenAPI 与 Swagger

- 核心包/API：`SetOpenApiPath`、`SetSwaggerPath`、接口元数据、字段描述。
- 样例：给已有接口生成 OpenAPI/Swagger 文档。
- 练习：让前面接口完整出现在 Swagger 中。

## 第二阶段：工程结构与基础组件

### 7. CLI 与工程分层

- 核心包/API：`gf init`、`gf gen ctrl`、`gf gen service`、`api/controller/logic/service/model/dao`、`gcmd.Command`。
- 目标：理解 GoFrame 官方推荐工程结构和代码生成。
- 练习：创建商城项目骨架。

### 8. 配置管理

- 核心包/API：`g.Cfg()`、`gcfg.Config`、`Get/MustGet`、配置分组、环境变量覆盖。
- 目标：建立开发、测试配置。
- 练习：配置服务端口、数据库、Redis。

### 9. 日志、错误与上下文

- 核心包/API：`g.Log()`、`glog.Logger`、`gerror.New`、`gerror.Wrap`、`gerror.Code`、`gcode.Code`、TraceID。
- 目标：定义商城错误码并统一映射 HTTP 响应。
- 练习：实现商品、订单、鉴权错误码。

### 10. 常用数据工具

- 核心包/API：`gjson`、`gtime`、`gconv`、`gvar`、泛型容器。
- 目标：熟悉业务中常见 JSON、时间、金额、动态值处理。
- 练习：解析支付 JSON，转换时间和金额字段。

## 第三阶段：MySQL、ORM 与业务分层

### 11. 连接 MySQL 与认识 ORM

- 核心包/API：MySQL 驱动、`g.DB()`、`gdb.DB`、`gdb.Model`、`Ctx`、连接池。
- 目标：配置数据库连接并理解 ORM 基本对象。
- 练习：实现数据库健康检查。

### 12. ORM 增删改查

- 核心包/API：`Data`、`Where`、`Fields`、`Insert`、`Update`、`Delete`、`Scan`、`One`、`All`、`Limit`、`Offset`、`Count`。
- 目标：完成商品 CRUD、筛选和分页。
- 练习：实现商品列表分页和条件筛选。

### 13. DAO 代码生成与模型区别

- 核心包/API：`gf gen dao`、`dao`、`do`、`entity`、API 模型、业务模型。
- 目标：明确哪些代码可手写、哪些代码不可手改。
- 练习：生成商品、分类、用户、订单相关 DAO。

### 14. Controller、Service、Logic 协作

- 核心内容：生成的服务接口、注册函数、输入输出模型、依赖方向。
- 目标：把 Controller 从“写业务”调整为“接收请求并调用 Service”。
- 练习：把商品 CRUD 从 Controller 重构到 Logic。

### 15. 事务、库存与并发

- 核心包/API：`gdb.DB.Transaction`、`gdb.TX`、行锁、回滚、事务中的 `ctx`。
- 目标：创建订单时保证并发下库存不会变成负数。
- 练习：实现下单扣库存和事务回滚。

## 第四阶段：缓存与两套鉴权

### 16. Redis 与缓存模式

- 核心包/API：`g.Redis()`、`gredis.Redis`、`gcache.Cache`、TTL、序列化、Cache Aside。
- 目标：缓存商品详情并在更新后失效。
- 练习：实现商品详情缓存。

### 17. GoFrame Session 鉴权

- 核心包/API：`ghttp.Session`、`Set`、`Get`、`Remove`、`Clear`、Redis 会话存储、认证中间件。
- 目标：实现后台 Session 登录、退出和接口保护。
- 练习：保护商品管理接口。

### 18. JWT 鉴权

- 核心包/API：`jwt/v5`、bcrypt、Claims、Access Token、Refresh Token、Redis 撤销列表。
- 目标：实现前台 JWT 登录、刷新、退出失效和管理员权限中间件。
- 练习：实现刷新 Token 和登出撤销。

## 第五阶段：生产项目能力

### 19. HTTP Client 与外部支付

- 核心包/API：`g.Client()`、`gclient.Client`、`gclient.Request`、`gclient.Response`、Context、超时和错误处理。
- 目标：用本地 Mock 支付服务模拟外部调用。
- 练习：处理支付超时和重复回调。

### 20. 定时任务、订单过期与幂等

- 核心包/API：`gcron`、任务 Entry、单实例执行、幂等键。
- 目标：自动取消超时订单并归还库存。
- 练习：实现订单超时取消任务。

### 21. 测试与可替换接口

- 核心包/API：`testing`、`httptest`、`gtest`、表格测试、Service Mock、数据库集成测试。
- 目标：覆盖商品校验、库存事务、Session 和 JWT。
- 练习：补充关键业务测试。

### 22. 上线前收尾与综合题

- 核心内容：健康检查、结构化日志、链路信息、优雅关闭、配置密钥、`gf build`、Docker Compose。
- 目标：完成可 Docker 化运行的商城 API。
- 综合题：在不参考完整答案的情况下完成商城 API 并通过最终验收。

## 项目接口与数据模型

项目包含：

- 用户、分类、商品、订单、订单明细五张核心表。
- 金额使用整数“分”，避免浮点精度问题。
- 商品 CRUD、筛选和分页。
- 创建、查询、取消订单及库存事务。
- Session 与 JWT 两套登录入口。
- Redis 商品缓存、会话存储、令牌撤销和幂等控制。
- 统一响应：`code`、`message`、`data`、`traceId`。
- OpenAPI/Swagger、健康检查及 Docker 化运行。

## 最终验收

- `go test ./...`、`go vet ./...` 通过。
- 非法参数得到稳定的错误码和响应结构。
- 并发下不会超卖，事务失败能够完整回滚。
- 商品更新后缓存正确失效。
- Session、JWT、刷新和退出逻辑均可验证。
- 重复创建订单或支付回调不会重复扣库存。
- MySQL、Redis 或支付服务异常时有明确日志和可控响应。
- Docker Compose 能启动应用、MySQL 和 Redis。

## 默认边界

第一阶段不加入前端页面、模板引擎、gRPC、服务注册、消息队列和 Kubernetes。

这些内容在商城 API 完成后，可以按需要进入进阶阶段。
