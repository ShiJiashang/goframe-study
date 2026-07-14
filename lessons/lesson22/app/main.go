package main

import (
	"context"
	"net/http"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
)

type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	TraceID string `json:"traceId"`
	CostMs  int64  `json:"costMs"`
}

type LiveReq struct {
	g.Meta `path:"/health/live" method:"get" tags:"System" summary:"存活检查"`
}

type LiveRes struct {
	Status string `json:"status" dc:"进程状态"`
}

type ReadyReq struct {
	g.Meta `path:"/health/ready" method:"get" tags:"System" summary:"就绪检查"`
}

type ReadyRes struct {
	Status string `json:"status" dc:"整体就绪状态"`
	MySQL  string `json:"mysql" dc:"MySQL状态"`
	Redis  string `json:"redis" dc:"Redis状态"`
}

type ConfigReq struct {
	g.Meta `path:"/system/config-safe" method:"get" tags:"System" summary:"安全配置预览"`
}

type ConfigRes struct {
	AppName      string `json:"appName"`
	Env          string `json:"env"`
	Debug        bool   `json:"debug"`
	JwtSecretSet bool   `json:"jwtSecretSet"`
	JwtSecretLen int    `json:"jwtSecretLen"`
}

type LogDemoReq struct {
	g.Meta `path:"/system/log-demo" method:"post" tags:"System" summary:"结构化日志示例"`

	OrderID        int64  `json:"orderId" d:"1001"`
	UserID         int64  `json:"userId" d:"12"`
	IdempotencyKey string `json:"idempotencyKey" d:"demo-key"`
}

type LogDemoRes struct {
	Logged bool   `json:"logged"`
	Trace  string `json:"trace"`
}

type SlowReq struct {
	g.Meta `path:"/system/slow" method:"get" tags:"System" summary:"慢请求，用于观察优雅关闭"`

	Seconds int `json:"seconds" in:"query" d:"5"`
}

type SlowRes struct {
	SleptSeconds int    `json:"sleptSeconds"`
	Status       string `json:"status"`
}

type ShutdownAfterReq struct {
	g.Meta `path:"/system/shutdown-after" method:"post" tags:"System" summary:"延迟关闭服务，教学演示用"`

	Seconds int `json:"seconds" in:"query" d:"5"`
}

type ShutdownAfterRes struct {
	Scheduled    bool `json:"scheduled"`
	AfterSeconds int  `json:"afterSeconds"`
}

type SystemController struct {
	server *ghttp.Server
}

func main() {
	ctx := context.Background()
	server := g.Server()

	address := cfgString(ctx, "server.address", ":8022")
	server.SetAddr(address)
	server.SetOpenApiPath(cfgString(ctx, "server.openapiPath", "/api.json"))
	server.SetSwaggerPath(cfgString(ctx, "server.swaggerPath", "/swagger"))
	server.SetGraceful(true)
	server.SetGracefulTimeout(5)
	server.SetGracefulShutdownTimeout(5)

	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(TraceMiddleware, ResponseMiddleware)
		group.Bind(&SystemController{server: server})
	})

	g.Log().Info(ctx, "lesson22 app starting", g.Map{
		"address": address,
		"env":     cfgString(ctx, "app.env", "local"),
	})
	server.Run()
}

func (c *SystemController) Live(ctx context.Context, _ *LiveReq) (*LiveRes, error) {
	return &LiveRes{Status: "ok"}, nil
}

func (c *SystemController) Ready(ctx context.Context, _ *ReadyReq) (*ReadyRes, error) {
	res := &ReadyRes{
		Status: "ok",
		MySQL:  "unknown",
		Redis:  "unknown",
	}

	if _, err := g.Cfg().Get(ctx, "database.default.link"); err != nil {
		res.Status = "not_ready"
		res.MySQL = "not_configured"
		g.Log().Warning(ctx, "readiness mysql config missing", g.Map{"dependency": "mysql", "reason": "config_missing"})
		setHTTPStatus(ctx, http.StatusServiceUnavailable)
		return res, nil
	}
	if err := g.DB().PingMaster(); err != nil {
		res.Status = "not_ready"
		res.MySQL = "error"
		g.Log().Error(ctx, "readiness mysql failed", g.Map{"error": err.Error()})
		setHTTPStatus(ctx, http.StatusServiceUnavailable)
		return res, nil
	}
	res.MySQL = "ok"

	if _, err := g.Cfg().Get(ctx, "redis.default.address"); err != nil {
		res.Status = "not_ready"
		res.Redis = "not_configured"
		g.Log().Warning(ctx, "readiness redis config missing", g.Map{"dependency": "redis", "reason": "config_missing"})
		setHTTPStatus(ctx, http.StatusServiceUnavailable)
		return res, nil
	}
	if _, err := g.Redis().Do(ctx, "PING"); err != nil {
		res.Status = "not_ready"
		res.Redis = "error"
		g.Log().Error(ctx, "readiness redis failed", g.Map{"error": err.Error()})
		setHTTPStatus(ctx, http.StatusServiceUnavailable)
		return res, nil
	}
	res.Redis = "ok"

	return res, nil
}

