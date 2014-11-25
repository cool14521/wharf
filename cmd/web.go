package cmd

import (
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/redis"
	"github.com/codegangsta/cli"

	"github.com/dockercn/docker-bucket/models"
	_ "github.com/dockercn/docker-bucket/routers"
)

var CmdWeb = cli.Command{
	Name:        "web",
	Usage:       "启动 Docker Bucket 的 Web 服务",
	Description: "Docker Bucket 提供 Docker Registry 服务的同时，还提供 Build、 CI 和 CD 服务。",
	Action:      runWeb,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: "0.0.0.0",
			Usage: "Web 服务监听的 IP，默认 0.0.0.0",
		},
		cli.IntFlag{
			Name:  "port",
			Value: 80,
			Usage: "Web 服务监听的端口，默认 80",
		},
		cli.StringFlag{
			Name:  "conf",
			Value: "",
			Usage: "Web 服务的配置文件路径",
		},
	},
}

func runWeb(c *cli.Context) {
	var address, port string

	//TODO 检查 address / port 的合法性
	if len(c.String("address")) > 0 {
		address = c.String("address")
	}

	if len(c.String("port")) > 0 {
		port = strconv.Itoa(c.Int("port"))
	}

	//设定 HTTP 的静态文件处理地址
	beego.SetStaticPath(beego.AppConfig.String("docker::StaticPath"), fmt.Sprintf("%s/images", beego.AppConfig.String("docker::BasePath")))
	//初始化 Session
	models.InitSession()
	//初始化 数据库
	models.InitDb()
	//运行 Application
	beego.Run(fmt.Sprintf("%v:%v", address, port))
}
