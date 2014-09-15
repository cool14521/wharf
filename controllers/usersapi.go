/*
Docker Registry & Login
执行 docker login 命令流程：
    1. docker 向 registry 的服务器进行注册执行：POST /v1/users or /v1/users/ -> POSTUsers()
    2. 创建用户成功返回 201；提交的格式有误、无效的字段等返回 400；已经存在用户了返回 401。
    3. docker login 收到 401 的状态后，进行登录：GET /v1/users or /v1/users/ -> GETUsers()
    4. 在登录时，将用户名和密码进行 SetBasicAuth 处理，放到 HEADER 的 Authorization 中，例如：Authorization: Basic ZnNrOmZzaw==
    5. registry 收到登录的请求，Decode 请求 HEADER 中 Authorization 的部分进行判断。
    6. 用户名和密码正确返回 200；用户名密码错误返回 401；账户未激活返回 403 错误；其它错误返回 417 (Expectation Failed)
注：
    Decode HEADER authorization function named decodeAuth in https://github.com/dotcloud/docker/blob/master/registry/auth.go.
更新 Docker Registry User 的属性：
    1. 调用 PUT /v1/users/(username)/ 向服务器更新 User 的 Email 和 Password 属性。
    2. 参数包括 User Email 或 User Password，或两者都包括。
    3. 更新成功返回 204；传递的参数不是有效的 JSON 格式等错误返回 400；认证失败返回 401；用户没有激活返回 403；没有用户现实 404。
注：
    HTTP HEADER authorization decode 验证同 docker login 命令。
*/
package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dockercn/docker-bucket/models"
	"github.com/dockercn/docker-bucket/utils"
	"net/http"
)

type UsersAPIController struct {
	beego.Controller
}

func (this *UsersAPIController) Prepare() {
	this.EnableXSRF = false
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
}

func (this *UsersAPIController) PostUsers() {

	beego.Trace("Authorization:" + this.Ctx.Input.Header("Authorization"))
	//TODO 检查配置文件是否可以在命令行注册的设置进行不同的处理。
	openSignup, _ := beego.AppConfig.Bool("docker::OpenSignup")
	if openSignup {
		//此处需要抓取允许注册时候的协议

		//获得用户提交的登陆(注册)信息
		var createUserJson map[string]interface{}
		json.Unmarshal(this.Ctx.Input.CopyBody(), &createUserJson)

		user := new(models.User)
		//查看用户是否已经注册过：这里取出Username只是为了证明用户存在
		userName, uerInfoErr := user.GetUserInfo(createUserJson["username"].(string), "Username")

		if uerInfoErr == nil && len(userName) > 0 {
			this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
			this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\": \"We are not support create a account from cli.\"}"))
			return
		} else {
			user.CreateUser(createUserJson["username"].(string), createUserJson["password"].(string), createUserJson["email"].(string))
			this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
			this.Ctx.Output.Context.Output.SetStatus(http.StatusCreated)
			this.Ctx.Output.Context.Output.Body([]byte("{\"info\": \"create a account success.\"}"))
			return
		}
	} else {
		this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\": \"We are not support create a account from cli.\"}"))
	}
}

func (this *UsersAPIController) GetUsers() {

	username, passwd, err := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization"))
	if err != nil {
		beego.Error("[Decode Authoriztion Header] " + this.Ctx.Input.Header("Authorization") + " " + " error: " + err.Error())
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized) //根据 Specification ，解码 Basic Authorization 数据失败也认为是 401 错误。
		this.Ctx.Output.Body([]byte("{\"error\":\"Unauthorized\"}"))
		this.StopRun()
	}

	beego.Trace("username: " + username)
	beego.Trace("password: " + passwd)

	user := new(models.User)
	has, err := user.Get(username, passwd, true)

	if err != nil {
		//查询用户数据失败，返回 401 错误
		beego.Error("[Search User] " + username + " " + passwd + " has error: " + err.Error())
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Body([]byte("{\"error\":\"Unauthorized\"}"))
		this.StopRun()
	}

	if has == false {
		//没有查询到用户数据
		this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"User is not exist or not actived.\"}"))
		this.StopRun()
	}

	//这句没明白记录什么的
	user.History(0, user.Id, fmt.Sprintf("%s %s %s", models.FROM_CLI, models.ACTION_SIGNIN, this.Ctx.Input.Header("X-Real-IP")))

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("{\"OK\"}"))
}
