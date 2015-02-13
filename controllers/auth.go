package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
)

type AuthController struct {
	beego.Controller
}

func (this *AuthController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *AuthController) Get() {
	this.TplNames = "auth.html"

	this.Data["description"] = ""
	this.Data["author"] = ""

	this.Render()
}
