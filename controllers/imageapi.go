package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/astaxie/beego"

	"github.com/dockercn/docker-bucket/models"
	"github.com/dockercn/docker-bucket/utils"
)

type ImageAPIController struct {
	beego.Controller
}

func (i *ImageAPIController) URLMapping() {
	i.Mapping("GetImageJSON", i.GetImageJSON)
	i.Mapping("PutImageJson", i.PutImageJson)
	i.Mapping("PutImageLayer", i.PutImageLayer)
	i.Mapping("PutChecksum", i.PutChecksum)
	i.Mapping("GetImageAncestry", i.GetImageAncestry)
	i.Mapping("GetImageLayer", i.GetImageLayer)
}

func (this *ImageAPIController) Prepare() {
	beego.Debug("[" + this.Ctx.Request.Method + "] " + this.Ctx.Request.URL.String())

	//相应 docker api 命令的 Controller 屏蔽 beego 的 XSRF ，避免错误。
	this.EnableXSRF = false

	//设置 Response 的 Header 信息，在处理函数中可以覆盖
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Endpoints", beego.AppConfig.String("docker::Endpoints"))

	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

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

			beego.Debug("[Token in Header]" + token)

			t := this.GetSession("token")

			//用 Header 中的 Token 和 Session 中得 Token 值进行比较，不相等返回错误退出执行
			if token != t {
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"HTTP Header 中的 Token 和 Session 的 Token 不同\"}"))
				this.StopRun()
			}

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
				if has, err := org.Has(namespace); err != nil {
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

//在 Push 的流程中，docker 客户端会先调用 GET /v1/images/:image_id/json 向服务器检查是否已经存在 JSON 信息。
//如果存在了 JSON 信息，docker 客户端就认为是已经存在了 layer 数据，不再向服务器 PUT layer 的 JSON 信息和文件了。
//如果不存在 JSON 信息，docker 客户端会先后执行 PUT /v1/images/:image_id/json 和 PUT /v1/images/:image_id/layer 。
func (this *ImageAPIController) GetImageJSON() {
	if this.GetSession("access") == "write" || this.GetSession("access") == "read" {
		//初始化加密签名
		sign := ""
		if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
			sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
		}

		//TODO 检查 imageID 的合法性
		imageId := string(this.Ctx.Input.Param(":image_id"))

		image := new(models.Image)
		has, err := image.GetPushed(imageId, sign, true, true)

		if err != nil {
			beego.Error(fmt.Sprintf("[API 用户] 查询 Image %s 时报错 ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"搜索 Image 错误\"}"))
			this.StopRun()
		}

		if has == true {
			var json []byte
			var checksum string
			//获取 Image 的 JSON 和 Checksum 数据返回给 Docker 命令
			if json, err = image.GetJSON(imageId, sign, true, true); err != nil {
				beego.Error(fmt.Sprintf("[API 用户] 查询 Image JSON %s 时报错 ", imageId, err.Error()))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"搜索 Image 的 JSON 数据错误\"}"))
				this.StopRun()
			}

			if checksum, err = image.GetChecksum(imageId, sign, true, true); err != nil {
				beego.Error(fmt.Sprintf("[API 用户] 查询 Image Checksum %s 时报错 ", imageId, err.Error()))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"搜索 Image 的 Checksum 数据错误\"}"))
				this.StopRun()
			}

			this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Checksum", checksum)
			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.Ctx.Output.Context.Output.Body(json)
			this.StopRun()

		} else {
			beego.Error(fmt.Sprintf("[API 用户] 没有查询到 Image ：%s ", imageId))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusNotFound)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"没有找到 Image 数据\"}"))
			this.StopRun()
		}
	} else {
		beego.Error("[API 用户] 查询 Image 时在 Session 中没有 write 或 read 的权限记录")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"没有访问 Image 数据的权限\"}"))
		this.StopRun()
	}
}

//向数据库写入 Layer 的 JSON 数据
func (this *ImageAPIController) PutImageJson() {
	if this.GetSession("access") == "write" {
		imageId := this.Ctx.Input.Param(":image_id")

		//初始化加密签名
		sign := ""
		if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
			sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
		}

		image := new(models.Image)

		//TODO: 检查 JSON 是否合法
		//TODO: 检查 JSON 的逻辑性是否合法
		json := string(this.Ctx.Input.CopyBody())

		if err := image.PutJSON(imageId, sign, json); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] 向数据库写入 Image  [%s] 的 JSON [%s] 信息错误: %s"), imageId, json, err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"向数据库写入 Image 的 JSON 数据错误\"}"))
			this.StopRun()

		}

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.Ctx.Output.Context.Output.Body([]byte(""))

	} else {
		beego.Error("[API 用户] 查询 Image 时在 Session 中没有 write 的权限记录")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"没有访问 Image 数据的写权限\"}"))
		this.StopRun()
	}
}

