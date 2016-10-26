package middleware

import (
	"html/template"

	"gopkg.in/macaron.v1"
)

func SetMiddlewares(m *macaron.Macaron) {
	//Static root folder
	m.Use(macaron.Static("external", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	//Log
	m.Map(Log)
	m.Use(logger(setting.RunMode))

	//Recovery
	m.Use(macaron.Recovery())
}
