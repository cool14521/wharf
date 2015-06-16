package middleware

import (
	"github.com/Unknwon/macaron"

	_ "github.com/macaron-contrib/session/redis"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Use(macaron.Static("external", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	m.Map(Log)
	m.Use(logger())

	m.Use(macaron.Recovery())
}
