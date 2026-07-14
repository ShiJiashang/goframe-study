# 第 17 课：GoFrame Session 鉴权

## 本课目标

这一课实现 Session 登录、查询当前用户、退出和后台接口保护：

- `*ghttp.Request.Session`。
- `Set`、`SetMap`、`Get`、`Remove`、`RemoveAll`。
- Redis Session 存储。
- Session 认证中间件。
- Cookie 中的 SessionID 和 Redis 中的会话数据分别是什么。

## 和 Gin 的对应关系

Gin 本身不提供 Session，常搭配第三方中间件：

```go
session := sessions.Default(c)
session.Set("userId", userID)
session.Save()
```

GoFrame 的每个 `*ghttp.Request` 已关联 Session：

```go
r := ghttp.RequestFromCtx(ctx)
err := r.Session.Set("userId", userID)
```

规范路由 Controller 没有直接接收 `*ghttp.Request`，所以仍通过 `ctx` 取得当前请求。

## Session 的工作方式

登录成功后：

```text
浏览器 Cookie：mall_session_id=随机SessionID
Redis：mall:session:{SessionID} → {userId, username, role}
```

客户端只保存 SessionID，不保存完整用户对象。后续请求携带 Cookie，服务端用 SessionID 从 Redis 读取登录状态。

Session 和 JWT 的区别放到第 18 课对比。

## 配置 Redis Session 存储

第 16 课 Redis 已配置完成。在 `internal/cmd/cmd.go` 创建 Server 后、启动前加入：

```go
s := g.Server()

s.SetSessionStorage(
    gsession.NewStorageRedis(g.Redis(), "mall:session:"),
)
s.SetSessionMaxAge(24 * time.Hour)
s.SetSessionCookieMaxAge(24 * time.Hour)
s.SetSessionIdName("mall_session_id")
```

需要导入：

```go
import (
    "time"

    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/gsession"
)
```

函数签名：

```go
func NewStorageRedis(
    redis *gredis.Redis,
    prefix ...string,
) *gsession.StorageRedis

func (s *Server) SetSessionStorage(storage gsession.Storage)
func (s *Server) SetSessionMaxAge(ttl time.Duration)
func (s *Server) SetSessionCookieMaxAge(maxAge time.Duration)
func (s *Server) SetSessionIdName(name string)
```

- `g.Redis()`：默认 Redis 客户端。
- `mall:session:`：Redis key 前缀。
- `SessionMaxAge`：服务端会话有效期。
- `SessionCookieMaxAge`：客户端 Cookie 有效期。
- `SessionIdName`：Cookie 和可选请求头使用的名称。

## 取得当前 Session

```go
r := ghttp.RequestFromCtx(ctx)
session := r.Session
```

变量类型：

```text
r        *ghttp.Request
session  *gsession.Session
```

`ghttp.Session` 是 `gsession.Session` 的类型别名，因此文档中可能看到两个包名。

## Session 常用方法

### `Set`

```go
func (s *Session) Set(key string, value any) error
```

```go
err := r.Session.Set("userId", int64(1))
```

设置或覆盖一个会话字段。

### `SetMap`

```go
func (s *Session) SetMap(data map[string]any) error
```

```go
err := r.Session.SetMap(map[string]any{
    "userId":  int64(1),
    "username": "demo",
    "role":     "admin",
})
```

一次写入多个字段。

### `Get`

```go
func (s *Session) Get(
    key string,
    def ...any,
) (*gvar.Var, error)
```

```go
value, err := r.Session.Get("userId")
```

- 找到时返回 `*gvar.Var`。
- 未登录或 key 不存在时可能返回 `nil`。
- `def` 是可选默认值。

读取前先处理 `err` 和 `nil`：

```go
if err != nil || value == nil || value.Int64() <= 0 {
    // 未登录或 Session 存储异常
}
```

### `Remove`

```go
func (s *Session) Remove(keys ...string) error
```

只删除部分字段：

```go
err := r.Session.Remove("role")
```

### `RemoveAll`

```go
func (s *Session) RemoveAll() error
```

退出登录时清空整个 Session：

```go
err := r.Session.RemoveAll()
```

最初计划中写的 `Clear` 是“清空会话”的概念；GoFrame v2.10.2 实际方法名是 `RemoveAll()`。

### `Id`

```go
func (s *Session) Id() (string, error)
```

返回当前 SessionID。不要把它写入普通业务日志或返回给不可信客户端。

## 可运行样例：Session 登录

API 定义示意：

```go
type SessionLoginReq struct {
    g.Meta  `path:"/auth/session/login" method:"post" tags:"Auth" summary:"Session登录"`
    Username string `json:"username" v:"required#用户名不能为空"`
    Password string `json:"password" v:"required#密码不能为空"`
}

type SessionLoginRes struct {
    UserID   int64  `json:"userId"`
    Username string `json:"username"`
}

type SessionMeReq struct {
    g.Meta `path:"/auth/session/me" method:"get" tags:"Auth" summary:"当前Session用户"`
}

type SessionMeRes struct {
    UserID   int64  `json:"userId"`
    Username string `json:"username"`
    Role     string `json:"role"`
}

type SessionLogoutReq struct {
    g.Meta `path:"/auth/session/logout" method:"post" tags:"Auth" summary:"Session退出"`
}

type SessionLogoutRes struct{}
```

