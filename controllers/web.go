package controllers

import (
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
)

type WebController struct {
	beego.Controller
}

func (this *WebController) Prepare() {
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *WebController) GetIndex() {
	this.TplNames = "index.html"
	this.Render()
	this.StopRun()
}

func (this *WebController) GetAuth() {
	this.TplNames = "auth.html"
	this.Render()
	this.StopRun()
}

func (this *WebController) GetDashboard() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		beego.Error("[WEB API] Load session failure")
		this.Ctx.Redirect(http.StatusMovedPermanently, "/auth")
	} else {
		this.TplNames = "dashboard.html"
		this.Data["username"] = user.Username

		this.Render()
		this.StopRun()
	}
}

func (this *WebController) GetSetting() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		beego.Error("[WEB API] Load session failure")
		this.Ctx.Redirect(http.StatusMovedPermanently, "/auth")
	} else {
		this.TplNames = "setting.html"
		this.Data["username"] = user.Username

		this.Render()
		this.StopRun()
	}
}
