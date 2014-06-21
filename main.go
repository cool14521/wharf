package main

import (
  "github.com/astaxie/beego"
  "github.com/dockboard/docker-registry/models"
  _ "github.com/dockboard/docker-registry/routers"
)

func main() {
  models.InitDb()
  beego.Run()
}
