/*
Docker Push & Pull

执行 docker push 命令流程：
    1. docker 向 registry 服务器注册 repository： PUT /v1/repositories/<username>/<repository> -> PUTRepository()
    2. 参数是 JSON 格式的 <repository> 所有 image 的 id 列表，按照 image 的构建顺序排列。
    3. 根据 <repository> 的 <tags> 进行循环：
       3.1 获取 <image> 的 JSON 文件：GET /v1/images/<image_id>/json -> image.go#GETJSON()
       3.2 如果没有此文件或内容返回 404 。
       3.3 docker push 认为服务器没有 image 对应的文件，向服务器上传 image 相关文件。
           3.3.1 写入 <image> 的 JSON 文件：PUT /v1/images/<image_id>/json -> image.go#PUTJSON()
           3.3.2 写入 <image> 的 layer 文件：PUT /v1/images/<image_id>/layer -> image.go#PUTLayer()
           3.3.3 写入 <image> 的 checksum 信息：PUT /v1/images/<image_id>/checksum -> image.go#PUTChecksum()
       3.4 上传完此 tag 的所有 image 后，向服务器写入 tag 信息：PUT /v1/repositories/(namespace)/(repository)/tags/(tag) -> PUTTag()
    4. 所有 tags 的 image 上传完成后，向服务器发送所有 images 的校验信息，PUT /v1/repositories/(namespace)/(repo_name)/images -> PUTRepositoryImages()

执行 docker pull 命令流程：
    1. docker 访问 registry 服务器 repository 的 images 信息：GET /v1/repositories/<username>/<repository>/images -> GetRepositoryImages()
    2. docker 访问 registry 服务器 repository 的 tags 信息：GET /v1/repositoies/<username>/<repository>/tags -> GetRepositoryTags()
    3. 根据 <repository> 的 <tags> 中 image 信息进行循环：
      3.1 获取 <image> 的 Ancestry 信息：GET /v1/images/<image_id>/ancestry -> GetImageAncestry()
      3.2 获取 <image> 的 JSON 数据: GET /v1/images/<image_id>/json -> GetImageJson()
      3.3 获取 <image> 的 Layer 文件: GET /v1/images/<image_id/layer -> GetImageLayer()

*/
package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/astaxie/beego"

	"github.com/dockercn/docker-bucket/models"
	"github.com/dockercn/docker-bucket/utils"
)

type RepositoryAPIController struct {
	beego.Controller
}

func (r *RepositoryAPIController) URLMapping() {
	r.Mapping("PutTag", r.PutTag)
	r.Mapping("PutRepositoryImages", r.PutRepositoryImages)
	r.Mapping("GetRepositoryImages", r.GetRepositoryImages)
	r.Mapping("GetRepositoryTags", r.GetRepositoryTags)
	r.Mapping("PutRepository", r.PutRepository)
}

func (this *RepositoryAPIController) Prepare() {
	//相应 docker api 命令的 Controller 屏蔽 beego 的 XSRF ，避免错误。
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))

	if beego.AppConfig.String("docker::Standalone") == "true" {
		//单机运行模式，检查 Basic Auth 的认证。
		if len(this.Ctx.Input.Header("Authorization")) == 0 {
			//没有 Basic Auth 的认证，返回错误信息。
			beego.Error("没有 Authorization 信息的 API 访问")
			this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"服务器只支持 Basic Auth 验证模式，请联系系统管理员\"}"))
			this.StopRun()
		} else {
			beego.Debug("Debug: " + this.Ctx.Input.Header("Authorization"))
			//Standalone True 模式，检查是否 Basic
			if strings.Index(this.Ctx.Input.Header("Authorization"), "Basic") == -1 {
				beego.Error("Authorization 中 Auth 的格式错误: " + this.Ctx.Input.Header("Authorization"))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"HTTP Header 的 Authorization 格式错误\"}"))
				this.StopRun()
			}

			//Decode Basic Auth 进行用户的判断
			username, passwd, err := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization"))
			if err != nil {
				beego.Error(fmt.Sprintf("[解码 Basic Auth] %s 错误： %s ", this.Ctx.Input.Header("Authorization"), err.Error()))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"解码 HTTP Header 的 Basic Auth 信息错误\"}"))
				this.StopRun()
			}

			//判断 Header 信息里面的用户数据是否存在
			user := new(models.User)
			has, err := user.Get(username, passwd, true)
			if err != nil {
				//查询用户数据失败，返回 401 错误
				beego.Error(fmt.Sprintf("[API 用户] 查询用户错误： ", err.Error()))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"查询用户错误\"}"))
				this.StopRun()
			}

			if has == true {
				//查询到用户数据，在以下的 Action 处理函数中使用 this.Data["username"]
				this.Data["username"] = username
				this.Data["passwd"] = passwd

				//根据 Namespace 查询组织数据
				namespace := string(this.Ctx.Input.Param(":namespace"))
				org := new(models.Organization)
				if has, err := org.Get(namespace, true); err != nil {
					beego.Error(fmt.Sprintf("查询组织名称 %s 时错误 %s", namespace, err.Error()))
					this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
					this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"查询组织数据报错。\"}"))
					this.StopRun()
				} else if has == false {
					this.Data["org"] = ""
				} else {
					//查询到组织数据，在以下的 Action 处理函数中使用 this.Data["org"]
					this.Data["org"] = namespace
				}
			} else {
				//查询用户数据失败，返回 401 错误
				beego.Error(fmt.Sprintf("[API 用户] 没有查询到用户：%s ", username))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"没有查询到用户\"}"))
				this.StopRun()
			}
		}
	} else {
		beego.Error("非 Standalone 模式登录尝试错误")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"服务器只支持 Basic Auth 验证模式，请联系系统管理员\"}"))
		this.StopRun()
	}

}

