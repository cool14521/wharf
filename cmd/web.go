package cmd

import (
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/redis"
	"github.com/codegangsta/cli"
	"github.com/dockboard/docker-registry/backup"
	"github.com/dockboard/docker-registry/models"
	_ "github.com/dockboard/docker-registry/routers"
	. "github.com/qiniu/api/conf"
)

var CmdWeb = cli.Command{
	Name:        "web",
	Usage:       "Start Docker Registry Web Server",
	Description: "Docker web service provide Docker Registry API service and web view for search & comment",
	Action:      runRegistry,
	Flags: []cli.Flag{
		cli.StringFlag{"address", "", "docker registry listen ip"},
		cli.IntFlag{"port", 9911, "docker registry listen port"},
		cli.StringFlag{"qiniu_access", "", "Qiniu.com's access key for backup docker image layer file."},
		cli.StringFlag{"qiniu_secret", "", "Qiniu.com's secret key for backup docker image layer file."},
	},
}

func runRegistry(c *cli.Context) {
	var address, port string

	beego.Trace("[Address] " + strconv.Itoa(len(c.String("address"))))
	beego.Trace("[Port] " + strconv.Itoa(c.Int("port")))

	//检查 address / port 的合法性
	if len(c.String("address")) > 0 {
		address = c.String("address")
	}

	if len(c.String("port")) > 0 {
		port = strconv.Itoa(c.Int("port"))
	}

	if len(c.String("qiniu_access")) > 0 {
		ACCESS_KEY = c.String("qiniu_access")
	}

	if len(c.String("qiniu_secret")) > 0 {
		SECRET_KEY = c.String("qiniu_secret")
	}

	if len(c.String("qiniu_access")) > 0 && len(c.String("qiniu_secret")) > 0 {
		backup.Backup = true

		go backup.QiniuBackup(backup.UploadChan, backup.ResultChan)
		go backup.QiniuResult(backup.ResultChan)
	}

	beego.SessionProvider = "redis"
	beego.SessionSavePath = "127.0.0.1:6379"

	models.InitDb()
	beego.Run(fmt.Sprintf("%v:%v", address, port))
}
