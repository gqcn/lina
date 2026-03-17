package main

import (
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/cmd"
)

func main() {
	c, err := gcmd.NewFromObject(cmd.Main{})
	if err != nil {
		panic(err)
	}
	c.Run(gctx.GetInitCtx())
}
