# 第 21 课：测试与可替换接口

## 本课目标

这一课不追求“为了覆盖率而覆盖”，而是给商城最危险的行为建立保护网：

- 商品参数校验是否稳定。
- 并发创建订单是否会超卖。
- Session 登录与退出是否有效。
- JWT 的签名、类型、过期和撤销是否有效。
- 调支付服务时是否正确处理超时和异常。

本课使用：

- 标准库 `testing`。
- 标准库 `net/http/httptest`。
- GoFrame `test/gtest`。
- 表格测试。
- 用接口替换 Service 或外部支付客户端。
- 独立测试数据库做集成测试。

## 和 Gin 的对应关系

Gin 通常把路由交给 `httptest.NewRecorder`；GoFrame 的 Controller 依赖 `context.Context`，纯业务 Logic 更适合直接调用测试。

建议把测试分三层：

```text
Controller/HTTP 测试：请求绑定、中间件、状态码、响应格式
Logic 单元测试：业务分支，替换外部依赖
数据库集成测试：事务、锁、唯一约束、DAO SQL
```

不要把所有行为都塞进 HTTP 测试。库存并发是否安全，必须使用真实 MySQL 验证；一个假的内存 DAO 无法证明 `SELECT ... FOR UPDATE` 正确。

## 直接运行本课样例

```bash
go test -v ./lessons/lesson21/example
```

样例文件：

- `example/calculator.go`：按“分”计算总金额。
- `example/calculator_test.go`：标准库表格测试。
- `example/calculator_gtest_test.go`：GoFrame 断言。

## `testing`

一个测试函数必须符合：

```go
func TestXxx(t *testing.T)
```

- 文件名必须以 `_test.go` 结尾。
- 函数名必须以 `Test` 开头。
- `t` 的类型是 `*testing.T`，代表当前测试的状态和控制对象。

常用方法：

```go
t.Run(name string, f func(t *testing.T)) bool
t.Errorf(format string, args ...any)
t.Fatalf(format string, args ...any)
t.Helper()
t.Cleanup(f func())
```

- `Run`：创建一个命名子测试。
- `Errorf`：记录失败，但当前测试函数可以继续。
- `Fatalf`：记录失败并立即停止当前测试。
- `Helper`：把当前函数标成辅助函数，失败行号指向调用者。
- `Cleanup`：测试结束时执行清理，适合恢复全局 Service 或删除临时数据。

## 表格测试

本课样例的核心：

```go
tests := []struct {
	name      string
	priceCent int64
	quantity  int64
	want      int64
	wantErr   bool
}{
	{name: "正常计算", priceCent: 1999, quantity: 2, want: 3998},
	{name: "价格为零", priceCent: 0, quantity: 2, wantErr: true},
}
```

- `tests`：匿名结构体切片，每个元素是一种输入和预期。
- `name`：子测试名称，失败输出更容易定位。
- `want`：预期结果。
- `wantErr`：是否预期发生错误。

逐项执行：

```go
for _, test := range tests {
	t.Run(test.name, func(t *testing.T) {
		got, err := CalculateTotal(test.priceCent, test.quantity)
		if (err != nil) != test.wantErr {
			t.Fatalf("错误状态不符：err=%v", err)
		}
		if got != test.want {
			t.Errorf("got=%d want=%d", got, test.want)
		}
	})
}
```

表格测试适合商品校验：同一套执行逻辑覆盖商品名为空、价格为零、库存为负等输入。

## `gtest.C`

```go
gtest.C(t *testing.T, f func(t *gtest.T))
```

- 第一个 `t`：标准库测试对象。
- `f`：测试闭包。
- 闭包里的 `t`：GoFrame 的 `*gtest.T`，提供更简短的断言。

常用断言：

```go
t.Assert(actual, expected)
t.AssertEQ(actual, expected)
t.AssertNil(value)
```

- `Assert`：宽松比较，适合普通值。
- `AssertEQ`：要求值与类型都一致；`int64(7500)` 和 `int(7500)` 不相等。
- `AssertNil`：断言值为空。

例如：

```go
gtest.C(t, func(t *gtest.T) {
	total, err := CalculateTotal(2500, 3)
	t.AssertNil(err)
	t.AssertEQ(total, int64(7500))
})
```

`gtest` 不是新的测试运行器，最终仍由 `go test` 执行。

## `httptest.NewServer`

测试商城调用支付服务时，不要真的启动第 19 课的 9001 端口。标准库可以建立临时 HTTP 服务：

```go
httptest.NewServer(handler http.Handler) *httptest.Server
```

- `handler`：收到请求后执行的 HTTP 处理器。
- 返回值：监听随机本地端口的测试服务器。
- `server.URL`：随机服务地址。
- `server.Close()`：关闭服务。

示例：

```go
func TestPaymentClient_Pay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pay" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tradeNo":"T100","status":"paid"}`))
	}))
	defer server.Close()

	client := payment.NewPaymentClient(server.URL)
	output, err := client.Pay(context.Background(), payment.PayInput{
		OrderNo: "O100", AmountCent: 9900,
	})
	if err != nil {
		t.Fatal(err)
	}
	if output.TradeNo != "T100" {
		t.Fatalf("tradeNo=%s", output.TradeNo)
	}
}
```

这里测试的是商城的支付客户端，不是 Mock 服务本身。还应分别让 handler 返回 502、非法 JSON，或者故意等待到客户端超时。

## 为什么 Service 要设计成接口

假设 Controller 直接调用：

```go
service.Product().List(ctx, input)
```

生成的 `service.IProduct` 是接口。测试 Controller 时，可以提供一个假的实现：

```go
type fakeProductService struct {
	listOutput *model.ProductListOutput
	listErr    error
}

