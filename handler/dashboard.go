package handler

import (
	"github.com/Unknwon/macaron"
)

func DashboardHandler(ctx *macaron.Context) {
	ctx.HTML(200, "dashboard")
}
