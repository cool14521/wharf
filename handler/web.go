package handler

import (
	"github.com/Unknwon/macaron"
)

func DashboardHandler(ctx *macaron.Context) {
	ctx.HTML(200, "dashboard")
}

func SettingHandler(ctx *macaron.Context) {
	ctx.HTML(200, "setting")
}

func AdminAuthHandler(ctx *macaron.Context) {
	ctx.HTML(200, "admin/auth")
}
