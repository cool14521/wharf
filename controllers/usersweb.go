package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/dockercn/wharf/models"
)

type UsersWebController struct {
	beego.Controller
}

func (this *UsersWebController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *UsersWebController) GetProfile() {
	user, ok := this.Ctx.Input.CruSession.Get("user").(models.User)
	if !ok {
		beego.Error(fmt.Sprintf("[WEB 用户] session加载失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s","url":"/auth"}`, "session加载失败")))
		return
	}
	user2json, err := json.Marshal(user)
	if err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] session解码json失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s","url":"/auth"}`, err.Error)))
		return
	}
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(user2json)
	return
}

func (this *UsersWebController) GetUserExist() {
	users := make([]models.User, 0)

	user := new(models.User)
	if _, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		beego.Error(fmt.Sprintf("err=,", err))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "用户查询不存在")))
		return
	}

	users = append(users, *user)

	users4Json, err := json.Marshal(users)
	if err != nil {
		beego.Error(fmt.Sprintf("用户列表json序列化失败，err=,", err))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "用户列表json序列化失败")))
	}
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(users4Json)
	return
}
