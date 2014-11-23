package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
)

type PingAPIController struct {
	beego.Controller
}

type PingResult struct {
	Result bool
}

func (this *PingAPIController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)

	//相应 docker api 命令的 Controller 屏蔽 beego 的 XSRF ，避免错误。
	this.EnableXSRF = false

	//设置 Response 的 Header 信息，在处理函数中可以覆盖
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))
}

func (this *PingAPIController) GetPing() {
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	pingResult := PingResult{Result: true}
	this.Data["json"] = &pingResult
	this.ServeJson()
}
