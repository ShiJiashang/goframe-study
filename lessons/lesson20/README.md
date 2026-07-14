# 第 20 课：定时任务、订单过期与幂等

## 本课目标

这一课学两个商城里特别容易重复执行的业务：

- 定时取消超时未支付订单，并归还库存。
- 使用 `Idempotency-Key`，避免重复创建订单和重复扣库存。

本课代码已经写进工作区：

```text
lessons/lesson20/orderapp/main.go
```

为了先把流程看懂，本课样例用内存 `map` 模拟商品、订单和幂等记录。真实项目里要换成 MySQL 事务、Redis `SET NX EX` 和数据库唯一约束。

## 和 Gin 的对应关系

Gin 主要处理 HTTP 请求。定时任务一般额外引入 cron 库，或者单独启动 worker。

GoFrame 自带 `gcron`：

```go
gcron.AddSingleton(ctx, "*/2 * * * * *", job, "cancel-expired-orders")
```

这件事和 Gin 里的 handler 不是一类东西：

```text
HTTP handler：
    用户请求来了才执行

gcron 定时任务：
    没有用户请求，也会按时间执行
```

所以你刚才问“取消订单不是需要订单 id 吗”，答案是：

```text
手动取消某个订单：需要 orderId
定时取消过期订单：先扫描数据库找出过期 orderId，再逐个取消
```

定时任务入口只有 `ctx` 是正常的，因为订单 ID 是它自己从订单表查出来的。

## 运行本课代码

启动服务：

```bash
go run lessons/lesson20/orderapp/main.go
```

服务地址：

```text
http://127.0.0.1:8010
```

查看商品库存：

```bash
curl http://127.0.0.1:8010/api/products/1
```

创建订单：

```bash
curl -X POST http://127.0.0.1:8010/api/orders \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: create-order-001' \
  -d '{"productId":1,"quantity":2,"expireSeconds":6}'
```

再次用同一个幂等键请求：

```bash
curl -X POST http://127.0.0.1:8010/api/orders \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: create-order-001' \
  -d '{"productId":1,"quantity":2,"expireSeconds":6}'
```

第二次响应里的 `repeated` 会是 `true`，并且不会再扣库存。

等 6 秒左右，再查商品库存：

```bash
curl http://127.0.0.1:8010/api/products/1
```

后台定时任务每 2 秒扫描一次，订单过期后会自动取消，并把库存加回来。

如果你想手动触发扫描：

```bash
curl -X POST http://127.0.0.1:8010/api/jobs/cancel-expired
```

如果你不想订单被自动取消，可以在过期前支付：

```bash
curl -X POST http://127.0.0.1:8010/api/orders/1001/pay
```

## 路由说明

本课 `main` 注册了 5 个接口：

```go
server.Group("/api", func(group *ghttp.RouterGroup) {
    group.GET("/products/:id", handleGetProduct(store))
    group.GET("/orders/:id", handleGetOrder(store))
    group.POST("/orders", handleCreateOrder(store))
    group.POST("/orders/:id/pay", handlePayOrder(store))
    group.POST("/jobs/cancel-expired", handleCancelExpired(store))
})
```

- `GET /api/products/:id`：查看商品库存。
- `GET /api/orders/:id`：查看订单。
- `POST /api/orders`：创建订单，会扣库存。
- `POST /api/orders/:id/pay`：把订单改成已支付。
- `POST /api/jobs/cancel-expired`：手动触发取消过期订单。

## 核心类型

### `OrderStatus`

```go
type OrderStatus string

const (
    StatusPending   OrderStatus = "pending"
    StatusPaid      OrderStatus = "paid"
    StatusCancelled OrderStatus = "cancelled"
)
```

- `OrderStatus`：订单状态类型，底层是 `string`。
- `pending`：待支付。
- `paid`：已支付。
- `cancelled`：已取消。

为什么不直接到处写字符串？

因为字符串写散了容易拼错，比如 `canceled`、`cancelled` 混用。常量能统一状态值。

### `Product`

```go
type Product struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Stock int    `json:"stock"`
}
```

- `ID`：商品 ID。
- `Name`：商品名。
- `Stock`：库存。
- `json` 标签：控制返回 JSON 的字段名。

### `Order`

```go
type Order struct {
    ID             int64       `json:"id"`
    ProductID      int64       `json:"productId"`
    Quantity       int         `json:"quantity"`
    Status         OrderStatus `json:"status"`
    IdempotencyKey string      `json:"idempotencyKey"`
    ExpireAt       string      `json:"expireAt"`
}
```

- `ID`：订单 ID。
- `ProductID`：买的是哪个商品。
- `Quantity`：购买数量。
- `Status`：订单状态。
- `IdempotencyKey`：创建订单时客户端传来的幂等键。
- `ExpireAt`：过期时间。超过这个时间还没支付，定时任务会取消。

### `CreateOrderInput`

```go
type CreateOrderInput struct {
    ProductID      int64
    Quantity       int
    IdempotencyKey string
    ExpireAfter    time.Duration
}
```