func (f *fakeProductService) List(
	ctx context.Context,
	in model.ProductListInput,
) (*model.ProductListOutput, error) {
	return f.listOutput, f.listErr
}
```

- `fakeProductService` 是测试替身。
- `listOutput` 决定成功时返回什么。
- `listErr` 决定是否模拟错误。
- 方法签名必须完整实现 `service.IProduct`；接口增加方法后，fake 也要补上。

如果生成的接口方法很多，可以在测试结构体里嵌入接口，再只覆盖本测试调用的方法：

```go
type fakeProductService struct {
	service.IProduct
	listOutput *model.ProductListOutput
}
```

未覆盖的方法如果被误调用会发生 panic，能及时暴露测试中没有预期的调用。

全局注册型 Service 要在测试结束后恢复：

```go
old := service.Product()
service.RegisterProduct(fake)
t.Cleanup(func() {
	service.RegisterProduct(old)
})
```

不要把会修改同一个全局 Service 的测试标成 `t.Parallel()`，否则它们会互相覆盖。

## 数据库集成测试

库存事务的测试必须连接独立数据库，例如：

```text
goframe_mall_test
```

禁止把自动清表的测试指向开发库或生产库。测试配置应单独放在测试环境中，并在开始前确认数据库名带 `_test`。

典型结构：

```go
func TestCreateOrder_NoOversell(t *testing.T) {
	ctx := context.Background()
	resetTestDatabase(t, ctx)
	seedProduct(t, ctx, 1, 5)

	var success atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := service.Order().Create(ctx, model.CreateOrderInput{
				ProductID: 1,
				Quantity:  1,
				IdempotencyKey: fmt.Sprintf("test-%d", index),
			})
			if err == nil {
				success.Add(1)
			}
		}(i)
	}
	wg.Wait()

	if success.Load() != 5 {
		t.Fatalf("成功订单数=%d want=5", success.Load())
	}
	assertProductStock(t, ctx, 1, 0)
}
```

- `sync.WaitGroup`：等待所有 goroutine 完成。
- `atomic.Int64`：并发安全地统计成功次数。
- 每个并发请求使用不同幂等键，因为这里要测试库存竞争，不是重复请求。
- 还要查询最终订单数和库存，不能只看返回错误。

运行竞态检测：

```bash
go test -race ./...
```

它主要发现 Go 内存数据竞争；数据库是否超卖仍要靠业务断言。

## Session 测试什么

使用一个临时测试服务器和 Cookie Jar，至少验证：

1. 未登录访问 `/admin/me` 被拒绝。
2. 登录响应产生 Session Cookie。
3. 携带同一个 Cookie 访问成功。
4. 退出后再次访问被拒绝。

测试前给 Server 配置独立 Redis DB 或带测试前缀的 Session Storage，结束后清理测试键。

## JWT 测试什么

JWT 工具函数适合表格测试：

- 正常 Access Token。
- 已过期 Token。
- 用错误密钥签名。
- 把 token 字符串任意改一个字符。
- Refresh Token 调业务接口。
- `jti` 写入 Redis 撤销列表后再次访问。
- 普通用户与管理员角色。

时间相关逻辑最好让签发函数接收 TTL 或可替换的当前时间，不要让测试真的等待 15 分钟。

## 常用命令

```bash
go test ./...
```

运行当前模块全部测试。即使你没写测试，它仍会编译各包；没有 `_test.go` 的包会显示 `[no test files]`。

```bash
go test -v ./...
```

显示每个测试和子测试名称。

```bash
go test -run TestCreateOrder ./internal/logic/order
```

只运行名称匹配的测试。

```bash
go test -race ./...
```

启用 Go 数据竞争检测。

```bash
go test -cover ./...
```

显示语句覆盖率。覆盖率是提示，不代表测试质量。

```bash
go vet ./...
```

静态检查可疑代码，例如错误的格式化参数；它不运行测试，也不能替代测试。

## 练习：为商城核心行为补测试

照着做：

1. 先运行本课 `calculator` 样例。
2. 给商品新增参数写表格测试，至少 5 个用例。
3. 用 `httptest.NewServer` 测支付成功、502、非法 JSON 和超时。
4. 用 Service fake 测 Controller 成功和错误响应。
5. 建立独立 MySQL 测试库，写 20 并发、库存 5 的订单测试。
6. 写 Session 的未登录、登录、退出测试。
7. 写 JWT 的正常、过期、类型错误、撤销、角色测试。
8. 运行 `go test -race ./...` 和 `go vet ./...`。

## 验收条件

- 商品校验包含正常与边界用例，失败时检查稳定错误码。
- 支付客户端四种分支均有自动测试。
- Controller 测试不连接真实支付服务。
- 库存集成测试使用独立测试库，并证明最终库存不会小于零。
- Session 和 JWT 的登录、退出/撤销都可自动验证。
- 测试会清理 Redis、数据库记录和全局 Service 注册。
- `go test ./...`、`go test -race ./...`、`go vet ./...` 通过。

完成后把测试输出和你最不确定的两个测试发给我，我先评审测试是否真的验证了业务，而不只看覆盖率数字。
