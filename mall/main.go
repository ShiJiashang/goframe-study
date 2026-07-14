package main

import (
	_ "goframe-study/mall/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"goframe-study/mall/internal/cmd"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func main() {
	cmd.Seed.Run(gctx.GetInitCtx())
}
