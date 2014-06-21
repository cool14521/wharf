package controllers

import (
  "github.com/astaxie/beego"
)

type PingController struct {
  beego.Controller
}

type PingResult struct {
  Result bool
}

func (this *PingController) Prepare() {
  this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
  this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
  this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
}

func (this *PingController) GetPing() {
  this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
  pingResult := PingResult{Result: true}
  this.Data["json"] = &pingResult
  this.ServeJson()
}
