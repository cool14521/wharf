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

	"github.com/dockercn/docker-bucket/global"
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
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)

	//相应 docker api 命令的 Controller 屏蔽 beego 的 XSRF ，避免错误。
	this.EnableXSRF = false

	//设置 Response 的 Header 信息，在处理函数中可以覆盖
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", global.BucketConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", global.BucketConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", global.BucketConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", global.BucketConfig.String("docker::Encrypt"))

	//检查 Authorization 的 Header 信息是否存在。
	if len(this.Ctx.Input.Header("Authorization")) == 0 {
		//不存在 Authorization 信息返回错误信息
		beego.Error("[API 用户] Docker 命令访问 HTTP API 的 Header 中没有 Authorization 信息: ")
		beego.Error(this.Ctx.Request.Header)

		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"没有找到 Authorization 的认证信息\"}"))
		this.StopRun()

	} else {

		beego.Debug("[Authorization] " + this.Ctx.Input.Header("Authorization"))

		//检查是否 Basic Auth
		if strings.Index(this.Ctx.Input.Header("Authorization"), "Basic") == -1 {

			//非 Basic Auth ，检查 Token
			if strings.Index(this.Ctx.Input.Header("Authorization"), "Token") == -1 {
				beego.Error("[API 用户] Docker 命令访问 HTTP API 的 Header 中没有 Basic Auth 和 Token 的信息 ")
				beego.Error(this.Ctx.Request.Header)

				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"在 HTTP Header Authorization 中没有找到 Basic Auth 和 Token 信息\"}"))
				this.StopRun()
			}

			//使用正则获取 Token 的值
			r, _ := regexp.Compile(`Token (?P<token>\w+)`)
			tokens := r.FindStringSubmatch(this.Ctx.Input.Header("Authorization"))
			_, token := tokens[0], tokens[1]

			beego.Debug("[Token in Header] " + token)

			t := this.GetSession("token")

			//用 Header 中的 Token 和 Session 中得 Token 值进行比较，不相等返回错误退出执行
			if token != t {
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"HTTP Header 中的 Token 和 Session 的 Token 不同\"}"))
				this.StopRun()
			}

			this.Data["username"] = this.GetSession("username")
			this.Data["org"] = this.GetSession("org")

		} else {
			//解码 Basic Auth 进行用户的判断
			username, passwd, err := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization"))

			if err != nil {
				beego.Error(fmt.Sprintf("[API 用户] 解码 Header Authorization 的 Basic Auth [%s] 时遇到错误： %s", this.Ctx.Input.Header("Authorization"), err.Error()))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"解码 Authorization 的 Basic Auth 信息错误\"}"))
				this.StopRun()
			}

			//根据解码的数据，在数据库中查询用户
			user := new(models.User)
			has, err := user.Get(username, passwd)
			if err != nil {
				//查询用户数据失败，返回 401 错误
				beego.Error(fmt.Sprintf("[API 用户] 在数据库中查询用户数据遇到错误：%s", err.Error()))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"在数据库中查询用户数据时出现数据库错误\"}"))
				this.StopRun()
			}

			if has == true {
				//查询到用户数据，在以下的 Action 处理函数中使用 this.Data["username"]
				this.Data["username"] = username
				this.Data["passwd"] = passwd

				//根据 Namespace 查询组织数据
				namespace := string(this.Ctx.Input.Param(":namespace"))
				org := new(models.Organization)
				if has, _, err := org.Has(namespace); err != nil {
					beego.Error(fmt.Sprintf("[API 用户] 查询组织名称 %s 时错误 %s", namespace, err.Error()))
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
				//没有查询到用户数据，返回 401 错误
				beego.Error(fmt.Sprintf("[API 用户] 没有查询到用户：%s ", username))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"没有查询到用户\"}"))
				this.StopRun()
			}
		}
	}

}

