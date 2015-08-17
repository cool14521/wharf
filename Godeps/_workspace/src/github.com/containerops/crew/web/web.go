package web

import (
	"github.com/Unknwon/macaron"

	"github.com/containerops/crew/middleware"
	"github.com/containerops/crew/router"
)

func SetCrewMacaron(m *macaron.Macaron) {
	//Setting Middleware
	middleware.SetMiddlewares(m)
	//Setting Router
	router.SetRouters(m)
}
