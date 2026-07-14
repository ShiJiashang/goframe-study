# 第 18 课：JWT 鉴权

## 本课目标

这一课实现第二套登录方式：JWT。完成后你应该能解释：

- 密码为什么存 bcrypt 哈希，不能存明文。
- Access Token 和 Refresh Token 分别做什么。
- JWT 的 Claims、签名、过期时间和 `jti` 是什么。
- 无状态 JWT 为什么还需要 Redis 撤销列表。
- 登录、刷新、退出和管理员鉴权怎样串起来。

只安装本课真正使用的两个第三方包：

```bash
cd mall
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
```

## 和 Gin 的对应关系

Gin 和 GoFrame 都不会替你规定 JWT 格式。核心流程相同：

```text
读取 Authorization 请求头
        ↓
取出 Bearer 后面的 token
        ↓
校验签名、算法、过期时间和撤销状态
        ↓
把 userId、role 放进请求上下文
        ↓
继续执行 Controller
```

Gin 常用 `c.GetHeader`、`c.Set`；GoFrame 对应使用：

```go
authorization := r.Header.Get("Authorization")
r.SetParam("userId", claims.UserID)
r.Middleware.Next()
```

后面的 Controller 可以通过 `ghttp.Request.Get` 读取：

```go
userID := g.RequestFromCtx(ctx).Get("userId").Int64()
```

## 先分清四个概念

### 密码哈希

注册时把密码转换为不可逆哈希；登录时比较密码和哈希。数据库不保存明文密码。

### Access Token

访问业务接口时携带，生命周期应较短，本课设为 15 分钟。

### Refresh Token

Access Token 过期后用它换取新令牌，生命周期较长，本课设为 7 天。它不能直接访问普通业务接口。

### 撤销列表

JWT 签发后，单靠签名无法主动让它失效。退出登录时将令牌的唯一编号 `jti` 写入 Redis，鉴权时再检查它是否已撤销。

## Claims 类型

```go
type Claims struct {
	UserID   int64  `json:"userId"`
	Role     string `json:"role"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}
```

- `Claims`：我们定义的 JWT 载荷类型。
- `UserID`：当前用户 ID。
- `Role`：例如 `user` 或 `admin`。
- `TokenType`：固定为 `access` 或 `refresh`，防止拿 Refresh Token 调业务接口。
- `jwt.RegisteredClaims`：JWT 标准字段集合，包含 `ExpiresAt`、`IssuedAt`、`ID` 等。
- `ID`：标准字段 `jti`，应当每个 token 都不同。

`Claims` 只是编码，不是加密。拿到 token 的人可以读到内容，所以不要放密码、身份证号等敏感数据。

## bcrypt 的函数

```go
bcrypt.GenerateFromPassword(password []byte, cost int) ([]byte, error)
```

- `password`：明文密码的字节切片。
- `cost`：计算成本；通常使用 `bcrypt.DefaultCost`。
- 返回值：密码哈希和错误。

```go
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

登录时使用：

```go
bcrypt.CompareHashAndPassword(hashedPassword, password []byte) error
```

密码正确返回 `nil`，不正确返回错误：

```go
if err := bcrypt.CompareHashAndPassword(
	[]byte(user.PasswordHash),
	[]byte(req.Password),
); err != nil {
	return nil, gerror.New("用户名或密码错误")
}
```

## 创建并签名 Token

```go
jwt.NewWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token
```

- `method`：签名算法，本课用 `jwt.SigningMethodHS256`。
- `claims`：要放入 token 的声明。
- 返回 `*jwt.Token`，此时还没有生成最终字符串。

```go
token.SignedString(key any) (string, error)
```

- `key`：HS256 使用的密钥字节切片。
- 返回最终的 JWT 字符串和错误。

完整函数：

```go
func createToken(userID int64, role, tokenType string, ttl time.Duration, secret []byte) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        guid.S(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
```

- `guid.S()`：GoFrame `util/guid` 包生成随机唯一字符串，用作 `jti`。
- `ttl`：token 的有效时长。
- `jwt.NewNumericDate`：把 `time.Time` 包装成 JWT 标准时间类型。
- `secret`：签名密钥，应从配置或环境变量读取，不能提交真实生产密钥。

配置示例：

```yaml
auth:
  jwtSecret: "仅供本地学习，请在部署时由环境变量覆盖"
```

读取配置：

