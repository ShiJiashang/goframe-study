# 第 11 课：连接 MySQL 与认识 ORM

## 本课目标

这一课只完成三件事：

- 让 GoFrame 加载 MySQL 驱动并连接 `goframe_mall`。
- 分清 `g.DB()`、`gdb.DB`、`gdb.Model` 和 `Ctx(ctx)`。
- 实现 `GET /health/database` 数据库健康检查。

本课不做商品 CRUD，CRUD 放在第 12 课。

## 和 Gin 的对应关系

Gin 只负责 HTTP，不自带 ORM。Gin 项目通常自己创建 `*sql.DB` 或 GORM：

```go
db, err := sql.Open("mysql", dsn)
if err != nil {
    panic(err)
}

r.GET("/health/database", func(c *gin.Context) {
    err := db.PingContext(c.Request.Context())
    // 输出响应
})
```

GoFrame 把数据库配置、连接池和 ORM 入口统一起来：

```go
db := g.DB()
err := db.Ctx(ctx).PingMaster()
```

注意：`g.DB()` 并不是每次请求都新建一条连接。它返回可复用的数据库对象，底层连接由连接池管理。

## 第一步：启动 MySQL 8

如果你已有本地 MySQL 8，可以直接使用。否则可以临时启动容器：

```bash
docker run --name goframe-mysql \
  -e MYSQL_ROOT_PASSWORD=12345678 \
  -e MYSQL_DATABASE=goframe_mall \
  -p 3306:3306 \
  -d mysql:8.0
```

查看是否就绪：

```bash
docker logs -f goframe-mysql
```

看到 `ready for connections` 后按 `Ctrl+C`，只会退出日志查看，不会停止 MySQL。

导入本课表结构。在仓库根目录执行：

```bash
docker exec -i goframe-mysql \
  mysql -uroot -p12345678 \
  < lessons/lesson11/schema.sql
```

验证表：

```bash
docker exec goframe-mysql \
  mysql -uroot -p12345678 -e \
  'USE goframe_mall; SHOW TABLES; SELECT * FROM products;'
```

本课 SQL 使用 `InnoDB`。第 15 课的事务和行锁依赖 InnoDB。

## 第二步：安装并注册 MySQL 驱动

进入商城目录：

```bash
cd mall
go get github.com/gogf/gf/contrib/drivers/mysql/v2@v2.10.2
```

GoFrame 的 ORM 接口和具体数据库驱动是分开的：

```text
github.com/gogf/gf/v2/database/gdb       ORM 接口
github.com/gogf/gf/contrib/drivers/mysql/v2  MySQL 实现
```

在 `mall/internal/cmd/cmd.go` 的 import 中加入：

```go
import (
    _ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)
```

前面的 `_` 是空白导入：我们不直接调用这个包的函数，只需要执行它的 `init()`，让 MySQL 驱动注册到 GoFrame。

如果漏掉它，程序可能报“找不到 mysql 数据库驱动”。

## 第三步：配置数据库

修改 `mall/manifest/config/config.yaml`：

```yaml
database:
  default:
    link: "mysql:root:12345678@tcp(127.0.0.1:3306)/goframe_mall?parseTime=true&loc=Local"
    debug: true
    maxIdle: 10
    maxOpen: 20
    maxLifeTime: "30m"
```

字段作用：

- `database`：GoFrame 数据库配置根节点。
- `default`：数据库分组名；不传分组调用 `g.DB()` 时使用它。
- `link`：驱动、用户名、密码、地址和数据库名。
- `debug`：开发时记录 SQL 和耗时，生产环境通常关闭。
- `maxIdle`：连接池最多保留多少空闲连接。
- `maxOpen`：连接池最多同时打开多少连接。
- `maxLifeTime`：一条连接最多复用多久。

不要把生产密码提交到 GitHub。当前密码只用于本地学习。

## `g.DB()` 与 `gdb.DB`

导入：

```go
import "github.com/gogf/gf/v2/frame/g"
```

函数签名：

```go
func DB(name ...string) gdb.DB
```

调用：

```go
db := g.DB()
```

- `name ...string`：可选数据库分组名；不传时使用 `default`。
- 返回 `gdb.DB`：GoFrame ORM 的数据库接口。
- 变量 `db` 的静态类型是 `gdb.DB`，不是某个 MySQL 私有类型。

如果配置了另一个分组 `analytics`：

```go
analyticsDB := g.DB("analytics")
```

`gdb.DB` 是接口，常用方法包括：

```go
type DB interface {
    Ctx(ctx context.Context) DB
    Model(tableNameOrStruct ...any) *Model
    PingMaster() error
    Transaction(ctx context.Context, f func(context.Context, TX) error) error
}
```

