# 第 19 课：HTTP Client 与外部支付

## 本课目标

这一课让商城调用一个本地 Mock 支付服务，学习：

- `gclient.New()`、`g.Client()` 和 `gclient.Client`。
- GET/POST 请求、JSON 请求体、请求上下文和超时。
- 如何读取状态码与响应体，以及为什么要关闭响应。
- 第三方服务超时、返回非 2xx、返回错误 JSON 时怎样处理。
- 支付回调为什么必须幂等。

真实支付还涉及签名、证书和平台验签，本课暂不接入任何真实支付渠道。

## 这一课到底在做什么

前面几课我们大多是在写“别人请求我们的商城 API”。第 19 课反过来：商城后端也会主动请求别人。

链路是：

```text
浏览器/curl
    ↓
商城 API：POST /api/orders/:id/pay
    ↓
paymentclient.Client.Pay
    ↓
Mock 支付服务：POST http://127.0.0.1:9001/pay
    ↓
商城 API 解析支付结果并返回统一 JSON
```

这节课要你看懂三件事：

1. `ghttp.Request`：别人请求商城时，商城收到的请求。
2. `gclient.Client`：商城请求支付服务时，商城主动发出的 HTTP 客户端。
3. `gclient.Response`：支付服务返回给商城的响应，不是商城最终返回给前端的响应。

## 和 Gin 的对应关系

Gin 是服务端 Web 框架，向外发送 HTTP 请求通常仍使用 `net/http.Client`。GoFrame 提供 `gclient.Client`，把常用的请求参数、JSON、前缀、Header 和调试能力封装起来。

请求进入商城：

```go
func (c *ControllerV1) Pay(ctx context.Context, req *PayReq) (*PayRes, error)
```

商城再作为客户端调用支付服务：

```go
response, err := gclient.New().ContentJson().Post(ctx, paymentURL, req)
```

这两个方向不要混淆：`ghttp.Request` 表示别人请求商城，`gclient.Response` 表示商城请求别人后收到的响应。

## 本地 Mock 支付服务

代码在：

```text
lessons/lesson19/mock-payment/main.go
```

打开第一个终端：

```bash
go run lessons/lesson19/mock-payment/main.go
```

服务地址是 `http://127.0.0.1:9001`。先直接测试：

```bash
curl -X POST http://127.0.0.1:9001/pay \
  -H 'Content-Type: application/json' \
  -d '{"orderNo":"O202607130001","amountCent":9900}'
```

预期响应：

```json
{"tradeNo":"MOCK-O202607130001","status":"paid"}
```

两个特殊订单号用于测试异常：

- `TIMEOUT`：服务等待 3 秒才响应。
- `FAIL`：服务返回 HTTP 502。

## 可运行商城 API Demo

我已经把商城侧样例写进工作区：

```text
lessons/lesson19/paymentclient/client.go
lessons/lesson19/mall-api/main.go
```

打开第二个终端：

```bash
go run lessons/lesson19/mall-api/main.go
```

商城 API 地址是：

```text
http://127.0.0.1:8009
```

先查一笔内存订单：

```bash
curl http://127.0.0.1:8009/api/orders/1
```

发起支付：

```bash
curl -X POST http://127.0.0.1:8009/api/orders/1/pay
```

预期响应大概是：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "order": {
      "id": "1",
      "orderNo": "O202607140001",
      "amountCent": 9900,
      "status": "paid",
      "tradeNo": "MOCK-O202607140001"
    },
    "payment": {
      "tradeNo": "MOCK-O202607140001",
      "status": "paid"
    }
  }
}
```

测试支付服务返回 502：

```bash
curl -X POST 'http://127.0.0.1:8009/api/orders/2/pay?mockOrderNo=FAIL'
```

测试支付服务超时：

```bash
curl -X POST 'http://127.0.0.1:8009/api/orders/2/pay?mockOrderNo=TIMEOUT'
```

`mock-payment` 会睡 3 秒，但商城客户端只等 1.5 秒，所以你会看到商城提前返回错误。

测试重复回调：

```bash
curl -X POST http://127.0.0.1:8009/api/payments/callback \
  -H 'Content-Type: application/json' \
  -d '{"tradeNo":"T10001","orderNo":"O202607140002"}'
```

再发一次同样请求：

```bash
curl -X POST http://127.0.0.1:8009/api/payments/callback \
  -H 'Content-Type: application/json' \
  -d '{"tradeNo":"T10001","orderNo":"O202607140002"}'