```go
func loadJWTSecret(ctx context.Context) ([]byte, error) {
	value, err := g.Cfg().GetEffective(ctx, "auth.jwtSecret")
	if err != nil {
		return nil, gerror.Wrap(err, "读取 JWT 密钥失败")
	}
	if value.IsEmpty() {
		return nil, gerror.New("JWT 密钥未配置")
	}
	return value.Bytes(), nil
}
```

`GetEffective` 会按“命令行参数 > 环境变量 > 配置文件”的优先级读取；配置键 `auth.jwtSecret` 对应环境变量 `AUTH_JWTSECRET`。普通配置可用 `MustGet`，密钥需要环境变量覆盖，因此这里使用 `GetEffective` 并显式处理错误。

## 解析与验证 Token

```go
jwt.ParseWithClaims(
	tokenString string,
	claims jwt.Claims,
	keyFunc jwt.Keyfunc,
	options ...jwt.ParserOption,
) (*jwt.Token, error)
```

- `tokenString`：请求传来的 JWT。
- `claims`：用于接收载荷的目标，例如 `&Claims{}`。
- `keyFunc`：返回验证签名所需密钥的函数。
- `options`：额外验证规则。
- 返回已解析 token 和错误。

本课明确限制算法为 HS256：

```go
func parseToken(tokenString string, secret []byte) (*Claims, error) {
	claims := new(Claims)
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		return nil, gerror.New("无效或已过期的 token")
	}
	return claims, nil
}
```

不要只解析 Claims 而不检查签名、算法和 `token.Valid`。

## 登录接口的核心代码

```go
type LoginRes struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func Login(ctx context.Context, userID int64, role string) (*LoginRes, error) {
	secret, err := loadJWTSecret(ctx)
	if err != nil {
		return nil, err
	}

	access, err := createToken(userID, role, "access", 15*time.Minute, secret)
	if err != nil {
		return nil, gerror.Wrap(err, "创建 access token 失败")
	}
	refresh, err := createToken(userID, role, "refresh", 7*24*time.Hour, secret)
	if err != nil {
		return nil, gerror.Wrap(err, "创建 refresh token 失败")
	}
	return &LoginRes{AccessToken: access, RefreshToken: refresh}, nil
}
```

实际项目要先根据用户名查询用户，再用 bcrypt 校验密码；这里省略 DAO 查询以突出 JWT 流程。

## Access Token 鉴权中间件

```go
func Auth(r *ghttp.Request) {
	value := r.Header.Get("Authorization")
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		r.SetError(gerror.New("缺少 Bearer token"))
		return
	}

	ctx := r.Context()
	secret, err := loadJWTSecret(ctx)
	if err != nil {
		r.SetError(err)
		return
	}
	claims, err := parseToken(parts[1], secret)
	if err != nil {
		r.SetError(err)
		return
	}
	if claims.TokenType != "access" {
		r.SetError(gerror.New("该 token 不能访问业务接口"))
		return
	}

	revoked, err := g.Redis().Get(ctx, "mall:jwt:revoked:"+claims.ID)
	if err != nil {
		r.SetError(gerror.Wrap(err, "检查 token 状态失败"))
		return
	}
	if !revoked.IsNil() {
		r.SetError(gerror.New("token 已退出登录"))
		return
	}

	r.SetParam("userId", claims.UserID)
	r.SetParam("role", claims.Role)
	r.SetParam("tokenClaims", claims)
	r.Middleware.Next()
}
```

关键点：

- `r.Header.Get` 获取请求头。
- `strings.SplitN(value, " ", 2)` 只分成两段。
- 中间件失败时 `SetError` 后直接 `return`，不要调用 `Next()`。
- Redis 出错时也拒绝请求。本课选择“安全优先”，不能在无法确认撤销状态时放行。
- `SetParam` 类似 Gin 的 `c.Set`，供后续中间件和 Controller 使用。

## 管理员中间件

管理员中间件必须放在 JWT 中间件之后：

```go
func AdminOnly(r *ghttp.Request) {
	if r.Get("role").String() != "admin" {
		r.SetError(gerror.New("需要管理员权限"))
		return
	}
	r.Middleware.Next()
}
```

绑定顺序：

```go
group.Middleware(middleware.Auth)
group.Group("/admin", func(admin *ghttp.RouterGroup) {
	admin.Middleware(middleware.AdminOnly)
	admin.Bind(controller.Admin)
})
```

