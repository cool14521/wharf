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
	//"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type UsersAPIController struct {
	beego.Controller
}

func (u *UsersAPIController) URLMapping() {
	u.Mapping("PostUsers", u.PostUsers)
	u.Mapping("GetUsers", u.GetUsers)
}

func (this *UsersAPIController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)

	//相应 docker api 命令的 Controller 屏蔽 beego 的 XSRF ，避免错误。
	this.EnableXSRF = false

	//设置 Response 的 Header 信息，在处理函数中可以覆盖
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))
}

//返回 401 错误会让 docker 命令行执行 GET /v1/users 的登录动作。
//返回 400 和 403 会在 Docker 客户端抛出错误并终止。
//如果支持 docker 命令行创建账户，在创建成功后返回 201 状态吗。
func (this *UsersAPIController) PostUsers() {
	//根据配置文件中得是否可以注册处理逻辑
	this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
	this.Ctx.Output.Context.Output.Body([]byte("{\"错误\": \"不支持从 docker 命令行创建用户\"}"))
}

func (this *UsersAPIController) GetUsers() {

	username, passwd, err := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization"))
	if err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 用户登录的 Basic Auth 验证错误: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Body([]byte("{\"错误\":\"用户登录验证失败\"}"))
		this.StopRun()
	}

	user := new(models.User)
	err = user.Get(username, passwd)

	if err != nil {
		//查询用户数据失败，返回 401 错误
		beego.Error(fmt.Sprintf("[API 用户] 查询用户错误： ", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Body([]byte("{\"错误\":\"用户登录验证失败\"}"))
		this.StopRun()
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("{\"状态\":\"登录成功\"}"))
}