这是 service 层输入模型，不是 HTTP 请求结构。

- `ProductID`：商品 ID。
- `Quantity`：数量。
- `IdempotencyKey`：幂等键。
- `ExpireAfter`：多久后过期。

### `Store`

```go
type Store struct {
    mu          sync.Mutex
    nextOrderID int64
    products    map[int64]*Product
    orders      map[int64]*Order
    idem        map[string]idempotencyRecord
}
```

这个 `Store` 模拟数据库和 Redis。

- `mu`：互斥锁，保护下面几个 map。
- `nextOrderID`：模拟自增订单 ID。
- `products`：模拟商品表。
- `orders`：模拟订单表。
- `idem`：模拟 Redis 幂等键。

真实项目对应关系：

```text
products map -> products 表
orders map   -> orders 表
idem map     -> Redis + orders.idempotency_key 唯一索引
mu           -> 数据库事务/行锁/Redis 原子命令
```

## `StartOrderJobs`

```go
func StartOrderJobs(ctx context.Context, store *Store) error
```

- `ctx`：注册定时任务时的上下文。
- `store`：本课的内存数据仓库。
- 返回 `error`：注册任务失败时返回。

核心代码：

```go
entry, err := gcron.AddSingleton(
    ctx,
    "*/2 * * * * *",
    func(jobCtx context.Context) {
        cancelled, err := store.CancelExpired(jobCtx)
        if err != nil {
            g.Log().Error(jobCtx, "cancel expired orders failed", err)
            return
        }
        if cancelled > 0 {
            g.Log().Info(jobCtx, "expired orders cancelled", "count", cancelled)
        }
    },
    "cancel-expired-orders",
)
```

### `gcron.AddSingleton`

```go
gcron.AddSingleton(
    ctx context.Context,
    pattern string,
    job gcron.JobFunc,
    name ...string,
) (*gcron.Entry, error)
```

- `ctx`：注册任务的上下文。
- `pattern`：cron 表达式。
- `job`：真正执行的函数，形式是 `func(ctx context.Context)`。
- `name`：任务名。
- 返回 `*gcron.Entry` 和 `error`。

`AddSingleton` 的作用是：

```text
如果上一次任务还没跑完，下一次到了时间也不重叠执行。
```

但它只管当前 Go 进程。你部署 3 个实例时，3 个实例都会各自跑任务。多实例要用 Redis 分布式锁、数据库抢占，或者只启动一个 worker。

### Cron 表达式

GoFrame `gcron` 常用六段格式：

```text
秒 分 时 日 月 星期
```

本课用：

```text
*/2 * * * * *
```

意思是每 2 秒执行一次，方便你观察。真实项目可以改成每分钟：

```text
0 */1 * * * *
```

## 创建订单与幂等

创建订单入口：

```go
func (s *Store) CreateOrder(ctx context.Context, input CreateOrderInput) (*Order, bool, error)
```

返回值：

- `*Order`：创建出的订单，或者重复请求时返回第一次创建的订单。
- `bool`：是否是重复请求。`true` 表示同一个 `Idempotency-Key` 已经创建过订单。
- `error`：错误。

核心流程：

```text
检查 Idempotency-Key
    ↓
检查商品 ID、购买数量
    ↓
加锁
    ↓
检查 idem map 里是否已有这个 key
    ↓
有：返回原订单，不再扣库存
    ↓
没有：检查库存、扣库存、创建订单、保存幂等记录
    ↓
解锁
```

关键代码：

```go
if record, ok := s.idem[input.IdempotencyKey]; ok && record.Expires.After(now) {
    order, ok := s.orders[record.OrderID]
    if !ok {
        return nil, false, gerror.New("idempotency record points to missing order")
    }
    return cloneOrder(order), true, nil
}
```

意思是：

```text
这个幂等键以前成功创建过订单
    -> 找到原订单
    -> repeated 返回 true
    -> 不再扣库存
```

真正创建时：

```go
product.Stock -= input.Quantity
s.orders[order.ID] = order
s.idem[input.IdempotencyKey] = idempotencyRecord{
    OrderID:  order.ID,
    Expires: now.Add(10 * time.Minute),
}
```

- `product.Stock -= input.Quantity`：扣库存。
- `s.orders[order.ID] = order`：保存订单。
- `s.idem[...] = ...`：保存幂等键和订单 ID 的关系。

真实项目里，这里不能只靠内存 map。应该是：

```text
Redis SET key value NX EX
    ↓
MySQL 事务扣库存、创建订单
    ↓
orders.idempotency_key 加唯一索引
    ↓
成功后 Redis key 写入 orderId
```

## 取消过期订单

定时任务调用：

```go
cancelled, err := store.CancelExpired(jobCtx)
```

函数签名：

```go
func (s *Store) CancelExpired(ctx context.Context) (int, error)
```

- `ctx`：定时任务上下文。
- 返回 `int`：本次取消了多少订单。
- 返回 `error`：扫描过程是否失败。

为什么这里没有 `orderId`？

因为它不是取消某一笔订单，而是扫描所有过期订单：

