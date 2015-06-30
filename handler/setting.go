package handler

import (
	"github.com/Unknwon/macaron"
)

func SettingHandler(ctx *macaron.Context) {
	ctx.HTML(200, "setting")
}
