# 第 10 课：常用数据工具

## 本课目标

这节课用一个第三方支付数据解析接口学习：

- `gjson`：读取结构不固定的 JSON 数据。
- `gvar.Var`：包装类型不确定的动态值。
- `gconv`：把动态值转换成 Go 类型。
- `gtime`：解析、格式化时间并获取时间戳。
- `garray.TArray[T]`：使用 GoFrame 泛型数组并对标签去重。

本课接口：

```text
POST /debug/payment/parse
```

## 和 Gin 的对应关系

Gin 主要负责 HTTP 请求绑定。遇到第三方支付这种字段类型不稳定的数据，通常还需要标准库：

```go
var payload map[string]any
c.ShouldBindJSON(&payload)

amountText := payload["amountCent"].(string)
amountCent, err := strconv.ParseInt(amountText, 10, 64)
paidAt, err := time.Parse("2006-01-02 15:04:05", payload["paidAt"].(string))
```

这里的类型断言：

```go
payload["amountCent"].(string)
```

在实际类型不是 `string` 时会 panic。GoFrame 的工具可以更方便地读取和转换动态数据：

```go
paymentJSON := gjson.New(req.Payload)
amountValue := paymentJSON.Get("amountCent", 0)
amountCent := gconv.Int64(amountValue.Val())
paidAt, err := gtime.StrToTime(paymentJSON.Get("paidAt", "").String())
```

注意：方便转换不等于可以省略业务校验，转换后仍要检查金额是否合理。

## 请求和响应类型

文件：

```text
mall/api/debug/v1/debug.go
```

请求类型：

```go
type PaymentParseReq struct {
    g.Meta `path:"/debug/payment/parse" method:"post" tags:"Debug" summary:"解析第三方支付数据"`

    Payload g.Map `json:"payload" v:"required#支付数据不能为空" dc:"第三方支付原始数据"`
}
```

`g.Map` 的底层含义可以理解为：

```go
map[string]any
```

使用它是因为第三方支付数据中每个字段的类型可能不同：字符串、数字、布尔值或数组。

响应类型：

```go
type PaymentParseRes struct {
    TradeNo         string   `json:"tradeNo"`
    AmountCent      int64    `json:"amountCent"`
    PaidAt          string   `json:"paidAt"`
    PaidAtTimestamp int64    `json:"paidAtTimestamp"`
    Success         bool     `json:"success"`
    Tags            []string `json:"tags"`
}
```

## `gjson`：读取动态 JSON

导入包：

```go
import "github.com/gogf/gf/v2/encoding/gjson"
```

创建 JSON 对象：

```go
paymentJSON := gjson.New(req.Payload)
```

函数签名：

```go
func New(data any, safe ...bool) *gjson.Json
```

- `data any`：要包装的数据，可以是 JSON 字符串、`map`、结构体等。
- `safe ...bool`：可选参数，传 `true` 时启用并发安全；当前请求内局部使用，不需要传。
- 返回 `*gjson.Json`：支持按路径读取数据的 JSON 对象。

读取字段：

```go
tradeNoValue := paymentJSON.Get("tradeNo", "")
```

方法签名：

```go
func (j *Json) Get(pattern string, def ...any) *gvar.Var
```

- 接收者 `j *Json`：前面创建的 JSON 对象。
- `pattern`：字段路径，支持嵌套路径，例如 `buyer.name`。
- `def`：字段不存在时使用的默认值。
- 返回 `*gvar.Var`：包装了真实数据的动态值。

例如：

```go
paymentJSON.Get("buyer.name", "匿名用户").String()
paymentJSON.Get("items.0.productId", 0).Int64()
```

## `gvar.Var`：动态值包装器

`gjson.Get` 返回的不是固定的 `string` 或 `int64`，而是：

```go
*gvar.Var
```

它表示“当前有一个值，但原始类型不一定是什么”。

本课使用：

```go
tradeNo := paymentJSON.Get("tradeNo", "").String()

amountValue := paymentJSON.Get("amountCent", 0)
rawAmount := amountValue.Val()
```

常用方法：