```go
for _, order := range s.orders {
    if order.Status == StatusPending && !expireAt.After(now) {
        expiredIDs = append(expiredIDs, order.ID)
    }
}
```

然后逐个取消：

```go
for _, orderID := range expiredIDs {
    ok, err := s.cancelOne(ctx, orderID)
    ...
}
```

单个取消函数：

```go
func (s *Store) cancelOne(ctx context.Context, orderID int64) (bool, error)
```

- `orderID`：这时就需要订单 ID 了。
- `bool`：是否真的取消成功。
- `error`：取消过程是否失败。

核心保护：

```go
if order.Status != StatusPending || expireAt.After(time.Now()) {
    return false, nil
}
```

这句很重要。因为扫描到订单之后，用户可能刚好支付了。取消前必须再次确认：

```text
还是 pending
并且确实已经过期
```

真正取消：

```go
order.Status = StatusCancelled
product.Stock += order.Quantity
```

意思是：

```text
订单改成 cancelled
库存归还
```

真实项目里，这一段要放进数据库事务，并对订单行 `LockUpdate()`。

## HTTP 创建订单接口

```go
func handleCreateOrder(store *Store) func(r *ghttp.Request)
```

这是一个“返回 handler 的函数”。

- `store *Store`：把内存仓库传进去。
- 返回值 `func(r *ghttp.Request)`：真正注册给 GoFrame 的 HTTP 处理函数。

核心代码：

```go
order, repeated, err := store.CreateOrder(r.Context(), CreateOrderInput{
    ProductID:      r.Get("productId", 1).Int64(),
    Quantity:       r.Get("quantity", 1).Int(),
    IdempotencyKey: r.Header.Get("Idempotency-Key"),
    ExpireAfter:    time.Duration(expireSeconds) * time.Second,
})
```

- `r.Context()`：当前 HTTP 请求上下文。
- `r.Get("productId", 1)`：从 body/query/form 取参数，默认值是 `1`。
- `r.Get("quantity", 1)`：购买数量，默认值是 `1`。
- `r.Header.Get("Idempotency-Key")`：从请求头读取幂等键。
- `ExpireAfter`：把秒数转换成 `time.Duration`。

## Redis `SET NX EX` 在真实项目里怎么换

本课内存代码：

```go
if record, ok := s.idem[input.IdempotencyKey]; ok && record.Expires.After(now) {
    return cloneOrder(order), true, nil
}
```

真实项目应换成 Redis：

```go
ttlSeconds := int64(600)
result, err := g.Redis().Set(
    ctx,
    "mall:idempotency:create-order:"+idempotencyKey,
    "processing",
    gredis.SetOption{
        TTLOption: gredis.TTLOption{EX: &ttlSeconds},
        NX:        true,
    },
)
if err != nil {
    return nil, gerror.Wrap(err, "占用幂等键失败")
}
if result == nil || result.IsNil() {
    return nil, gerror.New("该请求正在处理或已经处理")
}
```

- `NX: true`：只有键不存在时才写入。
- `EX`：设置过期时间，避免程序崩溃后 key 永远不释放。
- `SET NX EX` 是一条 Redis 命令，判断和设置 TTL 是原子的。

不要拆成：

```go
SetNX(...)
Expire(...)
```

因为程序可能在两条命令中间崩溃，导致 key 没有 TTL。

## 真实 MySQL 版取消订单应该怎么写

本课内存代码是：

```go
order.Status = StatusCancelled
product.Stock += order.Quantity
```

真实项目必须用事务：

```text
开始事务
    按订单 ID 查询并 LockUpdate 锁住订单行
    再次确认 status=pending 且 expire_at <= now
    查询订单明细
    逐项归还库存
    把订单改成 cancelled
提交事务
```

原因：

```text
定时任务取消订单
用户支付订单
用户手动取消订单
```

这些可能同时发生。事务和行锁保证同一笔订单不会被两条流程同时改坏。

## 本课练习

照着当前代码做一个小改造：新增“手动取消订单”接口。

接口：

```text
POST /api/orders/:id/cancel
```

要求：

1. 在路由里注册 `/orders/:id/cancel`。
2. 给 `Store` 增加 `CancelOrder(ctx, orderID)` 方法。
3. 如果订单是 `pending`，改成 `cancelled` 并归还库存。
4. 如果订单是 `paid`，返回错误：已支付订单不能取消。
5. 如果订单已经是 `cancelled`，直接返回当前订单，不要重复归还库存。
6. handler 里返回统一 JSON。

你可以照着 `handlePayOrder` 和 `cancelOne` 写。

## 验收条件

- 同一个 `Idempotency-Key` 连续请求两次，只创建一张订单。
- 第二次响应 `repeated=true`。
- 库存只扣一次。
- 订单过期后自动变成 `cancelled`。
- 过期取消后库存准确归还一次。
- 已支付订单不会被定时任务取消。
- `go test ./lessons/lesson20/...` 通过。
- `go vet ./lessons/lesson20/...` 通过。

完成后把你新增的 `CancelOrder` 和 handler 发我，我先评审，再给参考答案。