func (this *RepositoryAPIController) PutRepository() {
	username := this.Data["username"].(string)
	passwd := this.Data["passwd"].(string)
	org := this.Data["org"].(string)

	beego.Debug("Username: " + username)

	//获取namespace/repository
	namespace := string(this.Ctx.Input.Param(":namespace"))
	repository := string(this.Ctx.Input.Param(":repo_name"))

	//加密签名
	sign := ""
	if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
	}

	beego.Debug("Sign: " + sign)

	//创建或更新 Repository 数据
	//也可以采用 ioutil.ReadAll(this.Ctx.Request.Body) 的方式读取 body 数据
	//TODO 检查 JSON 字符串是否合法
	//TODO 检查 逻辑是否合法

	beego.Debug("JSON: " + string(this.Ctx.Input.CopyBody()))

	//从 API 创建的 Repository 默认是 Public 的。
	repo := new(models.Repository)
	if err := repo.Add(username, repository, org, sign, string(this.Ctx.Input.CopyBody())); err != nil {
		beego.Error("更新/新建 repository 数据错误")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新/新建 repository 数据错误\"}"))
		this.StopRun()
	}
	//repo.SetAgent(username, repository, org, this.Ctx.Input.Header("User-Agent"))

	//如果 Request 的 Header 中含有 X-Docker-Token 且为 True，需要在返回值设置 Token 值。
	//否则客户端报错 Index response didn't contain an access token
	if this.Ctx.Input.Header("X-Docker-Token") == "true" {
		//创建 token 并保存
		//需要加密的字符串为 UserName + UserPassword + 时间戳
		token := utils.GeneralToken(username + passwd)
		this.SetSession("token", token)
		//在返回值 Header 里面设置 Token
		this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Token", token)
		this.Ctx.Output.Context.ResponseWriter.Header().Set("WWW-Authenticate", token)
	}

	this.SetSession("namespace", namespace)
	this.SetSession("repository", repository)
	this.SetSession("access", "write")

	//操作正常的输出
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Endpoints", beego.AppConfig.String("docker::Endpoints"))

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("\"\""))
}

func (this *RepositoryAPIController) PutTag() {
	beego.Debug("Namespace: " + this.Ctx.Input.Param(":namespace"))
	beego.Debug("Repository: " + this.Ctx.Input.Param(":repo_name"))
	beego.Debug("Tag: " + this.Ctx.Input.Param(":tag"))

	username := this.Data["username"].(string)
	org := this.Data["org"].(string)

	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")
	//加密签名
	//sign := string(this.Ctx.Input.Header("X-Docker-Sign"))

	tag := this.Ctx.Input.Param(":tag")

	//从 HTTP Body 中获取 Image 的 Value
	r, _ := regexp.Compile(`"([[:alnum:]]+)"`)
	imageIds := r.FindStringSubmatch(string(this.Ctx.Input.CopyBody()))

	repo := new(models.Repository)
	if err := repo.SetTag(username, repository, org, tag, imageIds[1]); err != nil {
		beego.Error("[Update Tag] " + namespace + " " + repository + " " + tag + " error: " + err.Error())
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Update the tag data error.\"}"))
		this.StopRun()
	}

	//操作正常的输出
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("\"\""))
}

//Push 命令的最后一步，所有的检查操作，通知操作都在此函数进行。
func (this *RepositoryAPIController) PutRepositoryImages() {
	//username := this.Data["username"].(string)
	//org := this.Data["org"].(string)

	//获取namespace/repository
	//namespace := string(this.Ctx.Input.Param(":namespace"))
	//repository := string(this.Ctx.Input.Param(":repo_name"))
	//加密签名
	//sign := string(this.Ctx.Input.Header("X-Docker-Sign"))

	//repo := new(models.Repository)

	//TODO 计算 repository 的存储量
	//TODO 计算所有的 Image 是不是 UPloaded
	//TODO 计算所有的 Image 是不是 Checksumed

	//TODO 设定 repository 的 Uploaded
	//TODO 设定 repository 的 Checksumed
	//TODO 设定 repository 的 Size

	//操作正常的输出
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("\"\""))
}

//获取一个 Repository 的 Image 信息
func (this *RepositoryAPIController) GetRepositoryImages() {

}

func (this *RepositoryAPIController) GetRepositoryTags() {

}
