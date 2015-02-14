package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type UserWebAPIV1Controller struct {
	beego.Controller
}

func (u *UserWebAPIV1Controller) URLMapping() {
	u.Mapping("GetProfile", u.GetProfile)
	u.Mapping("GetUserExist", u.GetUserExist)
}

func (this *UserWebAPIV1Controller) Prepare() {
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *UserWebAPIV1Controller) GetProfile() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()

	} else {

		this.Data["json"] = user

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()

	}
}

func (this *UserWebAPIV1Controller) GetUserExist() {
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

func (this *UserWebAPIV1Controller) Signup() {
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
func (this *UserWebAPIV1Controller) Signin() {
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
