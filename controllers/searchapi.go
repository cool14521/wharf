package controllers

import (
	"github.com/astaxie/beego"
)

type SearchAPIController struct {
	beego.Controller
}

func (this *SearchAPIController) Prepare() {
	this.EnableXSRF = false
}

func (this *SearchAPIController) GET() {

}
