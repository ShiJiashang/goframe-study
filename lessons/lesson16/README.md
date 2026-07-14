# 第 16 课：Redis 与 Cache Aside

## 本课目标

这一课给商品详情增加 Redis 缓存，并理解：

- `g.Redis()` 与 `*gredis.Redis`。
- `Get`、`SetEX`、`Del`、`TTL`。
- `gcache.Cache` 的本地缓存能力。
- Cache Aside：先查缓存，未命中查数据库并回填。
- 商品更新后为什么必须让缓存失效。

## 和 Gin 的对应关系

Gin 不提供 Redis 客户端，通常自行创建：

```go
rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
value, err := rdb.Get(c.Request.Context(), key).Result()
```

GoFrame 统一从配置获得 Redis 实例：

```go
value, err := g.Redis().Get(ctx, key)
```

`ctx` 仍然是当前接口的上下文，用于取消、超时和链路日志。

## 启动 Redis 7

```bash
docker run --name goframe-redis \
  -p 6379:6379 \
  -d redis:7-alpine
```

验证：

```bash
docker exec goframe-redis redis-cli PING
```

预期：

```text
PONG
```

## 安装并注册 Redis 适配器

```bash
cd mall
go get github.com/gogf/gf/contrib/nosql/redis/v2@v2.10.2
```

在 `internal/cmd/cmd.go` 中加入空白导入：

```go
import (
    _ "github.com/gogf/gf/contrib/nosql/redis/v2"
)
```

它和 MySQL 驱动一样，通过 `init()` 向 GoFrame 注册实现。

## Redis 配置

在 `manifest/config/config.yaml` 增加：

```yaml
redis:
  default:
    address: "127.0.0.1:6379"
    db: 0
    idleTimeout: "60s"
    maxActive: 20
```

- `default`：默认 Redis 分组。
- `address`：Redis 地址。
- `db`：逻辑数据库编号。
- `idleTimeout`：空闲连接超时。
- `maxActive`：最大活跃连接数。

生产密码不能硬编码进仓库，应由环境变量或密钥管理系统提供。

## `g.Redis()`

函数签名：

```go
func Redis(name ...string) *gredis.Redis
```

```go
redis := g.Redis()
```

- `name ...string`：可选配置分组名；不传使用 `default`。
- 返回 `*gredis.Redis`：GoFrame Redis 客户端。
- `redis`：当前客户端变量，不是一条长期独占的 TCP 连接。

## 常用 Redis 方法

### `Get`

```go
func (r *Redis) Get(
    ctx context.Context,
    key string,
) (*gvar.Var, error)
```

```go
value, err := g.Redis().Get(ctx, "product:detail:1")
```

- 返回 `*gvar.Var`，可以调用 `String()`、`Bytes()`、`IsNil()`。
- key 不存在时不是普通业务错误，应按缓存未命中处理。
- Redis 连接异常时 `err != nil`。

### `SetEX`

```go
SetEX(
    ctx context.Context,
    key string,
    value any,
    ttlInSeconds int64,
) error
```

```go
err := g.Redis().SetEX(ctx, key, jsonText, 600)
```

- `value`：缓存内容。
- `ttlInSeconds`：过期时间，单位秒。
- 这里的 `600` 表示十分钟。

商品缓存必须设置 TTL，避免旧数据永久存在。

### `Del`

```go
func (r *Redis) Del(
    ctx context.Context,
    keys ...string,
) (int64, error)
```

```go
deleted, err := g.Redis().Del(ctx, key)
```

- `keys ...string`：可以删除一个或多个 key。
- `deleted`：实际删除数量；key 不存在时通常是 0。

### `TTL`

```go
func (r *Redis) TTL(
    ctx context.Context,
    key string,
) (int64, error)
```

返回剩余秒数。Redis 常见特殊值：

```text
-1  key存在但没有过期时间
-2  key不存在
```

## `gcache.Cache`

`gcache` 默认是当前进程内存缓存：

```go
cache := gcache.New()

err := cache.Set(ctx, "key", "value", time.Minute)
value, err := cache.Get(ctx, "key")
removed, err := cache.Remove(ctx, "key")
```

核心签名：

```go
func New(lruCap ...int) *gcache.Cache

func (c *Cache) Set(
    ctx context.Context,
    key any,
    value any,
    duration time.Duration,
) error

func (c *Cache) Get(
    ctx context.Context,
    key any,
) (*gvar.Var, error)

func (c *Cache) Remove(
    ctx context.Context,
    keys ...any,
) (*gvar.Var, error)
```

本地缓存和 Redis 的区别：

| 项目 | `gcache.New()` | `g.Redis()` |
| --- | --- | --- |
| 保存位置 | 当前 Go 进程内存 | Redis 服务 |
| 多实例共享 | 否 | 是 |
| 程序重启 | 丢失 | 通常保留 |
| 网络开销 | 无 | 有 |

