package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
)

type DashboardController struct {
	beego.Controller
}

func (this *DashboardController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *DashboardController) GetSetting() {
	this.TplNames = "setting.html"

	this.Data["description"] = ""
	this.Data["author"] = ""

	this.Render()
}

func (this *DashboardController) GetDashboard() {
	this.TplNames = "dashboard.html"

	this.Data["description"] = ""
	this.Data["author"] = ""
	this.Data["username"] = fmt.Sprint(this.GetSession("username"))

	this.Render()
}
