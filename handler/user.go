package handler

import (
	"github.com/Unknwon/macaron"
)

func AuthHandler(ctx *macaron.Context) {
	ctx.HTML(200, "auth")
}
