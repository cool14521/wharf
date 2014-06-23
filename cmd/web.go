package cmd

import (
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
  Flags:       []cli.Flag{},
}

func runRegistry(*cli.Context) {
  models.InitDb()
  beego.Run()
}
