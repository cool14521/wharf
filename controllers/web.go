package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
)

type WebController struct {
	beego.Controller
}

func (this *WebController) URLMapping() {
	this.Mapping("GetIndex", this.GetIndex)
	this.Mapping("GetAuth", this.GetAuth)
	this.Mapping("GetDashboard", this.GetDashboard)
	this.Mapping("GetSetting", this.GetSetting)
	this.Mapping("GetRepository", this.GetRepository)
	this.Mapping("GetAdmin", this.GetAdmin)
	this.Mapping("GetAdminAuth", this.GetAdminAuth)
	this.Mapping("GetSignout", this.GetSignout)
}

func (this *WebController) Prepare() {
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *WebController) GetIndex() {
	this.TplNames = "index.html"
	this.Render()
	return
}

func (this *WebController) GetAuth() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		this.TplNames = "auth.html"
		this.Render()

		return
	} else {
		this.Ctx.Redirect(http.StatusMovedPermanently, "/dashboard")
	}
}

func (this *WebController) GetDashboard() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		beego.Error("[WEB API] Load session failure")
		this.Ctx.Redirect(http.StatusMovedPermanently, "/auth")

		return
	} else {
		this.TplNames = "dashboard.html"
		this.Data["username"] = user.Username

		this.Render()
		return
	}
}

func (this *WebController) GetSetting() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		beego.Error("[WEB API] Load session failure")
		this.Ctx.Redirect(http.StatusMovedPermanently, "/auth")

		return
	} else {
		this.TplNames = "setting.html"
		this.Data["username"] = user.Username

		this.Render()
		return
	}
}

func (this *WebController) GetRepository() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == true {
		this.Data["username"] = user.Username
	}

	this.TplNames = "repository.html"

	this.Render()
	return
}

func (this *WebController) GetAdmin() {
	this.TplNames = "admin.html"

	this.Data["username"] = "genedna"

	this.Render()
	return
}

func (this *WebController) GetAdminAuth() {
	this.TplNames = "admin-auth.html"

	this.Render()
	return
}

func (this *WebController) GetSignout() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		this.TplNames = "auth.html"
		this.Render()

		return
	} else {
		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := user.Log(models.ACTION_SINGOUT, models.LEVELINFORMATIONAL, models.TYPE_WEB, user.UUID, memo); err != nil {
			beego.Error("[WEB] Log Erro:", err.Error())
		}

		this.Ctx.Input.CruSession.Delete("user")
		this.Ctx.Redirect(http.StatusMovedPermanently, "/auth")
	}
}
