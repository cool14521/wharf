package controllers

import (
	"net/http"

	"github.com/astaxie/beego"
)

type DroneAPIController struct {
	beego.Controller
}

func (this *DroneAPIController) Prepare() {
	this.EnableXSRF = false
}

func (this *DroneAPIController) PostYAML() {
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.Output.Body([]byte("{\"status\":\"OK\"}"))
}
