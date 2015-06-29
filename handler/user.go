package handler

import (
	"github.com/Unknwon/macaron"
)

func InitAuthHandler(ctx *macaron.Context) {
	ctx.HTML(200, "auth")
}
