# 第 8 课：配置管理

## 本课目标

这节课学习 GoFrame 的配置读取：

- `g.Cfg()`：获取配置对象。
- `gcfg.Config`：配置管理对象类型。
- `Get` / `MustGet`：读取配置值。
- `GetEffective`：按“命令行 > 环境变量 > 配置文件 > 默认值”的优先级读取。
- 配置分组：用 `server.address`、`app.name` 这种点号路径读取 YAML。
- 环境变量覆盖：例如 `APP_ENV=test` 覆盖 `app.env`。

这节课已经在 `mall` 项目里新增了一个配置接口：

```text
GET /config/app
```

## 与 Gin 的对应关系

Gin 本身不内置完整配置系统，常见做法是配合 Viper：

```go
viper.GetString("server.address")
```

GoFrame 内置配置组件：

```go
g.Cfg().MustGet(ctx, "server.address").String()
```

对应关系：

```text
Viper GetString       -> GoFrame MustGet(...).String()
Viper GetInt          -> GoFrame MustGet(...).Int()
Viper env override    -> GoFrame GetEffective
```

## 配置文件位置

`mall` 项目的默认配置文件在：

```text
mall/manifest/config/config.yaml
```

本课新增了：

```yaml
app:
  name:  "GoFrame Mall"
  env:   "dev"
  debug: true

product:
  pageSize: 10
```

已有服务配置：

```yaml
server:
  address:     ":8000"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"
```

YAML 的层级可以用点号读取：

```text
app.name           -> GoFrame Mall
app.env            -> dev
server.address     -> :8000
product.pageSize   -> 10
```

## API 定义

文件：

```text
mall/api/config/v1/config.go
```

```go
type AppReq struct {
    g.Meta `path:"/config/app" method:"get" tags:"Config" summary:"Get app config"`
}
```

说明：

- `AppReq`：配置查询接口的请求结构体。
- `g.Meta`：标准路由元数据。
- `path:"/config/app"`：接口路径。
- `method:"get"`：HTTP 方法。
- `tags:"Config"`：Swagger 分组。
- `summary:"Get app config"`：接口摘要。

响应：

```go
type AppRes struct {
    Name         string `json:"name" dc:"应用名称"`
    Env          string `json:"env" dc:"配置文件中的运行环境"`
    EffectiveEnv string `json:"effectiveEnv" dc:"命令行或环境变量覆盖后的运行环境"`
    Debug        bool   `json:"debug" dc:"是否开启调试"`
    Address      string `json:"address" dc:"HTTP监听地址"`
    PageSize     int    `json:"pageSize" dc:"默认分页大小"`
}
```

## Controller 实现

文件：

```text
mall/internal/controller/config/config_v1_app.go
```

核心代码：

```go
cfg := g.Cfg()
```

`g.Cfg()`：

- 返回值类型：`*gcfg.Config`。
- 作用：获取默认配置对象。
- 默认会读取项目配置文件。

读取配置：

```go
cfg.MustGet(ctx, "app.name", "GoFrame Mall").String()
```

逐项解释：

- `ctx`：上下文。
- `"app.name"`：配置路径。
- `"GoFrame Mall"`：默认值，配置不存在时使用。
- `MustGet`：读取失败会 panic，适合启动配置或明确存在的配置。
- 返回值是动态值，继续调用 `.String()` 转成字符串。

读取布尔值：

```go
cfg.MustGet(ctx, "app.debug", false).Bool()
```

读取整数：

```go
cfg.MustGet(ctx, "product.pageSize", 10).Int()
```

## `Get` 和 `MustGet`

`Get`：

```go
value, err := cfg.Get(ctx, "app.name", "default")
```

- 返回两个值：`*gvar.Var` 和 `error`。
- 适合你想自己处理错误的场景。

`MustGet`：

```go
value := cfg.MustGet(ctx, "app.name", "default")
```

- 只返回 `*gvar.Var`。
- 内部如果遇到错误会 panic。
- 适合简单读取。

本课 Controller 使用 `MustGet` 读取普通配置。

## `GetEffective`：支持环境变量覆盖

代码：

```go
effectiveEnv, err := cfg.GetEffective(ctx, "app.env", "dev")
```

优先级：

```text
命令行参数 > 环境变量 > 配置文件 > 默认值
```

`app.env` 对应的环境变量是：

```text
APP_ENV
```

运行时可以这样覆盖：

```bash
APP_ENV=test go run main.go
```

这时：

```text
env           仍然来自配置文件，是 dev
effectiveEnv 来自环境变量，是 test
```

这能帮助你区分：

```text
MustGet/Get       主要读配置文件
GetEffective      允许命令行和环境变量覆盖
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

请求：

```bash
curl -s http://127.0.0.1:8000/config/app
```

你会看到类似：

```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "name": "GoFrame Mall",
    "env": "dev",
    "effectiveEnv": "dev",
    "debug": true,
    "address": ":8000",
    "pageSize": 10
  }
}
```

环境变量覆盖：

```bash
APP_ENV=test go run main.go
```

再请求：

```bash
curl -s http://127.0.0.1:8000/config/app
```

此时应该看到：

```json
{
  "env": "dev",
  "effectiveEnv": "test"
}
```

注意：上面只截取了关键字段，真实响应外层还有 `code/message/data`。

## 课后题：新增开发和测试配置

这次我给你“照葫芦画瓢”的步骤。

目标：新增一个配置项 `product.defaultSort`，并让接口返回它。

### 第 1 步：改配置文件

打开：

```text
mall/manifest/config/config.yaml
```

找到：

```yaml
product:
  pageSize: 10
```

改成：

```yaml
product:
  pageSize: 10
  defaultSort: "createdAtDesc"
```

### 第 2 步：改响应结构体

打开：

```text
mall/api/config/v1/config.go
```

在 `AppRes` 里新增字段：

```go
DefaultSort string `json:"defaultSort" dc:"商品默认排序"`
```

### 第 3 步：改 Controller

打开：

```text
mall/internal/controller/config/config_v1_app.go
```

在返回值里新增：

```go
DefaultSort: cfg.MustGet(ctx, "product.defaultSort", "createdAtDesc").String(),
```

照着已有的：

```go
PageSize: cfg.MustGet(ctx, "product.pageSize", 10).Int(),
```

写就行。

### 第 4 步：运行检查

在 `mall` 目录执行：

```bash
go test ./...
go vet ./...
```

### 第 5 步：运行服务

```bash
go run main.go
```

### 第 6 步：请求接口

```bash
curl -s http://127.0.0.1:8000/config/app
```

通过条件：

- 外层 `code` 是 `0`。
- `data.defaultSort` 是 `"createdAtDesc"`。
- 原来的 `name/env/effectiveEnv/debug/address/pageSize` 还在。

### 加分步骤：环境变量覆盖

如果你想试试覆盖配置，再新增一个字段：

```go
EffectiveSort string `json:"effectiveSort" dc:"环境变量覆盖后的默认排序"`
```

Controller 里照着 `effectiveEnv` 写：

```go
effectiveSort, err := cfg.GetEffective(ctx, "product.defaultSort", "createdAtDesc")
if err != nil {
    return nil, err
}
```

响应里：

```go
EffectiveSort: effectiveSort.String(),
```

运行：

```bash
PRODUCT_DEFAULT_SORT=priceAsc go run main.go
```

请求后应该看到：

```json
"defaultSort": "createdAtDesc",
"effectiveSort": "priceAsc"
```