商城商品详情使用 Redis；很短暂、只在单进程使用的数据可以考虑 `gcache`。

## Cache Aside 流程

读取：

```text
查询缓存
├── 命中：返回缓存
└── 未命中：查询数据库 → 写缓存 → 返回
```

更新：

```text
更新数据库 → 删除缓存
```

不是“更新数据库后再把新值写缓存”。删除缓存更简单，下一次读取会根据数据库重新构建。

## 可运行样例：缓存商品详情

在商品 Logic 中增加：

```go
func productDetailCacheKey(id int64) string {
    return fmt.Sprintf("mall:product:detail:%d", id)
}

func (s *sProduct) Detail(
    ctx context.Context,
    in model.ProductDetailInput,
) (out *model.ProductDetailOutput, err error) {
    key := productDetailCacheKey(in.ID)

    cached, cacheErr := g.Redis().Get(ctx, key)
    if cacheErr != nil {
        g.Log().Warningf(ctx, "读取商品缓存失败 key=%s err=%v", key, cacheErr)
    } else if cached != nil && !cached.IsNil() && cached.String() != "" {
        out = new(model.ProductDetailOutput)
        if decodeErr := gjson.DecodeTo(cached.Bytes(), out); decodeErr == nil {
            return out, nil
        }

        // 缓存内容损坏时删除，继续查数据库。
        _, _ = g.Redis().Del(ctx, key)
    }

    var product entity.Products
    err = dao.Products.Ctx(ctx).
        Where(dao.Products.Columns().Id, in.ID).
        Scan(&product)
    if err != nil {
        return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "查询商品失败")
    }
    if product.Id == 0 {
        return nil, gerror.NewCode(consts.CodeProductNotFound)
    }

    out = &model.ProductDetailOutput{
        ID:         int64(product.Id),
        CategoryID: int64(product.CategoryId),
        Name:       product.Name,
        PriceCent:  int64(product.PriceCent),
        Stock:      int(product.Stock),
        Status:     int(product.Status),
    }

    encoded, encodeErr := gjson.Encode(out)
    if encodeErr != nil {
        return nil, gerror.Wrap(encodeErr, "编码商品缓存失败")
    }
    if cacheErr = g.Redis().SetEX(ctx, key, encoded, 600); cacheErr != nil {
        // 缓存故障时商品详情仍可从数据库返回，但必须记录日志。
        g.Log().Warningf(ctx, "写入商品缓存失败 key=%s err=%v", key, cacheErr)
    }

    return out, nil
}
```

关键变量：

- `key`：商品缓存键，带统一前缀。
- `cached`：Redis 返回的 `*gvar.Var`。
- `cacheErr`：缓存错误；本例选择降级查数据库。
- `product`：数据库 Entity。
- `encoded`：序列化后的 JSON 字节。

不要缓存数据库错误。商品不存在是否做短 TTL 的空值缓存属于后续优化，本课先不做。

## 更新后删除缓存

商品更新成功后：

```go
key := productDetailCacheKey(in.ID)

if _, err = g.Redis().Del(ctx, key); err != nil {
    return nil, gerror.WrapCode(
        gcode.CodeOperationFailed,
        err,
        "商品已更新但缓存失效失败",
    )
}
```

必须先确认数据库更新成功，再删除缓存。

## 运行和观察

```bash
gf run main.go
```

连续请求两次商品详情，然后查看 Redis：

```bash
docker exec goframe-redis redis-cli \
  GET 'mall:product:detail:1'

docker exec goframe-redis redis-cli \
  TTL 'mall:product:detail:1'
```

第一次应执行数据库查询并写缓存；第二次应直接命中缓存。

## 本课练习：更新商品后缓存失效

照着做：

1. 请求商品详情，确认 Redis 中出现 key。
2. 调用商品更新接口修改价格。
3. 在商品 Update Logic 成功后调用 `Del`。
4. 确认 key 被删除。
5. 再次请求详情，确认返回新价格并重新写入缓存。
6. 停止 Redis 后请求详情，确认仍能从 MySQL 返回，并产生 Warning 日志。

## 验收条件

- 商品详情第一次查数据库，第二次命中 Redis。
- key 格式统一为 `mall:product:detail:{id}`。
- TTL 大于 0，不存在永不过期的商品详情缓存。
- 商品更新后缓存被删除，再次读取是新数据。
- 损坏缓存会被删除并从数据库恢复。
- Redis 异常时有明确日志，读取接口能够按设计降级。
- `go test ./...` 和 `go vet ./...` 通过。

完成后提交两次请求日志、Redis 的 GET/TTL 结果和 Update 的失效代码。
