package middleware

import (
	"github.com/Unknwon/macaron"

	"github.com/containerops/wrench/setting"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Use(macaron.Static("static", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	InitLog(setting.RunMode, setting.LogPath)

	m.Map(Log)
	m.Use(logger(setting.RunMode))

	m.Use(macaron.Recovery())
}
