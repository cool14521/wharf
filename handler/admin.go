package handler

import (
	"github.com/Unknwon/macaron"
)

func AdminAuthHandler(ctx *macaron.Context) {
	ctx.HTML(200, "admin/auth")
}
