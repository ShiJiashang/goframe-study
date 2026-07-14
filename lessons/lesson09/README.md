# 第 9 课：日志、错误与上下文

## 本课目标

这节课学习 GoFrame 项目里最常用的三件事：

- `g.Log()`：写日志。
- `gerror`：创建和包装错误。
- `gcode`：定义业务错误码。
- `context.Context`：把请求链路信息传下去。
- `gtrace.GetTraceID(ctx)`：获取当前请求的 TraceID。

本课已经在 `mall` 项目里新增调试接口：

```text
GET /debug/error
```

## 与 Gin 的对应关系

Gin 里你可能会这样写：

```go
log.Println("query product failed")
c.JSON(404, gin.H{"message": "商品不存在"})
```

GoFrame 更推荐：

```go
g.Log().Error(ctx, err)
return nil, gerror.NewCode(consts.CodeProductNotFound)
```

然后统一响应中间件会把错误码和错误消息输出给客户端。

## 自定义错误码

文件：

```text
mall/internal/consts/errors.go
```

代码：

```go
var (
    CodeProductNotFound  = gcode.New(10001, "商品不存在", nil)
    CodeOrderStockNotEnough = gcode.New(20001, "库存不足", nil)
    CodeAuthUnauthorized = gcode.New(30001, "未登录", nil)
)
```

`gcode.New`：

```go
func New(code int, message string, detail any) gcode.Code
```

参数解释：

- `code int`：业务错误码，例如 `10001`。
- `message string`：默认错误消息，例如 `"商品不存在"`。
- `detail any`：错误详情，当前不用，传 `nil`。

返回值：

- `gcode.Code`：GoFrame 的错误码接口。

为什么要自定义错误码？

```text
HTTP 状态码只能表达大类问题。
业务错误码能表达具体业务问题。
```

例如：

```text
10001 商品不存在
20001 库存不足
30001 未登录
```

## API 定义

文件：

```text
mall/api/debug/v1/debug.go
```

```go
type ErrorReq struct {
    g.Meta `path:"/debug/error" method:"get" tags:"Debug" summary:"Debug error and log"`

    Type string `json:"type" in:"query" d:"ok" dc:"错误类型：ok/notfound/wrap/panic"`
}
```

字段说明：

- `Type`：从 query 读取。
- `d:"ok"`：默认值是 `ok`。
- 可选值：`ok`、`notfound`、`wrap`、`panic`。

响应：

```go
type ErrorRes struct {
    TraceID string `json:"traceId" dc:"链路ID"`
    Message string `json:"message" dc:"调试消息"`
}
```

## 写日志：`g.Log()`

代码：

```go
g.Log().Infof(ctx, "debug error request type=%s traceId=%s", req.Type, traceID)
```

逐项解释：

- `g.Log()`：获取默认日志对象，类型是 `*glog.Logger`。
- `Infof`：按格式写 info 级别日志，类似 `fmt.Printf`。
- `ctx`：上下文。传入它后，日志能带上 TraceID。
- `%s`：字符串占位符。
- `req.Type`：请求中的错误类型。
- `traceID`：当前请求链路 ID。

常用日志级别：

```go
g.Log().Debug(ctx, "debug message")
g.Log().Info(ctx, "info message")
g.Log().Warning(ctx, "warning message")
g.Log().Error(ctx, "error message")
```

简单记：

```text
Debug    调试细节
Info     正常关键流程
Warning  可恢复但值得注意
Error    已经出错，需要排查
```

## TraceID

代码：

```go
traceID := gtrace.GetTraceID(ctx)
```

作用：

- 从 `context.Context` 里取当前请求的 TraceID。
- 同一次请求中的日志会带同一个 TraceID。
- 排查问题时，可以用 TraceID 串起请求、日志和响应。

浏览器或 curl 响应头里也能看到：

```text
Trace-Id: xxxxx
```

## 创建错误：`gerror.NewCode`

代码：

```go
err = gerror.NewCode(consts.CodeProductNotFound)
return nil, err
```

`gerror.NewCode`：

```go
func NewCode(code gcode.Code, text ...string) error
```

参数解释：

- `code gcode.Code`：错误码。
- `text ...string`：可选错误消息，不传就使用 code 自带的 message。

返回值：

- `error`：标准 Go 错误接口。

请求：