//向本地硬盘写入 Layer 的文件
func (this *ImageAPIController) PutImageLayer() {
	if this.GetSession("access") == "write" {
		imageId := string(this.Ctx.Input.Param(":image_id"))

		//初始化加密签名
		sign := ""
		if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
			sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
		}

		image := new(models.Image)

		//TODO 保存文件的磁盘路径调度

		//处理 Layer 文件保存的目录
		basePath := beego.AppConfig.String("docker::BasePath")
		imagePath := fmt.Sprintf("%v/images/%v", basePath, imageId)
		layerfile := fmt.Sprintf("%v/images/%v/layer", basePath, imageId)

		if len(sign) > 0 {
			layerfile = fmt.Sprintf("%s-%s", layerfile, sign)
		}

		//如果目录不存在，就创建目录
		if !utils.IsDirExists(imagePath) {
			os.MkdirAll(imagePath, os.ModePerm)
		}

		//如果存在了文件就移除文件
		if _, err := os.Stat(layerfile); err == nil {
			os.Remove(layerfile)
		}

		//写入 Layer 文件
		//TODO 超大的文件占内存，影响并发的情况。
		data, _ := ioutil.ReadAll(this.Ctx.Request.Body)

		if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] %s 文件写入磁盘错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"文件写入磁盘错误\"}"))
			this.StopRun()
		}

		//更新 Image 的文件本地存储路径
		if err := image.PutLocation(imageId, sign, layerfile); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] %s 更新 Image Layer 本地存储路径标志错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\": \"更新 Image Layer 本地存储路径标志错误\"}"))
			this.StopRun()
		}

		//更新 Image 的 Uploaded
		if err := image.PutUploaded(imageId, sign, true); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] %s 更新 Image 上传完成标志错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\": \"更新 Image 上传完成标志错误\"}"))
			this.StopRun()
		}

		//更新 Image 的 Size
		if err := image.PutSize(imageId, sign, int64(len(data))); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] %s 更新 Image Size 错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\": \"更新 Image Size 错误\"}"))
			this.StopRun()
		}

		//成功则返回 200
		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.Ctx.Output.Context.Output.Body([]byte(""))
	} else {
		beego.Error("[API 用户] 写入 Image Layer 时在 Session 中没有 write 的权限记录")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"没有写入 Image Layer 文件的权限\"}"))
		this.StopRun()
	}

}

func (this *ImageAPIController) PutChecksum() {
	if this.GetSession("access") == "write" {

		beego.Debug("[Cookie] " + this.Ctx.Input.Header("Cookie"))
		beego.Debug("[X-Docker-Checksum] " + this.Ctx.Input.Header("X-Docker-Checksum"))
		beego.Debug("[X-Docker-Checksum-Payload] " + this.Ctx.Input.Header("X-Docker-Checksum-Payload"))

		//初始化加密签名
		sign := ""
		if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
			sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
		}

		imageId := string(this.Ctx.Input.Param(":image_id"))

		image := new(models.Image)

		//TODO 检查上传的 Layer 文件的 SHA256 和传上来的 Checksum 的值是否一致？

		if err := image.PutChecksumed(imageId, sign, true); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] %s 更新 Image Checksumed 标志错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Image Checksumed 标志错误\"}"))
			this.StopRun()
		}

		if err := image.PutChecksum(imageId, sign, this.Ctx.Input.Header("X-Docker-Checksum")); err != nil {
			beego.Error(fmt.Sprint("[API 用户] %s 更新 Image Checksum 错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Image Checksum 错误\"}"))
			this.StopRun()
		}

		if err := image.PutPayload(imageId, sign, this.Ctx.Input.Header("X-Docker-Checksum-Payload")); err != nil {
			beego.Error(fmt.Sprint("[API 用户] %s 更新 Image Payload 标志错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Image Payload 标志错误\"}"))
			this.StopRun()
		}

		if err := image.PutAncestry(imageId, sign); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] %s 更新 Image Ancestry 错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Image Ancestry 错误\"}"))
			this.StopRun()
		}

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.Ctx.Output.Context.Output.Body([]byte(""))
	} else {
		beego.Error("[API 用户] 写入 Image Checksum 时在 Session 中没有 write 的权限记录")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"没有写入 Image Checksum 的权限\"}"))
		this.StopRun()
	}
}

func (this *ImageAPIController) GetImageAncestry() {
	if this.GetSession("access") == "read" {
		imageId := string(this.Ctx.Input.Param(":image_id"))

		//初始化加密签名
		sign := ""
		if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
			sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
		}

		image := new(models.Image)

		if ancestry, err := image.GetAncestry(imageId, sign, true, true); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] %s 读取 Ancestry 数据错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 Ancestry 数据错误\"}"))
			this.StopRun()
		} else {
			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.Ctx.Output.Context.Output.Body(ancestry)
		}
	} else {
		beego.Error("[API 用户] 读取 Image Ancestry 时在 Session 中没有 read 的权限记录")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"没有读取 Image Ancestry 的权限\"}"))
		this.StopRun()
	}
}

func (this *ImageAPIController) GetImageLayer() {
	if this.GetSession("access") == "read" {
		imageId := string(this.Ctx.Input.Param(":image_id"))

		//初始化加密签名
		sign := ""
		if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
			sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
		}

		image := new(models.Image)

		if layerfile, err := image.GetLocation(imageId, sign, true, true); err != nil {
			beego.Error(fmt.Sprintf("[API 用户] %s 读取 Layer 数据错误: %s ", imageId, err.Error()))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 Image Layer 数据错误\"}"))
			this.StopRun()
		} else {

			if _, err := os.Stat(layerfile); err != nil {
				beego.Error(fmt.Sprintf("[API 用户] %s 读取 Layer 文件状态错误：%s", imageId, err.Error()))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 Image Layer 文件状态错误\"}"))
				this.StopRun()
			}

			if file, err := ioutil.ReadFile(layerfile); err != nil {
				beego.Error(fmt.Sprintf("[API 用户] %s 读取 Layer 文件错误：%s", imageId, err.Error()))
				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 Image Layer 文件错误\"}"))
				this.StopRun()
			} else {
				//读取的文件放在 HTTP Body 中返回给 Docker 命令
				this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
				this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
				this.Ctx.Output.Context.Output.Body(file)
			}

		}
	} else {
		beego.Error("[API 用户] 读取 Image Layer 文件时在 Session 中没有 read 的权限记录")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"没有读取 Image Layer 文件的权限\"}"))
		this.StopRun()
	}
}