登录 Controller 核心：

```go
func (c *ControllerV1) SessionLogin(
    ctx context.Context,
    req *v1.SessionLoginReq,
) (res *v1.SessionLoginRes, err error) {
    // 本课为了只关注 Session，使用固定学习账号。
    // 第18课会替换为数据库用户和 bcrypt 密码校验。
    if req.Username != "demo" || req.Password != "demo123" {
        return nil, gerror.NewCode(
            consts.CodeAuthUnauthorized,
            "用户名或密码错误",
        )
    }

    r := ghttp.RequestFromCtx(ctx)
    err = r.Session.SetMap(map[string]any{
        "userId":  int64(1),
        "username": "demo",
        "role":     "admin",
    })
    if err != nil {
        return nil, gerror.WrapCode(
            gcode.CodeOperationFailed,
            err,
            "保存登录状态失败",
        )
    }

    return &v1.SessionLoginRes{
        UserID:   1,
        Username: "demo",
    }, nil
}
```

这里的明文学习密码绝不能用于生产项目。

查询当前用户：

```go
func (c *ControllerV1) SessionMe(
    ctx context.Context,
    req *v1.SessionMeReq,
) (res *v1.SessionMeRes, err error) {
    r := ghttp.RequestFromCtx(ctx)

    userID, err := r.Session.Get("userId")
    if err != nil {
        return nil, gerror.WrapCode(gcode.CodeOperationFailed, err, "读取Session失败")
    }
    if userID == nil || userID.Int64() <= 0 {
        return nil, gerror.NewCode(consts.CodeAuthUnauthorized)
    }

    username, _ := r.Session.Get("username", "")
    role, _ := r.Session.Get("role", "")

    return &v1.SessionMeRes{
        UserID:   userID.Int64(),
        Username: username.String(),
        Role:     role.String(),
    }, nil
}
```

退出：

```go
func (c *ControllerV1) SessionLogout(
    ctx context.Context,
    req *v1.SessionLogoutReq,
) (res *v1.SessionLogoutRes, err error) {
    r := ghttp.RequestFromCtx(ctx)
    if err = r.Session.RemoveAll(); err != nil {
        return nil, gerror.WrapCode(gcode.CodeOperationFailed, err, "退出失败")
    }
    return &v1.SessionLogoutRes{}, nil
}
```

## Session 认证中间件

```go
func SessionAuth(r *ghttp.Request) {
    userID, err := r.Session.Get("userId")
    if err != nil {
        r.SetError(gerror.WrapCode(
            gcode.CodeOperationFailed,
            err,
            "读取登录状态失败",
        ))
        return
    }

    if userID == nil || userID.Int64() <= 0 {
        r.SetError(gerror.NewCode(consts.CodeAuthUnauthorized))
        return
    }

    // 类似 Gin 的 c.Set，供后续 Handler 使用。
    r.SetParam("currentUserId", userID.Int64())
    r.Middleware.Next()
}
```

如果认证失败，不调用 `Next()`，后续 Controller 不会执行；最外层统一响应中间件读取 `r.GetError()` 并输出错误。

注册时把登录接口和受保护接口分组：

```go
group.Bind(auth.NewV1())

group.Group("/admin", func(adminGroup *ghttp.RouterGroup) {
    adminGroup.Middleware(middleware.SessionAuth)
    adminGroup.Bind(admin.NewV1())
})
```

后台 Controller 的 `g.Meta path` 使用相对路径，例如 `path:"/me"`，最终路由是 `/admin/me`。

## 使用 curl 保存 Cookie

登录并保存 Cookie：

```bash
curl -c /tmp/mall-cookie.txt \
  -X POST 'http://127.0.0.1:8000/auth/session/login' \
  -H 'Content-Type: application/json' \
  -d '{"username":"demo","password":"demo123"}'
```

携带 Cookie：

```bash
curl -b /tmp/mall-cookie.txt \
  'http://127.0.0.1:8000/auth/session/me'
```

退出：

```bash
curl -b /tmp/mall-cookie.txt \
  -X POST 'http://127.0.0.1:8000/auth/session/logout'
```

查看 Redis Session key：

```bash
docker exec goframe-redis redis-cli \
  KEYS 'mall:session:*'
```

生产环境不要使用 `KEYS *` 扫描大型 Redis；本命令仅用于本地学习。

## 本课练习：保护后台商品接口

照着做：

1. 增加 `SessionAuth` 中间件。
2. 把后台商品新增、更新、删除放到受保护路由组。
3. 中间件把 `currentUserId` 写入请求参数。
4. 后台接口通过 `r.GetParam("currentUserId")` 读取操作者 ID 并记录日志。
5. 退出后用同一 Cookie 再请求，必须返回未登录。
6. 停止 Redis 后请求受保护接口，应返回可控错误并记录日志。

## 验收条件

- 登录响应设置 `mall_session_id` Cookie。
- Redis 中出现带 TTL 的 `mall:session:` key。
- 未登录不能访问后台接口。
- 登录后可以访问，且能读取当前用户 ID。
- 退出后旧 Session 不能继续使用。
- 认证失败时没有调用下游 Controller。
- Redis 异常时有明确日志和可控响应。
- `go test ./...` 和 `go vet ./...` 通过。

完成后提交登录、受保护请求、退出后请求以及 Redis TTL 结果。
