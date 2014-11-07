package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
)

type AuthWebController struct {
	beego.Controller
}

func (this *AuthWebController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)

	//设置 Response 的 Header 信息，在处理函数中可以覆盖
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *AuthWebController) Signin() {
	//获得用户提交的登陆(注册)信息
	var u map[string]interface{}
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &u); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 解码用户注册发送的 JSON 数据失败: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"解码用户发送的 JSON 数据失败\"}"))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("[Web 用户] 用户登陆数据: %s", string(this.Ctx.Input.CopyBody())))
	beego.Debug(fmt.Sprintf("[Web 用户] 用户登陆数据: %s", u["email"].(string)))

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"登录成功\"}"))
}

func (this *AuthWebController) ResetPasswd() {

}

func (this *AuthWebController) Signup() {

}