```

第二次会返回 `duplicate: true`，表示重复回调被忽略。

注意：这个 demo 为了让你直接跑，用的是内存 map。真实项目必须用数据库唯一键保证幂等，后面 SQL 会讲。

## `gclient.New()` 与 `g.Client()`

```go
gclient.New() *gclient.Client
g.Client() *gclient.Client
```

它们都会得到 GoFrame HTTP 客户端，类型是 `*gclient.Client`。

- `gclient.New()`：直接从 `net/gclient` 包创建一个新客户端。
- `g.Client()`：从 `frame/g` 包拿一个客户端入口，写起来更短。

本课实际代码用的是 `gclient.New()`，因为 `paymentclient/client.go` 本身就是一个独立客户端包，不需要额外引入 `frame/g`。

常用链式方法：

```go
ContentJson() *Client
Timeout(t time.Duration) *Client
Header(m map[string]string) *Client
Prefix(prefix string) *Client
```

- `ContentJson`：设置 JSON 请求头，并把结构体参数编码为 JSON。
- `Timeout`：设置整个请求的超时时间。
- `Header`：批量设置请求头。
- `Prefix`：给后续相对 URL 添加统一前缀。
- 返回值仍是 `*Client`，所以可以继续链式调用。
- 这些链式方法内部会基于当前 client 克隆一个新 client，再设置对应参数。

建议为支付服务建立一个长期复用的客户端。工作区代码在 `lessons/lesson19/paymentclient/client.go`，实际写法是：

```go
type Client struct {
	http *gclient.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 1500 * time.Millisecond
	}
	return &Client{
		http: gclient.New().
			ContentJson().
			Timeout(timeout).
			Prefix(strings.TrimRight(baseURL, "/")),
	}
}
```

- `Client` 是我们自己的类型，不是框架类型。
- `http` 保存可复用的 GoFrame HTTP 客户端。
- `baseURL`：支付服务基础地址。
- `timeout`：调用支付服务最多等待多久；传 `0` 或负数时默认 `1500ms`。
- `strings.TrimRight`：去掉末尾 `/`，避免 `Prefix` 里出现多余斜杠。
- `Prefix`：把基础地址绑定到客户端上，后面请求可以只写 `"/pay"`。
- 返回值 `*Client`：我们自己封装的支付客户端。

## `Post` 的签名

```go
Post(ctx context.Context, url string, data ...any) (*gclient.Response, error)
```

- `ctx`：当前请求上下文；上游取消请求时，下游调用也能被取消。
- `url`：目标地址。
- `data`：可选请求数据；使用 `ContentJson` 后可传结构体或 map。
- 返回 `*gclient.Response` 和网络层错误。

`Get` 的形式相同：

```go
Get(ctx context.Context, url string, data ...any) (*gclient.Response, error)
```

网络错误与 HTTP 错误不同：

- DNS 失败、连接拒绝、超时：`err != nil`。
- 支付服务成功返回 HTTP 502：通常 `err == nil`，要继续检查 `response.StatusCode`。

本课的真实调用是：

```go
response, err := c.http.Post(ctx, "/pay", input)
```

- `c.http`：类型是 `*gclient.Client`，已经通过 `Prefix(baseURL)` 绑定了支付服务地址。
- `ctx`：当前请求上下文，从商城 API 的 `r.Context()` 传进来。
- `"/pay"`：相对路径，最终会拼成 `http://127.0.0.1:9001/pay`。
- `input`：请求体，结构体字段通过 `json` 标签编码成 JSON。
- `response`：支付服务返回的 HTTP 响应。
- `err`：网络层或客户端层错误，比如连接失败、超时。

## `gclient.Response`

`gclient.Response` 内嵌标准库的 `*http.Response`，所以可以访问：

```go
response.StatusCode
response.Header
```

读取正文：

```go
body := response.ReadAll()
text := response.ReadAllString()
```

请求成功后要关闭响应：

```go
defer response.Close()
```

`defer` 会在当前函数返回前执行。未关闭响应可能导致连接不能正常复用并耗尽资源。

本课里有三个判断层次。

第一层：请求没正常完成。

```go
response, err := c.http.Post(ctx, "/pay", input)
if err != nil {
    return nil, gerror.Wrap(err, "request payment service failed")
}
```

这代表连接失败、支付服务没启动、超时等。此时可能根本没有可用响应体。

第二层：请求完成了，但 HTTP 状态码不是成功。

```go
if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
    return nil, gerror.Newf(...)
}
```

比如 Mock 支付服务返回 `502`，这时 `err` 通常是 `nil`，所以必须检查 `StatusCode`。

