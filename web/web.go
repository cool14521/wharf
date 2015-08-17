package web

import (
	"fmt"

	"github.com/Unknwon/macaron"

	"github.com/containerops/wharf/middleware"
	"github.com/containerops/wharf/router"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/setting"
)

func SetWharfMacaron(m *macaron.Macaron) {
	//Setting Database
	if err := db.InitDB(setting.DBURI, setting.DBPasswd, setting.DBDB); err != nil {
		fmt.Printf("Connect Database Error %s", err.Error())
	}
	//Setting Middleware
	middleware.SetMiddlewares(m)
	//Setting Router
	router.SetRouters(m)
}