## 刷新 Token

刷新接口必须确认收到的是 Refresh Token：

```go
func Refresh(ctx context.Context, refreshToken string) (string, error) {
	secret, err := loadJWTSecret(ctx)
	if err != nil {
		return "", err
	}
	claims, err := parseToken(refreshToken, secret)
	if err != nil {
		return "", err
	}
	if claims.TokenType != "refresh" {
		return "", gerror.New("必须使用 refresh token")
	}

	revoked, err := g.Redis().Get(ctx, "mall:jwt:revoked:"+claims.ID)
	if err != nil {
		return "", err
	}
	if !revoked.IsNil() {
		return "", gerror.New("refresh token 已失效")
	}
	return createToken(claims.UserID, claims.Role, "access", 15*time.Minute, secret)
}
```

更严格的生产做法是每次刷新同时签发新的 Refresh Token，并撤销旧 Refresh Token，这叫“Refresh Token 轮换”。

## 退出登录：写入 Redis 撤销列表

```go
func revokeToken(ctx context.Context, claims *Claims) error {
	seconds := int64(time.Until(claims.ExpiresAt.Time).Seconds())
	if seconds <= 0 {
		return nil
	}
	key := "mall:jwt:revoked:" + claims.ID
	return g.Redis().SetEX(ctx, key, 1, seconds)
}
```

`SetEX(ctx, key, value, ttlSeconds) error` 在写入的同时设置过期时间。Redis 键只保留到 token 原本的过期时间，不会永久堆积。

退出接口应同时接收并撤销当前 Access Token 和对应 Refresh Token。只撤销 Access Token 时，Refresh Token 仍能换出新的 Access Token。

## Session 和 JWT 怎么选

| 对比项   | Session             | JWT               |
| ----- | ------------------- | ----------------- |
| 客户端保存 | Session ID Cookie   | Token             |
| 服务端状态 | 必须保存会话              | 基本无状态，撤销时仍需 Redis |
| 主动退出  | 删除 Session          | 将 `jti` 加入撤销列表    |
| 适合场景  | 浏览器后台、传统 Web        | App、前后端分离、跨服务调用   |
| 主要风险  | Cookie/Session 配置错误 | 密钥泄漏、token 生命周期过长 |

商城同时实现两套入口是为了学习；实际项目通常根据客户端和安全要求选择，不要求所有接口同时支持两种凭证。

## 练习：完成 JWT 登录、刷新、退出和管理员接口

照着下面顺序做：

1. 建立 `internal/model/jwt_claims.go`，写出 `Claims`。
2. 建立 JWT 工具文件，先实现 `createToken` 和 `parseToken`。
3. 注册一个测试用户，把密码用 bcrypt 生成哈希后保存。
4. 登录成功返回 Access Token 和 Refresh Token。
5. 写 `Auth` 中间件，解析 Bearer Token，并检查 Redis 撤销键。
6. 写 `/auth/refresh`，只允许 Refresh Token 换新 Access Token。
7. 写 `/auth/logout`，撤销两个 token。
8. 写 `/admin/products`，依次经过 `Auth` 和 `AdminOnly`。

请求示例：

```bash
curl -X POST http://127.0.0.1:8000/auth/jwt/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"123456"}'

curl http://127.0.0.1:8000/admin/products \
  -H 'Authorization: Bearer 这里替换为accessToken'
```

## 验收条件

- 数据库中只有 bcrypt 哈希，没有明文密码。
- Access Token 可以访问受保护接口，Refresh Token 不可以。
- Access Token 过期或签名被修改时返回稳定错误响应。
- 普通用户访问管理员接口被拒绝，管理员可以访问。
- Refresh Token 可以换取新 Access Token。
- 退出后，原 Access Token 和 Refresh Token 都不能再使用。
- Redis 撤销键带 TTL，且不会永久保存。
- 鉴权失败时中间件没有调用 `r.Middleware.Next()`。

## 交付物汇总（本地实现）

### 目录结构