func (this *RepositoryAPIController) PutRepository() {
	username := this.Data["username"].(string)
	passwd := this.Data["passwd"].(string)
	org := this.Data["org"].(string)

	beego.Debug("[Username] " + username)
	beego.Debug("[Org] " + org)

	//获取namespace/repository
	namespace := string(this.Ctx.Input.Param(":namespace"))
	repository := string(this.Ctx.Input.Param(":repo_name"))

	//判断namespace和username的关系，处理权限的问题
	//TODO：组织的镜像仓库权限判断
	if namespace != username {
		beego.Error(fmt.Sprintf("[API 用户] 更新/新建 repository 数据错误, 用户名和命名空间不相同。"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新/新建镜像仓库数据错误\"}"))
		this.StopRun()
	}

	//加密签名
	sign := ""
	if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
	}

	beego.Debug("[Sign] " + sign)

	//创建或更新 Repository 数据
	//也可以采用 ioutil.ReadAll(this.Ctx.Request.Body) 的方式读取 body 数据
	//TODO 检查 JSON 字符串是否合法
	//TODO 检查 逻辑是否合法

	beego.Debug("[JSON] " + string(this.Ctx.Input.CopyBody()))

	//从 API 创建的 Repository 默认是 Public 的。
	repo := new(models.Repository)
	if err := repo.PutJSON(username, repository, org, sign, string(this.Ctx.Input.CopyBody())); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 更新/新建 repository 数据错误: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新/新建镜像仓库数据错误\"}"))
		this.StopRun()
	}

	if err := repo.PutAgent(username, repository, org, sign, this.Ctx.Input.Header("User-Agent")); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 更新 User Agent 数据错误: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 User Agent 数据错误\"}"))
		this.StopRun()
	}

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

	this.SetSession("username", username)
	this.SetSession("org", org)
	this.SetSession("namespace", namespace)
	this.SetSession("repository", repository)
	this.SetSession("access", "write")

	//操作正常的输出
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Endpoints", global.BucketConfig.String("docker::Endpoints"))

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte("\"\""))
}

func (this *RepositoryAPIController) PutTag() {
	if this.GetSession("access") == "write" {

		beego.Debug("[Namespace] " + this.Ctx.Input.Param(":namespace"))
		beego.Debug("[Repository] " + this.Ctx.Input.Param(":repo_name"))
		beego.Debug("[Tag] " + this.Ctx.Input.Param(":tag"))

		username := this.Data["username"].(string)
		org := this.Data["org"].(string)

		namespace := this.Ctx.Input.Param(":namespace")
		repository := this.Ctx.Input.Param(":repo_name")

		//加密签名
		sign := ""
		if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
			sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
		}

		beego.Debug("[Sign] " + sign)

		tag := this.Ctx.Input.Param(":tag")

		//从 HTTP Body 中获取 Image 的 Value
		r, _ := regexp.Compile(`"([[:alnum:]]+)"`)
		imageIds := r.FindStringSubmatch(string(this.Ctx.Input.CopyBody()))

		repo := new(models.Repository)
		if err := repo.PutTag(username, repository, org, sign, tag, imageIds[1]); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] 更新 %s/%s 的 Tag [%s:%s] 错误: %s", namespace, repository, tag, imageIds[1], err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Tag 数据错误\"}"))
			this.StopRun()
		}

		//操作正常的输出
		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.Ctx.Output.Context.Output.Body([]byte("\"\""))
	} else {
		beego.Error("[API 用户] 更新 Repository 的 Tag 信息时在 Session 中没有 write 的权限记录")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"没有更新 Repository Tag 数据的写权限\"}"))
		this.StopRun()
	}
}

