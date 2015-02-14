package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type AuthWebController struct {
	beego.Controller
}

func (this *AuthWebController) Prepare() {
	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *AuthWebController) Signup() {
	var user models.User
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &user); err != nil {
		beego.Error(fmt.Sprintf(`{"error":"%s"`, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error)))
		this.StopRun()
	}
	beego.Debug(fmt.Sprintf("[Web 用户] 用户注册: %s", string(this.Ctx.Input.CopyBody())))
	//判断用户是否存在，存在返回错误
	if has, _, err := user.Has(user.Username); err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] 注册用户错误: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error())))
		return
	} else if has {
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "用户名已经存在，请重新注册！")))
		return
	}
	//生成UUID
	user.UUID = string(utils.GeneralKey(user.Username))

	if err := user.Save(); err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] 存入ledis错误: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error())))
		return
	}
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "用户注册成功！")))
	return
}

//用户登录处理逻辑
func (this *AuthWebController) Signin() {
	//获得用户提交的登陆(注册)信息
	var user models.User
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &user); err != nil {
		beego.Error(fmt.Sprintf(`{"error":"%s"`, err.Error))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error)))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("[Web 用户] 用户登陆: %s", string(this.Ctx.Input.CopyBody())))
	//验证用户登陆
	if err := user.Get(user.Username, user.Password); err != nil {
		fmt.Println(err)
		beego.Error(fmt.Sprintf(`{"error":"%s"`, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error())))
		return
	}
	//处理用户头像
	if user.Gravatar == "" {
		user.Gravatar = "/static/images/default_user.jpg"
	}

	//将user信息写入session中
	this.Ctx.Input.CruSession.Set("user", user)

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"登录成功\"}"))
	return
}
