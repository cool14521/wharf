package cmd

import (
  "fmt"
  "strconv"

  "github.com/astaxie/beego"
  "github.com/codegangsta/cli"
  "github.com/dockboard/docker-registry/models"
  _ "github.com/dockboard/docker-registry/routers"
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

  models.InitDb()
  beego.Run(fmt.Sprintf("%v:%v", address, port))
}
