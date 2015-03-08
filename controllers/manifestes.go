package controllers

import (
	"net/http"

	"github.com/astaxie/beego"
)

type ManifestsAPIV2Controller struct {
	beego.Controller
}

func (this *ManifestsAPIV2Controller) URLMapping() {
}

func (this *ManifestsAPIV2Controller) Prepare() {
	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)
	beego.Debug(this.Ctx.Request.URL)

	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *ManifestsAPIV2Controller) PutManifests() {
	beego.Debug(this.Ctx.Request.Body)
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}
