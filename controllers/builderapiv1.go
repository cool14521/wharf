package controllers

import (
	"github.com/astaxie/beego"
)

type BuilderAPIV1Controller struct {
	beego.Controller
}

func (this *BuilderAPIV1Controller) Prepare() {
	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)
}

func (this *BuilderAPIV1Controller) URLMapping() {
	this.Mapping("PostBuild", this.PostBuild)
	this.Mapping("GetStatus", this.GetStatus)
}

func (this *BuilderAPIV1Controller) PostBuild() {

}

func (this *BuilderAPIV1Controller) GetStatus() {

}
