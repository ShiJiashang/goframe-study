package cmd

import (
	"context"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gsession"
	"github.com/gogf/gf/v2/os/gtime"

	"goframe-study/mall/internal/controller/auth"
	"goframe-study/mall/internal/controller/config"
	"goframe-study/mall/internal/controller/debug"
	"goframe-study/mall/internal/controller/health"
	"goframe-study/mall/internal/controller/order"
	"goframe-study/mall/internal/controller/product"
	"goframe-study/mall/internal/dao"
	_ "goframe-study/mall/internal/logic"
	"goframe-study/mall/internal/middleware"
)

var (
	Main = gcmd.Command{
		Name:  "mall",
		Usage: "mall <command>",
		Brief: "mall api command line",
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			// 不带子命令时，默认启动 HTTP 服务。
			// 也就是说：
			//   ./mall
			// 等价于：
			//   ./mall server
			return runServer(ctx)
		},
		Examples: `
mall server
mall seed
mall seed --name "测试商品" --price 1999 --stock 100
mall seed --insert --name "测试商品" --price 1999 --stock 100
`,
	}

	Server = gcmd.Command{
		Name:  "server",
		Usage: "mall server",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			return runServer(ctx)
		},
	}

	Seed = gcmd.Command{
		Name:   "seed",
		Usage:  "mall seed [OPTION]",
		Brief:  "insert demo product seed data",
		Strict: true,
		Arguments: []gcmd.Argument{
			{
				Name:    "insert",
				Short:   "i",
				Brief:   "really insert seed data into database",
				Orphan:  true,
				Default: "",
			},
			{
				Name:    "name",
				Short:   "n",
				Default: "GoFrame 测试商品",
				Brief:   "product name",
			},
			{
				Name:    "price",
				Short:   "p",
				Default: "1999",
				Brief:   "product price, unit is cent",
			},
			{
				Name:    "stock",
				Short:   "s",
				Default: "100",
				Brief:   "product stock",
			},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			return runSeed(ctx, parser)
		},
	}
)

func init() {
	if err := Main.AddCommand(&Server, &Seed); err != nil {
		panic(err)
	}
}

func runServer(ctx context.Context) error {
	s := g.Server()

	// Session：Redis 存储 + Cookie 名 + TTL
	s.SetSessionStorage(
		gsession.NewStorageRedis(g.Redis(), "mall:session:"),
	)
	s.SetSessionMaxAge(24 * time.Hour)
	s.SetSessionCookieMaxAge(24 * time.Hour)
	s.SetSessionIdName("mall_session_id")

	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		// AuthGate 内部：非 /admin/* 放行；Session 或 JWT 二选一
		group.Middleware(middleware.AuthGate)
		// AdminOnly 内部按 /admin/* 前缀过滤；必须挂在 AuthGate 之后
		group.Middleware(middleware.AdminOnly)
		group.Bind(
			config.NewV1(),
			debug.NewV1(),
			health.NewV1(),
			order.NewV1(),
			auth.NewV1(),
			product.NewV1(),
		)
	})

	s.Run()
	return nil
}

func runSeed(ctx context.Context, parser *gcmd.Parser) error {
	name := parser.GetOpt("name", "GoFrame 测试商品").String()
	price := parser.GetOpt("price", 1999).Int()
	stock := parser.GetOpt("stock", 100).Int()
	insert := parser.GetOpt("insert") != nil

	g.Log().Info(ctx, "seed product preview",
		"name", name,
		"priceCent", price,
		"stock", stock,
		"insert", insert,
	)

	if !insert {
		g.Log().Info(ctx, "当前是预览模式，不会写数据库；如果确认要写入，请加 --insert")
		return nil
	}

	_, err := dao.Products.Ctx(ctx).Data(g.Map{
		"category_id": 1,
		"name":        name,
		"price_cent":  price,
		"stock":       stock,
		"status":      1,
		"created_at":  gtime.Now(),
		"updated_at":  gtime.Now(),
	}).Insert()
	if err != nil {
		return err
	}

	g.Log().Info(ctx, "测试商品写入完成")
	return nil
}
