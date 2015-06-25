package router

import (
	"github.com/Unknwon/macaron"
)

func SetRouters(m *macaron.Macaron) {

	m.Get("/", func(ctx *macaron.Context) {
		ctx.HTML(200, "index")
	})

}