第三层：HTTP 成功了，但响应体不是我们想要的 JSON。

```go
if err = gjson.DecodeTo(body, &output); err != nil {
    return nil, gerror.Wrap(err, "decode payment response failed")
}
```

这代表支付服务虽然返回了 200，但内容格式不符合约定。

## 完整支付客户端样例

工作区文件：

```text
lessons/lesson19/paymentclient/client.go
```

核心代码：

```go
type Client struct {
	http *gclient.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 1500 * time.Millisecond
	}
	return &Client{
		http: gclient.New().
			ContentJson().
			Timeout(timeout).
			Prefix(strings.TrimRight(baseURL, "/")),
	}
}

func (c *Client) Pay(ctx context.Context, input PayInput) (*PayOutput, error) {
	response, err := c.http.Post(ctx, "/pay", input)
	if err != nil {
		return nil, gerror.Wrap(err, "request payment service failed")
	}
	defer response.Close()

	body := response.ReadAll()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, gerror.Newf(
			"payment service returned status=%d body=%s",
			response.StatusCode,
			string(body),
		)
	}

	var output PayOutput
	if err = gjson.DecodeTo(body, &output); err != nil {
		return nil, gerror.Wrap(err, "decode payment response failed")
	}
	if output.TradeNo == "" || output.Status == "" {
		return nil, gerror.New("payment response misses tradeNo or status")
	}
	return &output, nil
}
```

### 包与变量

- `context`：传递取消信号、超时、Trace 信息。
- `net/http`：这里使用标准 HTTP 状态常量。
- `gjson`：把响应 JSON 解码进结构体。
- `gerror`：为底层错误补充业务上下文。
- `gclient`：GoFrame HTTP 客户端包。
- `input`：本次支付请求数据。
- `response`：支付服务返回的完整 HTTP 响应。
- `body`：响应正文的 `[]byte`。
- `output`：解析后的业务响应，不包含 HTTP 状态码和响应头。

## Context 与 Timeout 的区别

客户端超时：

```go
gclient.New().Timeout(1500 * time.Millisecond)
```

规定这次 HTTP 调用最多等待多久。

Context 超时：

```go
requestCtx, cancel := context.WithTimeout(ctx, time.Second)
defer cancel()
response, err := client.Post(requestCtx, url, input)
```

它可以控制一整段调用链，而不只是 HTTP 客户端。谁先到期就由谁取消请求。Controller 传入的 `ctx` 本来就关联当前 HTTP 请求，不要无故替换成 `context.Background()`。

本课商城 API 里用的是：

```go
payResult, err := client.Pay(r.Context(), paymentclient.PayInput{
    OrderNo:    orderNo,
    AmountCent: amountCent,
})
```

- `r.Context()`：来自 `*ghttp.Request`。
- 它会把请求取消、trace、超时等上下文传给支付调用。
- 如果前端请求断开，下游调用也有机会被取消。

不要这样写：

```go
client.Pay(context.Background(), input)
```

因为这会断开当前请求链路，日志、trace、取消信号都没了。

## 商城侧 API 代码讲解

工作区文件：

```text
lessons/lesson19/mall-api/main.go
```

### `order`

```go
type order struct {
    ID         string `json:"id"`
    OrderNo    string `json:"orderNo"`
    AmountCent int64  `json:"amountCent"`
    Status     string `json:"status"`
    TradeNo    string `json:"tradeNo,omitempty"`
}
```

- `ID`：商城内部订单 ID，用于 `/api/orders/:id`。
- `OrderNo`：支付服务识别的订单号。
- `AmountCent`：金额，单位是分。
- `Status`：订单状态，示例里有 `pending`、`paid`。
- `TradeNo`：支付服务返回的交易号。
- `omitempty`：空字符串时 JSON 不输出这个字段。

### `apiResponse`

```go
type apiResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data"`
}
```

这是本课自己的统一响应结构：

- `Code`：业务码。
- `Message`：提示信息。
- `Data`：真正的数据。

### `orders`、`callbacks` 和 `mu`

```go
var (
    mu        sync.Mutex
    orders    = map[string]*order{}
    callbacks = map[string]bool{}
)
```

- `orders`：模拟订单表。
- `callbacks`：模拟已经处理过的支付回调。
- `mu`：互斥锁，避免并发读写 map 崩溃。

真实项目里这里会换成 MySQL 表和事务。

### `handlePayOrder`

```go
func handlePayOrder(r *ghttp.Request)
```

- 参数 `r`：当前请求，类型是 `*ghttp.Request`。
- 返回值为空：因为它直接用 `r.Response.WriteJson` 写响应。

核心步骤：

```text
取路径参数 id
    ↓
