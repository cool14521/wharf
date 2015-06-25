package middleware

import (
	"github.com/Unknwon/macaron"
	_ "github.com/macaron-contrib/session/redis"
	"html/template"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Use(macaron.Static("external", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	m.Map(Log)
	m.Use(logger())
	//change the deafault template folder
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:  "views",
		Extensions: []string{".tmpl", ".html"},
		Funcs: []template.FuncMap{map[string]interface{}{
			"AppName": func() string {
				return "Macaron"
			},
			"AppVer": func() string {
				return "1.0.0"
			},
		}},
		Delims:          macaron.Delims{"{{", "}}"},
		Charset:         "UTF-8",
		IndentJSON:      true,
		IndentXML:       true,
		PrefixXML:       []byte("macaron"),
		HTMLContentType: "text/html",
	}))
	m.Use(macaron.Recovery())
	//set static resources folder
	m.Use(macaron.Static("external"))

}
