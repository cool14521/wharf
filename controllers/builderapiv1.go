package controllers

import (
	"github.com/astaxie/beego"
)

type BuilderAPIV1Controller struct {
	beego.Controller
}

func (this *BuilderAPIV1Controller) Prepare() {
}

func (this *BuilderAPIV1Controller) URLMapping() {
	this.Mapping("PostBuild", this.PostBuild)
	this.Mapping("GetStatus", this.GetStatus)
}

func (this *BuilderAPIV1Controller) PostBuild() {

}

func (this *BuilderAPIV1Controller) GetStatus() {

}
