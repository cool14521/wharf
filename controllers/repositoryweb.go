package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dockercn/wharf/models"
	"net/http"
)

func (this *RepositoryController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

type RepositoryController struct {
	beego.Controller
}

func (this RepositoryController) GetRepositoryAdd() {
	//加载session
	user, ok := this.Ctx.Input.CruSession.Get("user").(models.User)
	if !ok {
		beego.Error(fmt.Sprintf("[WEB 用户] session加载失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "session加载失败")))
		return
	}
	this.TplNames = "repositoryAdd.html"
	this.Data["username"] = user.Username

	this.Render()
}