```go
func (v *Var) Val() any
func (v *Var) String() string
func (v *Var) Int64() int64
func (v *Var) Bool() bool
func (v *Var) Strings() []string
```

- `Val()`：取出原始值，返回 `any`。
- `String()`：转换为字符串。
- `Int64()`：转换为 `int64`。
- `Bool()`：转换为 `bool`。
- `Strings()`：转换为 `[]string`。

`gvar.Var` 的转换方法内部也是调用 `gconv`。本课保留一部分显式 `gconv` 调用，方便你分别认识两个组件。

## `gconv`：类型转换

导入包：

```go
import "github.com/gogf/gf/v2/util/gconv"
```

本课代码：

```go
amountValue := paymentJSON.Get("amountCent", 0)
amountCent := gconv.Int64(amountValue.Val())

success := gconv.Bool(paymentJSON.Get("success", false).Val())

tags := gconv.Strings(paymentJSON.Get("tags", []string{}).Val())
```

函数签名：

```go
func Int64(anyInput any) int64
func Bool(anyInput any) bool
func Strings(anyInput any) []string
```

它们都接收 `any`，分别返回 `int64`、`bool` 和 `[]string`。

例如以下不同原始类型都可以转换：

```go
gconv.Int64("1999") // 1999
gconv.Int64(1999)   // 1999
gconv.Bool("true") // true
gconv.Bool(1)       // true
```

重要限制：这些便捷函数不返回 `error`。转换失败时通常得到对应类型的零值，例如：

```go
gconv.Int64("abc") // 0
```

所以代码转换后仍然检查：

```go
if amountCent <= 0 {
    return nil, gerror.NewCode(
        gcode.CodeInvalidParameter,
        "amountCent必须大于0",
    )
}
```

## `gtime`：时间处理

导入包：

```go
import "github.com/gogf/gf/v2/os/gtime"
```

解析支付时间：

```go
paidAt, parseErr := gtime.StrToTime(paidAtText)
```

函数签名：

```go
func StrToTime(str string, format ...string) (*gtime.Time, error)
```

- `str`：要解析的时间字符串。
- `format`：可选的 GoFrame 时间格式；标准日期时间可以省略。
- 返回 `*gtime.Time`：GoFrame 时间对象。
- 返回 `error`：格式无法解析时不为 `nil`。

解析失败时包装原始错误：

```go
if parseErr != nil {
    return nil, gerror.WrapCode(
        gcode.CodeInvalidParameter,
        parseErr,
        "paidAt格式错误",
    )
}
```

格式化：

```go
paidAt.Format("Y-m-d H:i:s")
```

方法签名：

```go
func (t *Time) Format(format string) string
```

这里使用的是 GoFrame 格式字符：

```text
Y 年，m 月，d 日，H 时，i 分，s 秒
```

获取 Unix 秒级时间戳：

```go
paidAt.Timestamp()
```

签名：

```go
func (t *Time) Timestamp() int64
```

## `garray.TArray[T]`：泛型数组

导入包：

```go
import "github.com/gogf/gf/v2/container/garray"
```

本课用它去除重复标签：

```go
tags := garray.NewTArrayFrom[string](
    gconv.Strings(paymentJSON.Get("tags", []string{}).Val()),
).Unique().Slice()
```

创建函数：

```go
func NewTArrayFrom[T comparable](array []T, safe ...bool) *TArray[T]
```

- `[string]`：泛型类型参数，表示数组只能存放字符串。
- `array []T`：初始切片。
- `safe`：可选并发安全开关，当前局部变量不需要开启。
- 返回 `*TArray[string]`。

链式方法：

```go
func (a *TArray[T]) Unique() *TArray[T]
func (a *TArray[T]) Slice() []T
```

- `Unique()`：原地删除重复值，并返回数组自己，便于继续链式调用。
- `Slice()`：取出普通 Go 切片，本例得到 `[]string`。

## 核心 Controller 代码

文件：

```text
mall/internal/controller/debug/debug_v1_payment_parse.go
```

代码执行顺序：

