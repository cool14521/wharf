package web

import (
	"fmt"

	"gopkg.in/macaron.v1"

	"github.com/containerops/wharf/middleware"
	"github.com/containerops/wharf/router"
)

func SetWharfMacaron(m *macaron.Macaron) {
	//Setting Middleware
	middleware.SetMiddlewares(m)
	//Setting Router
	router.SetRouters(m)
}
