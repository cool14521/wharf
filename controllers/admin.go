package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
)

type AdminController struct {
	beego.Controller
}

func (this *AdminController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *AdminController) GetAdmin() {
	this.TplNames = "admin.html"

	this.Data["description"] = ""
	this.Data["author"] = ""

	this.Render()
}