//Push 命令的最后一步，所有的检查操作，通知操作都在此函数进行。
func (this *RepositoryAPIController) PutRepositoryImages() {
	if this.GetSession("access") == "write" {
		beego.Debug("[Namespace] " + this.Ctx.Input.Param(":namespace"))
		beego.Debug("[Repository] " + this.Ctx.Input.Param(":repo_name"))

		username := this.Data["username"].(string)
		org := this.Data["org"].(string)

		namespace := this.Ctx.Input.Param(":namespace")
		repository := this.Ctx.Input.Param(":repo_name")

		//加密签名
		sign := ""
		if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
			sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
		}

		beego.Debug("[Sign] " + sign)

		beego.Debug("[Body] " + string(this.Ctx.Input.CopyBody()))

		repo := new(models.Repository)

		//TODO 检查仓库所有镜像的 Tag 信息和上传的 Tag 信息是否一致。
		//TODO 检查仓库所有镜像是否 Uploaded 为 True
		//TODO 检查仓库所有镜像是否 Checksumed 为 True

		//设定 repository 的 Uploaded
		if err := repo.PutUploaded(username, repository, org, sign, true); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] 更新 %s/%s 的 Uploaded 标志错误: %s", namespace, repository, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Uploaded 标志错误\"}"))
			this.StopRun()
		}
		//设定 repository 的 Checksumed
		if err := repo.PutChecksumed(username, repository, org, sign, true); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] 更新 %s/%s 的 Checksumed 标志错误: %s", namespace, repository, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Checksumed 标志错误\"}"))
			this.StopRun()
		}
		//设定 repository 的 Size
		if err := repo.PutSize(username, repository, org, sign); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] 更新 %s/%s 的 Size 数据错误: %s", namespace, repository, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Size 数据错误\"}"))
			this.StopRun()
		}

		//操作正常的输出
		this.Ctx.Output.Context.Output.SetStatus(http.StatusNoContent)
		this.Ctx.Output.Context.Output.Body([]byte("\"\""))
	} else {
		beego.Error("[API 用户] 更新 Repository 的 Checksum 信息时在 Session 中没有 write 的权限记录")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"没有更新 Repository Checksum 数据的写权限\"}"))
		this.StopRun()
	}
}

//获取一个 Repository 的 Image 信息
func (this *RepositoryAPIController) GetRepositoryImages() {
	beego.Debug("[Namespace] " + this.Ctx.Input.Param(":namespace"))
	beego.Debug("[Repository] " + this.Ctx.Input.Param(":repo_name"))

	username := this.Data["username"].(string)
	org := this.Data["org"].(string)

	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")

	//TODO：私有和组织的镜像仓库权限判断问题

	//加密签名
	sign := ""
	if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
	}

	beego.Debug("[Sign] " + sign)

	//TODO 私有镜像仓库权限判断

	repo := new(models.Repository)
	if json, err := repo.GetJSON(namespace, repository, org, sign, true, true); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 读取 %s/%s 的 JSON 数据错误: %s", namespace, repository, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 JSON 数据错误\"}"))
		this.StopRun()
	} else {
		//如果 Request 的 Header 中含有 X-Docker-Token 且为 True，需要在返回值设置 Token 值。
		//否则客户端报错 Index response didn't contain an access token
		if this.Ctx.Input.Header("X-Docker-Token") == "true" {
			//创建 token 并保存
			//需要加密的字符串为 UserName + UserPassword + 时间戳
			token := utils.GeneralToken(username + repository)
			this.SetSession("token", token)
			//在返回值 Header 里面设置 Token
			this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Token", token)
			this.Ctx.Output.Context.ResponseWriter.Header().Set("WWW-Authenticate", token)
		}

		this.SetSession("username", username)
		this.SetSession("org", org)
		this.SetSession("namespace", namespace)
		this.SetSession("repository", repository)
		//在 SetSession 中增加读权限
		this.SetSession("access", "read")
		//操作正常的输出
		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.Ctx.Output.Context.Output.Body(json)
	}
}

func (this *RepositoryAPIController) GetRepositoryTags() {
	if this.GetSession("access") == "read" {
		beego.Debug("[Namespace] " + this.Ctx.Input.Param(":namespace"))
		beego.Debug("[Repository] " + this.Ctx.Input.Param(":repo_name"))

		//username := this.Data["username"].(string)
		org := this.Data["org"].(string)

		namespace := this.Ctx.Input.Param(":namespace")
		repository := this.Ctx.Input.Param(":repo_name")

		//加密签名
		sign := ""
		if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
			sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
		}

		beego.Debug("[Sign] " + sign)

		repo := new(models.Repository)
		if tags, err := repo.GetTags(namespace, repository, org, sign, true, true); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] 读取 %s/%s 的 Tags 数据错误: %s", namespace, repository, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 Tag 数据错误\"}"))
			this.StopRun()
		} else {
			//TODO 私有镜像仓库权限判断
			//操作正常的输出
			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.Ctx.Output.Context.Output.Body(tags)
		}
	} else {
		beego.Error("[API 用户] 读取 Repository Tag 时在 Session 中没有 read 的权限记录")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"没有读取 Repository Tag 的权限\"}"))
		this.StopRun()
	}
}
