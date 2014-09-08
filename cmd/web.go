package cmd

import (
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/redis"
	"github.com/codegangsta/cli"
	"github.com/dockercn/docker-bucket/backup"
	"github.com/dockercn/docker-bucket/models"
	_ "github.com/dockercn/docker-bucket/routers"
	. "github.com/qiniu/api/conf"
)

var CmdWeb = cli.Command{
	Name:        "web",
	Usage:       "Start Docker Registry Web Server",
	Description: "Docker web service provide Docker Registry API service and web view for search & comment",
	Action:      runRegistry,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: "127.0.0.1",
			Usage: "Web 服务监听的 IP，默认 127.0.0.1",
		},
		cli.IntFlag{
			Name:  "port",
			Value: 9911,
			Usage: "Web 服务坚挺的端口，默认 9911",
		},
		cli.StringFlag{
			Name:  "qiniu_access",
			Value: "",
			Usage: "七牛云存储服务 Access Key",
		},
		cli.StringFlag{
			Name:  "qiniu_secret",
			Value: "",
			Usage: "七牛云存储服务 Secret Key",
		},
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