```
mall/
├── api/auth/
│   ├── auth.go                            # 接口 IAuthV1 新增三个 JWT 方法
│   └── v1/auth.go                         # JWTLoginReq/Res, JWTRefreshReq/Res, JWTLogoutReq/Res
├── internal/
│   ├── consts/errors.go                   # 新增 30101~30201 JWT/管理员错误码
│   ├── controller/auth/
│   │   ├── auth_v1_jwt_login.go           # 查数据库 + bcrypt 验证 + 签发双 token
│   │   ├── auth_v1_jwt_refresh.go         # refresh token 换 access token
│   │   └── auth_v1_jwt_logout.go          # 撤销 access + refresh
│   ├── middleware/
│   │   ├── session.go                     # tryParseSession 抽公用逻辑
│   │   ├── jwt.go                         # tryParseJWT + JWTAuth 独立中间件
│   │   └── authgate.go                    # AuthGate（Session 或 JWT 二选一）+ AdminOnly
│   ├── model/jwt_claims.go                # Claims 类型 + TokenType 常量
│   └── cmd/cmd.go                         # AuthGate + AdminOnly 挂到 group
├── utility/jwtutil/jwt.go                 # LoadSecret/Create/Parse/Revoke/CheckRevoked
└── manifest/config/config.yaml            # 新增 auth.jwtSecret

lessons/lesson18/seed.sql                  # 插入 admin/admin123 与 demo/demo123 的 bcrypt 哈希
```

### 关键设计

- **Session 与 JWT 并存**：`AuthGate` 中间件先看 Session，再看 Bearer JWT，任一有效即放行。浏览器与 App 场景共用一套后端。
- **AdminOnly 挂在 AuthGate 之后**：内部按 `/admin/*` 前缀过滤，非 admin role 直接 30201 拒绝。
- **撤销列表 TTL**：`SetEX(ctx, key, 1, seconds)` 中 `seconds` 由 `time.Until(claims.ExpiresAt.Time)` 计算，token 到期后 Redis 自动清理，不会永久堆积。
- **算法白名单**：`jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()})` 堵死 none/algorithm confusion 攻击。
- **密钥读取**：`g.Cfg().GetEffective(ctx, "auth.jwtSecret")` 按“命令行参数 > 环境变量 AUTH_JWTSECRET > 配置文件”优先级读取。

### 验收测试记录

| 场景 | 结果 | 说明 |
|---|---|---|
| admin/admin123 登录 | ✅ code=0，返回 accessToken/refreshToken/expiresIn=900/role=admin | |
| 密码错误 | ✅ code=30001 用户名或密码错误 | |
| 未带 token 访问 /admin/products/1 | ✅ code=30001 未登录 | |
| 带 access token 访问 /admin/products/1 | ✅ code=0 | |
| 带 refresh token 访问 /admin/products/1 | ✅ code=30103 token 类型不匹配 | |
| 签名被篡改 | ✅ code=30102 无效或已过期 | |
| refresh 换新 access | ✅ code=0 返回新 accessToken | |
| access 冒充 refresh 换 access | ✅ code=30103 必须使用 refresh token | |
| 退出登录（同时撤销两 token） | ✅ code=0 | |
| 退出后旧 access 访问 | ✅ code=30104 token 已退出登录 | |
| 退出后旧 refresh 换 access | ✅ code=30104 refresh token 已失效 | |
| 带 Session cookie 访问 /admin/products/1 | ✅ code=0（AuthGate 双通道生效）| |
| Redis 撤销键 TTL | ✅ access ~865s（<900），refresh ~604751s（<604800）| |
| demo 用户 JWT 访问 /admin/products/1 | ✅ code=30201 需要管理员权限 | |
| admin 用户 JWT 访问 /admin/products/1 | ✅ code=0 | |
| demo 用户访问 /products/1（公开）| ✅ code=0 | |

### 数据库准备

lesson11 已建 `users` 表（含 `password_hash` 列）。运行 seed：

```bash
docker exec -i goframe-mysql mysql -uroot -p12345678 goframe_mall \
    < lessons/lesson18/seed.sql
```

seed 后 `users` 表包含：`admin/admin123` 与 `demo/demo123` 两个账号，均为 bcrypt 哈希，无明文。role 由代码根据 username 判定（`admin` → `admin`，其他 → `user`）。

### 未做的事（留给下一课或生产化改造）

- Refresh Token 轮换（每次刷新签发新 refresh 并撤销旧的）
- 多端登录管理（按 userId 维度批量撤销所有 jti）
- role 存到 users 表或独立的 roles 表
- Session 登录接口也接数据库（当前仍是 lesson17 的固定账号 demo/demo123）


完成后把登录响应、刷新响应、退出后的访问响应，以及关键代码发给我。我先验收和提示，不提前给完整参考答案。