读取支付服务地址；没有配置时使用 http://127.0.0.1:9001
    ↓
从内存 orders 找订单
    ↓
调用 paymentclient.Client.Pay
    ↓
支付成功后更新订单状态
    ↓
返回统一 JSON
```

这一行是为了测试异常：

```go
if mockOrderNo := r.GetQuery("mockOrderNo").String(); mockOrderNo != "" {
    orderNo = mockOrderNo
}
```

所以你可以用：

```bash
curl -X POST 'http://127.0.0.1:8009/api/orders/2/pay?mockOrderNo=FAIL'
```

让商城拿 `FAIL` 去请求支付服务，从而模拟支付渠道 502。

支付服务地址来自：

```go
func getPaymentBaseURL(ctx context.Context) string {
    const defaultBaseURL = "http://127.0.0.1:9001"

    value, err := g.Cfg().Get(ctx, "payment.baseURL")
    if err != nil || value.String() == "" {
        return defaultBaseURL
    }
    return value.String()
}
```

- `defaultBaseURL`：默认 Mock 支付服务地址。
- `g.Cfg().Get(ctx, "payment.baseURL")`：尝试从配置读取。
- `err != nil`：没有配置文件或读取失败时，不让 demo 崩掉，直接用默认地址。
- `value.String() == ""`：配置值为空时也用默认地址。

### `handlePaymentCallback`

```go
func handlePaymentCallback(r *ghttp.Request)
```

它模拟第三方支付回调。

```go
tradeNo := r.Get("tradeNo").String()
orderNo := r.Get("orderNo").String()
```

- `r.Get`：从请求里取参数；这里可以取 JSON body 里的字段。
- `tradeNo`：支付平台交易号。
- `orderNo`：商城订单号。

幂等判断：

```go
if callbacks[tradeNo] {
    writeJSON(r, 0, "duplicate callback ignored", ...)
    return
}
callbacks[tradeNo] = true
```

意思是：同一个 `tradeNo` 只处理一次。第二次过来直接返回成功，但不再修改订单。

再次提醒：内存 map 只是教学演示。真实项目要靠数据库唯一键：

```sql
UNIQUE KEY uk_payment_callbacks_trade_no (trade_no)
```

## 支付回调为什么要幂等

第三方支付没有及时收到商城的成功响应时，通常会重复通知。同一个 `tradeNo` 可能到达多次。如果每次都更新余额、积分或订单，就会重复处理。

先执行本课 SQL：

```bash
mysql -uroot -p goframe_mall < lessons/lesson19/migration.sql
```

表上这一行是最终防线：

```sql
UNIQUE KEY uk_payment_callbacks_trade_no (trade_no)
```

回调事务的核心流程：

```text
验证回调签名与金额
        ↓
插入 payment_callbacks(trade_no)
        ↓
唯一键冲突 → 已经处理过，直接返回成功
        ↓
锁定订单并确认还是待支付
        ↓
更新订单为已支付
        ↓
提交事务后返回成功
```

不能只写“先查询 tradeNo 是否存在，再插入”。两个并发回调可能同时查到不存在。数据库唯一键才是并发下的硬保证。

## 练习：支付调用与重复回调

照着下面做：

1. 启动 `mock-payment/main.go`。
2. 启动 `mall-api/main.go`。
3. 访问 `GET /api/orders/1`，确认订单是 `pending`。
4. 访问 `POST /api/orders/1/pay`，确认订单变成 `paid`。
5. 用 `mockOrderNo=TIMEOUT` 验证 1.5 秒左右返回可控错误，而不是一直卡住。
6. 用 `mockOrderNo=FAIL` 验证商城能识别 HTTP 502。
7. 连续调用两次 `/api/payments/callback`，确认第二次返回 `duplicate: true`。
8. 进阶：把内存 `orders/callbacks` 改成 MySQL 表和事务，并执行 `migration.sql`。

## 验收条件

- 正常支付能解析 `tradeNo` 和 `status`。
- 所有成功创建的 `gclient.Response` 都会关闭。
- 网络错误、超时、非 2xx、非法 JSON 分别得到明确日志和可控响应。
- 下游调用沿用 Controller 的 `ctx`。
- 相同 `tradeNo` 重复回调不会重复修改订单。
- 幂等最终由数据库唯一键保证，不只依赖“先查询”。
- `go test ./...` 与 `go vet ./...` 通过。

完成后把四种调用结果和支付回调核心代码发给我，我先验收，再进入下一课。
