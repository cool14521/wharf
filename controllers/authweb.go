package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dockercn/docker-bucket/models"
	"net/http"
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

	beego.Debug(fmt.Sprintf("[Web 用户] 用户登陆: %s", string(this.Ctx.Input.CopyBody())))
	beego.Debug(fmt.Sprintf("[Web 用户] 用户登陆: %s", u["username"].(string)))
	//验证用户登陆
	user := new(models.User)
	if has, err := user.Get(fmt.Sprint(u["username"]), fmt.Sprint(u["passwd"])); err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] 登录查询错误: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"用户登陆失败\"}"))
		return
	} else if !has {
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"用户名或密码不存在\"}"))
		return
	}

	//写入session中
	this.SetSession("user", *user)

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"登录成功\"}"))
	return
}

func (this *AuthWebController) ResetPasswd() {
	var u map[string]interface{}
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &u); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 解码用户重置密码发送的 JSON 数据失败: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"解码用户发送的 JSON 数据失败\"}"))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("[Web 用户] 用户重置密码: %s", string(this.Ctx.Input.CopyBody())))

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"发送重置密码邮件成功\"}"))
	return
}

func (this *AuthWebController) Signup() {
	var u map[string]interface{}
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &u); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 解码用户注册发送的 JSON 数据失败: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"解码用户发送的 JSON 数据失败\"}"))
		this.StopRun()
	}
	beego.Debug(fmt.Sprintf("[Web 用户] 用户注册: %s", string(this.Ctx.Input.CopyBody())))
	//判断用户是否存在，存在返回错误；不存在创建用户数据
	user := new(models.User)
	if err := user.Put(fmt.Sprint(u["username"]), fmt.Sprint(u["password"]), fmt.Sprint(u["email"])); err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] 注册用户错误: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"用户注册失败\"}"))
		return
	}
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"用户注册成功\"}"))
	return
}

func (this *AuthWebController) Signout() {
	this.DelSession("user")
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"用户退出成功\"}"))
	return
}