```bash
curl -s 'http://127.0.0.1:8000/debug/error?type=notfound'
```

响应类似：

```json
{
  "code": 10001,
  "message": "商品不存在",
  "data": null
}
```

## 包装错误：`gerror.WrapCode`

代码：

```go
baseErr := errors.New("database query returned empty result")
err = gerror.WrapCode(consts.CodeProductNotFound, baseErr, "查询商品失败")
return nil, err
```

`errors.New` 是 Go 标准库函数：

```go
func New(text string) error
```

它只能创建普通错误，没有业务码。

`gerror.WrapCode`：

```go
func WrapCode(code gcode.Code, err error, text ...string) error
```

参数解释：

- `code`：业务错误码。
- `err`：原始错误。
- `text`：当前层补充说明。

为什么要包装？

```text
底层错误告诉你“技术原因”。
外层错误告诉你“业务动作”。
错误码告诉客户端“业务分类”。
```

例如：

```text
database query returned empty result
查询商品失败
商品不存在
```

以后排查日志时，包装错误会比一个孤零零的 `"商品不存在"` 更有用。

## `panic` 不适合普通业务错误

样例里有：

```go
panic("模拟 panic：生产代码不要这样处理业务错误")
```

这是为了让你知道：

```text
panic 是程序级异常，不是普通业务失败。
```

普通业务问题，比如商品不存在、库存不足、未登录，应该用：

```go
return nil, gerror.NewCode(...)
```

不要用：

```go
panic(...)
```

## 运行样例

进入项目：

```bash
cd mall
```

运行：

```bash
go run main.go
```

如果 `8000` 被占用，先停掉旧服务，或者临时改配置端口。

正常请求：

```bash
curl -s 'http://127.0.0.1:8000/debug/error'
```

商品不存在：

```bash
curl -s 'http://127.0.0.1:8000/debug/error?type=notfound'
```

包装错误：

```bash
curl -s 'http://127.0.0.1:8000/debug/error?type=wrap'
```

观察终端日志，重点看：

```text
TraceID
日志级别
错误消息
错误堆栈
```

## 课后题：给商品列表加错误码

这次继续按步骤做。

目标：把商品列表接口里的“未实现”错误，改成我们自己的业务错误。

### 第 1 步：打开商品列表 Controller

文件：

```text
mall/internal/controller/product/product_v1_list.go
```

你现在可能会看到：

```go
return nil, gerror.NewCode(gcode.CodeNotImplemented)
```

### 第 2 步：改 import

你需要用到：

```go
"github.com/gogf/gf/v2/frame/g"
"goframe-study/mall/internal/consts"
```

如果 `gcode` 不再用了，就删掉：

```go
"github.com/gogf/gf/v2/errors/gcode"
```

### 第 3 步：在方法里写日志

在 `List` 方法里先写：

```go
g.Log().Info(ctx, "list product request")
```

### 第 4 步：返回一个假数据

把原来的未实现错误改成成功返回：

```go
res = &v1.ListRes{
    List: []v1.ProductItem{
        {
            ID:        1001,
            Name:      "GoFrame Book",
            PriceCent: 9900,
        },
    },
    Total: 1,
}
return
```

### 第 5 步：加一个错误分支

如果 `req.Page > 100`，返回库存不足错误码，先借用它练手：

```go
if req.Page > 100 {
    return nil, gerror.NewCode(consts.CodeOrderStockNotEnough, "页码太大，暂时不允许查询")
}
```

虽然这个错误码名字和分页不完全匹配，但这节课重点是练：

```text
gerror.NewCode + 自定义 gcode
```

第 12 课做商品分页时我们再定义更准确的分页错误码。

### 第 6 步：检查

在 `mall` 目录执行：

```bash
go test ./...
go vet ./...
```

### 第 7 步：运行服务

```bash
go run main.go
```

### 第 8 步：验收

正常请求：

```bash
curl -s 'http://127.0.0.1:8000/product/list?page=1&size=10'
```

通过条件：

- `code` 是 `0`。
- `data.total` 是 `1`。
- 终端能看到 `list product request` 日志。

错误请求：

```bash
curl -s 'http://127.0.0.1:8000/product/list?page=101&size=10'
```

通过条件：

- `code` 不是 `0`。
- `message` 是 `"页码太大，暂时不允许查询"`。
