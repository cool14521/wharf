package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
)

type WebController struct {
	beego.Controller
}

func (this *WebController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *WebController) GetIndex() {
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.Output.Body([]byte("{\"status\":\"OK\"}"))
}

func (this *WebController) GetAuth() {
	this.TplNames = "auth.html"
	this.Render()
}

func (this *WebController) GetDashboard() {
	user, ok := this.Ctx.Input.CruSession.Get("user").(models.User)
	if !ok {
		beego.Error(fmt.Sprintf("[WEB 用户] session加载失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "session加载失败")))
		return
	}

	this.TplNames = "dashboard.html"
	this.Data["username"] = user.Username

	this.Render()
}

func (this *WebController) GetSetting() {
	user, ok := this.Ctx.Input.CruSession.Get("user").(models.User)
	if !ok {
		beego.Error(fmt.Sprintf("[WEB 用户] session加载失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "session加载失败")))
		return
	}

	this.TplNames = "setting.html"

	this.Data["username"] = user.Username

	this.Render()
}
