package controllers

import (
	"github.com/astaxie/beego"
)

type BuilderAPIV1Controller struct {
	beego.Controller
}

func (this *BuilderAPIV1Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		this.Data["json"] = map[string]string{"message": message}
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *BuilderAPIV1Controller) Prepare() {
	this.EnableXSRF = false
}

func (this *BuilderAPIV1Controller) URLMapping() {
	this.Mapping("PostBuild", this.PostBuild)
	this.Mapping("GetStatus", this.GetStatus)
}

func (this *BuilderAPIV1Controller) PostBuild() {

}

func (this *BuilderAPIV1Controller) GetStatus() {

}
