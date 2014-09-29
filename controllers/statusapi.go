package controllers

import (
	"github.com/astaxie/beego"
)

type StatusAPIController struct {
	beego.Controller
}

func (this *StatusAPIController) Prepare() {
	this.EnableXSRF = false
}

func (this *StatusAPIController) GET() {

}