```text
req.Payload
    ↓ gjson.New
动态 JSON 对象
    ↓ Get
*gvar.Var
    ↓ gconv / Var转换方法
string、int64、bool、[]string
    ↓ gtime / garray
时间转换、标签去重
    ↓
PaymentParseRes
```

关键变量：

- `paymentJSON`：包装请求中动态支付数据的 `*gjson.Json`。
- `tradeNo`：转换后的支付单号 `string`。
- `amountValue`：金额原始动态值 `*gvar.Var`。
- `amountCent`：转换后的分金额 `int64`。
- `paidAtText`：支付时间原始字符串。
- `paidAt`：解析后的 `*gtime.Time`。
- `parseErr`：时间解析错误。
- `success`：转换后的支付状态 `bool`。
- `tags`：去重后的 `[]string`。

## 运行样例

进入商城目录：

```bash
cd mall
gf run main.go
```

另开终端请求：

```bash
curl -X POST 'http://127.0.0.1:8000/debug/payment/parse' \
  -H 'Content-Type: application/json' \
  -d '{
    "payload": {
      "tradeNo": "PAY20260713001",
      "amountCent": "1999",
      "paidAt": "2026-07-13 20:30:00",
      "success": "true",
      "tags": ["wechat", "mall", "wechat"]
    }
  }'
```

预期 `data`：

```json
{
  "tradeNo": "PAY20260713001",
  "amountCent": 1999,
  "paidAt": "2026-07-13 20:30:00",
  "paidAtTimestamp": 1783945800,
  "success": true,
  "tags": ["wechat", "mall"]
}
```

时间戳会按照运行程序的时区解析。

## 课后练习：计算支付净到账金额

目标：支付数据新增 `channel` 和 `feeCent`，返回手续费及净到账金额。

请求增加：

```json
{
  "channel": "wechat",
  "feeCent": "100"
}
```

你按下面步骤修改。

### 第一步：修改响应结构体

打开：

```text
mall/api/debug/v1/debug.go
```

给 `PaymentParseRes` 增加：

```go
Channel       string `json:"channel" dc:"支付渠道"`
FeeCent       int64  `json:"feeCent" dc:"手续费，单位为分"`
NetAmountCent int64  `json:"netAmountCent" dc:"净到账金额，单位为分"`
```

请求原始数据仍放在 `Payload` 中，因此不需要给 `PaymentParseReq` 增加字段。

### 第二步：读取并转换字段

打开：

```text
mall/internal/controller/debug/debug_v1_payment_parse.go
```

在构造响应前照着 `tradeNo` 和 `amountCent` 写：

```go
channel := paymentJSON.Get("channel", "unknown").String()

feeValue := paymentJSON.Get("feeCent", 0)
feeCent := gconv.Int64(/* 从 feeValue 取出原始值 */)
```

把注释部分补成你在 `amountValue` 中见过的方法调用。

### 第三步：校验并计算

要求：

```text
feeCent 不能小于 0
feeCent 不能大于 amountCent
netAmountCent = amountCent - feeCent
```

不合法时返回：

```go
gerror.NewCode(
    gcode.CodeInvalidParameter,
    "feeCent必须在0到支付金额之间",
)
```

### 第四步：填写响应字段

在 `PaymentParseRes` 中填写新增的三个变量：

```go
Channel:       channel,
FeeCent:       feeCent,
NetAmountCent: netAmountCent,
```

最后执行：

```bash
gofmt -w api/debug/v1/debug.go internal/controller/debug/debug_v1_payment_parse.go
go test ./...
go vet ./...
```

## 验收条件

使用下面的数据：

```json
{
  "amountCent": "1999",
  "channel": "wechat",
  "feeCent": "100"
}
```

必须满足：

- `channel` 返回 `wechat`。
- `feeCent` 返回数字 `100`，不是字符串。
- `netAmountCent` 返回 `1899`。
- 重复的 `tags` 仍然被去重。
- `feeCent=2000` 时响应 `code=53`，消息为“feeCent必须在0到支付金额之间”。
- `go test ./...` 和 `go vet ./...` 通过。

完成后把代码或响应发给我，我会先评审并提示，不会直接给完整答案。