func (c *SystemController) ConfigSafe(ctx context.Context, _ *ConfigReq) (*ConfigRes, error) {
	secret, err := g.Cfg().GetEffective(ctx, "auth.jwtSecret", "")
	if err != nil {
		g.Log().Warning(ctx, "jwt secret config is not available", g.Map{"reason": "config_missing"})
	}
	secretText := ""
	if secret != nil {
		secretText = secret.String()
	}

	return &ConfigRes{
		AppName:      cfgString(ctx, "app.name", "GoFrame Mall Lesson22"),
		Env:          cfgString(ctx, "app.env", "local"),
		Debug:        cfgBool(ctx, "app.debug", false),
		JwtSecretSet: secretText != "",
		JwtSecretLen: len(secretText),
	}, nil
}

func (c *SystemController) LogDemo(ctx context.Context, req *LogDemoReq) (*LogDemoRes, error) {
	ctx, span := gtrace.NewSpan(ctx, "api.system.log_demo")
	defer span.End()

	g.Log().Info(ctx, "create order demo event", g.Map{
		"orderId":        req.OrderID,
		"userId":         req.UserID,
		"idempotencyKey": req.IdempotencyKey,
		"event":          "lesson22.log_demo",
	})

	return &LogDemoRes{
		Logged: true,
		Trace:  gtrace.GetTraceID(ctx),
	}, nil
}

func (c *SystemController) Slow(ctx context.Context, req *SlowReq) (*SlowRes, error) {
	if req.Seconds <= 0 {
		req.Seconds = 1
	}
	if req.Seconds > 30 {
		return nil, gerror.NewCode(gcode.CodeValidationFailed, "seconds must be <= 30")
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := 0; i < req.Seconds; i++ {
		select {
		case <-ctx.Done():
			return nil, gerror.Wrap(ctx.Err(), "request cancelled")
		case <-ticker.C:
			g.Log().Info(ctx, "slow request heartbeat", g.Map{
				"currentSecond": i + 1,
				"totalSeconds":  req.Seconds,
			})
		}
	}

	return &SlowRes{
		SleptSeconds: req.Seconds,
		Status:       "done",
	}, nil
}

func (c *SystemController) ShutdownAfter(ctx context.Context, req *ShutdownAfterReq) (*ShutdownAfterRes, error) {
	if req.Seconds <= 0 {
		req.Seconds = 5
	}
	if req.Seconds > 30 {
		return nil, gerror.NewCode(gcode.CodeValidationFailed, "seconds must be <= 30")
	}

	seconds := req.Seconds
	g.Log().Warning(ctx, "shutdown scheduled", g.Map{"afterSeconds": seconds})

	go func() {
		time.Sleep(time.Duration(seconds) * time.Second)
		bgCtx := context.Background()
		g.Log().Warning(bgCtx, "shutdown starting", g.Map{"afterSeconds": seconds})
		if err := c.server.Shutdown(); err != nil {
			g.Log().Error(bgCtx, "shutdown failed", g.Map{"error": err.Error()})
			return
		}
		g.Log().Info(bgCtx, "shutdown completed")
	}()

	return &ShutdownAfterRes{
		Scheduled:    true,
		AfterSeconds: seconds,
	}, nil
}

func cfgString(ctx context.Context, key string, def string) string {
	value, err := g.Cfg().Get(ctx, key)
	if err != nil || value.String() == "" {
		return def
	}
	return value.String()
}

func cfgBool(ctx context.Context, key string, def bool) bool {
	value, err := g.Cfg().Get(ctx, key)
	if err != nil {
		return def
	}
	return value.Bool()
}

func setHTTPStatus(ctx context.Context, status int) {
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		r.Response.Status = status
	}
}

func TraceMiddleware(r *ghttp.Request) {
	spanName := "http." + r.Method + " " + r.URL.Path
	ctx, span := gtrace.NewSpan(r.Context(), spanName)
	defer span.End()

	r.SetCtx(ctx)
	start := time.Now()
	r.Middleware.Next()

	status := r.Response.Status
	if status == 0 {
		status = http.StatusOK
	}
	g.Log().Info(r.Context(), "http request completed", g.Map{
		"method":  r.Method,
		"path":    r.URL.Path,
		"status":  status,
		"costMs":  time.Since(start).Milliseconds(),
		"traceId": gtrace.GetTraceID(r.Context()),
	})
}

func ResponseMiddleware(r *ghttp.Request) {
	start := time.Now()

	r.Middleware.Next()

	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}

	var (
		err                = r.GetError()
		res                = r.GetHandlerResponse()
		code    gcode.Code = gcode.CodeOK
		message            = code.Message()
		status             = r.Response.Status
	)
	if status == 0 {
		status = http.StatusOK
	}

	if err != nil {
		code = gerror.Code(err)
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
		message = err.Error()
		status = http.StatusInternalServerError
		if code == gcode.CodeValidationFailed {
			status = http.StatusBadRequest
		}
	} else if status >= http.StatusBadRequest {
		code = gcode.CodeInternalError
		message = "service is not ready"
	}

	r.Response.Status = status
	r.Response.WriteJson(ApiResponse{
		Code:    code.Code(),
		Message: message,
		Data:    res,
		TraceID: gtrace.GetTraceID(r.Context()),
		CostMs:  time.Since(start).Milliseconds(),
	})
}
