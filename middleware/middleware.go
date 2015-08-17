package middleware

import (
	"html/template"

	"github.com/Unknwon/macaron"
	_ "github.com/macaron-contrib/session/redis"

	"github.com/containerops/wrench/setting"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Use(macaron.Static("external", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	m.Map(Log)
	m.Use(logger(setting.RunMode))
	//modify  default template setting
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:       "views",
		Extensions:      []string{".tmpl", ".html"},
		Funcs:           []template.FuncMap{},
		Delims:          macaron.Delims{"<<<", ">>>"},
		Charset:         "UTF-8",
		IndentJSON:      true,
		IndentXML:       true,
		PrefixXML:       []byte("macaron"),
		HTMLContentType: "text/html",
	}))
	m.Use(macaron.Recovery())
}
