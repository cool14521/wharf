package controllers

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/dockercn/docker-bucket/global"
)

type StatusAPIController struct {
	beego.Controller
}

func (this *StatusAPIController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)

	//相应 docker api 命令的 Controller 屏蔽 beego 的 XSRF ，避免错误。
	this.EnableXSRF = false

	//设置 Response 的 Header 信息，在处理函数中可以覆盖
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", global.BucketConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", global.BucketConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", global.BucketConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", global.BucketConfig.String("docker::Encrypt"))
}

func (this *StatusAPIController) GET() {

}