使用接口的好处是业务代码不绑定 MySQL 驱动的具体结构体。

## `Ctx(ctx)`

```go
dbWithCtx := g.DB().Ctx(ctx)
```

签名：

```go
Ctx(ctx context.Context) gdb.DB
```

- 参数 `ctx`：当前 HTTP 请求的上下文。
- 返回新的 `gdb.DB`：绑定了当前上下文的浅拷贝。

它让数据库操作能够获得：

- 请求取消信号。
- 超时控制。
- TraceID 和链路日志信息。

建议每次请求都带上 `ctx`：

```go
g.DB().Ctx(ctx).Model("products")
```

## `gdb.Model`

创建模型：

```go
model := g.DB().Model("products").Ctx(ctx)
```

`Model` 签名：

```go
Model(tableNameOrStruct ...any) *gdb.Model
```

- 参数：表名、带别名的表名或结构体。
- 返回 `*gdb.Model`：SQL 构建对象。

`Ctx` 也可以在 Model 上调用：

```go
func (m *Model) Ctx(ctx context.Context) *Model
```

创建 Model 不会立即查询数据库：

```go
model := g.DB().Model("products").Ctx(ctx) // 还没有执行 SQL
count, err := model.Count()                // 这里才执行 SQL
```

## `PingMaster()`

```go
err := g.DB().Ctx(ctx).PingMaster()
```

签名：

```go
PingMaster() error
```

- 成功返回 `nil`。
- 连接失败、密码错误或数据库不可用时返回错误。
- `Master` 指写库；我们目前只有一个数据库节点。

## 可运行样例：数据库健康检查

新增 `mall/api/health/v1/health.go`：

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

type DatabaseReq struct {
    g.Meta `path:"/health/database" method:"get" tags:"Health" summary:"数据库健康检查"`
}

type DatabaseRes struct {
    Status       string `json:"status" dc:"数据库状态"`
    ProductCount int    `json:"productCount" dc:"商品数量"`
}
```

执行：

```bash
gf gen ctrl
```

在生成的 `mall/internal/controller/health/health_v1_database.go` 中实现：

```go
package health

import (
    "context"

    "github.com/gogf/gf/v2/errors/gcode"
    "github.com/gogf/gf/v2/errors/gerror"
    "github.com/gogf/gf/v2/frame/g"

    "goframe-study/mall/api/health/v1"
)

func (c *ControllerV1) Database(
    ctx context.Context,
    req *v1.DatabaseReq,
) (res *v1.DatabaseRes, err error) {
    db := g.DB().Ctx(ctx)

    if err = db.PingMaster(); err != nil {
        return nil, gerror.WrapCode(
            gcode.CodeDbOperationError,
            err,
            "数据库连接失败",
        )
    }

    productModel := db.Model("products")
    productCount, err := productModel.Count()
    if err != nil {
        return nil, gerror.WrapCode(
            gcode.CodeDbOperationError,
            err,
            "统计商品数量失败",
        )
    }

    return &v1.DatabaseRes{
        Status:       "ok",
        ProductCount: productCount,
    }, nil
}
```

关键变量：

- `db`：绑定请求 `ctx` 的 `gdb.DB` 接口值。
- `productModel`：指向 `products` 表的 `*gdb.Model`。
- `productCount`：`Count()` 返回的商品数量 `int`。
- `err`：连接或 SQL 执行错误。

最后在 `mall/internal/cmd/cmd.go` 中导入并绑定：

```go
import "goframe-study/mall/internal/controller/health"

group.Bind(
    health.NewV1(),
)
```

启动并请求：

```bash
gf run main.go
curl -s 'http://127.0.0.1:8000/health/database'
```

预期：

```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "status": "ok",
    "productCount": 2
  }
}
```

## 本课练习：补充数据库名称和耗时

在 `DatabaseRes` 增加：

```go
Database string `json:"database"`
CostMs   int64  `json:"costMs"`
```

照着做：

1. Controller 开头保存 `start := time.Now()`。
2. 响应中的 `Database` 固定填写 `goframe_mall`。
3. 使用 `time.Since(start).Milliseconds()` 计算 `CostMs`。
4. 暂停 MySQL 后再次请求，确认接口返回数据库错误而不是成功。

停止和重新启动容器：

```bash
docker stop goframe-mysql
docker start goframe-mysql
```

## 验收条件

- MySQL 正常时 `code=0`、`status=ok`、`productCount=2`。
- 响应包含 `database=goframe_mall` 和非负的 `costMs`。
- MySQL 停止时得到稳定的数据库错误响应。
- SQL 日志带有当前请求的 TraceID。
- `go test ./...` 和 `go vet ./...` 通过。

完成后提交代码和两次响应：MySQL 正常一次、停止一次。我先评审，再给参考答案。
