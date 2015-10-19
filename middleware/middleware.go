package middleware

import (
	"html/template"

	"github.com/Unknwon/macaron"
	_ "github.com/macaron-contrib/session/redis"

	"github.com/containerops/wrench/setting"
)

func SetMiddlewares(m *macaron.Macaron) {
	//Static root folder
	m.Use(macaron.Static("external", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))
	//Log
	InitLog(setting.RunMode, setting.LogPath)
	m.Map(Log)
	m.Use(logger(setting.RunMode))
	//Modify default template setting
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:       "views",
		Extensions:      []string{".tmpl", ".html"},
		Funcs:           []template.FuncMap{},
		Delims:          macaron.Delims{"<<<", ">>>"},
		Charset:         "UTF-8",
		IndentJSON:      true,
		IndentXML:       true,
		PrefixXML:       []byte(""),
		HTMLContentType: "text/html",
	}))
	//Recovery
	m.Use(macaron.Recovery())
}
