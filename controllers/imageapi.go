package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
	//相应 docker api 命令的 Controller 屏蔽 beego 的 XSRF ，避免错误。
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))

	if beego.AppConfig.String("docker::Standalone") == "true" {
		//单机运行模式，检查 Basic Auth 的认证。
		if len(this.Ctx.Input.Header("Authorization")) == 0 {
			//没有 Basic Auth 的认证，返回错误信息。
			this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Unauthorized\"}"))
			this.StopRun()
		} else {
			//Standalone True 模式，检查是否 Basic
			if strings.Index(this.Ctx.Input.Header("Authorization"), "Basic") == -1 {
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Unauthorized\"}"))
				this.StopRun()
			}

			//Decode Basic Auth 进行用户的判断
			username, passwd, err := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization"))

			beego.Trace("username: " + username)
			beego.Trace("password: " + passwd)

			if err != nil {
				beego.Error("[Decode Authoriztion Header] " + this.Ctx.Input.Header("Authorization") + " " + " error: " + err.Error())
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Unauthorized\"}"))
				this.StopRun()
			}

			user := new(models.User)
			has, err := user.Get(username, passwd, true)
			if err != nil {
				//查询用户数据失败，返回 401 错误
				beego.Error("[Search User] " + username + " " + passwd + " has error: " + err.Error())
				this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
				this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Unauthorized\"}"))
				this.StopRun()
			}

			if has == true {
				//查询到用户数据，在以下的 Action 处理函数中使用 this.Data["user"]
				//TODO 这里需要根据数据库特点改为存储 Key 么？
				this.Data["user"] = user
			} else {
				//没有查询到用户数据
				this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
				this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"User is not exist or not actived.\"}"))
				this.StopRun()
			}
		}
	} else {
		this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Service only support in standalone model.\"}"))
		this.StopRun()
	}
}

//在 Push 的流程中，docker 客户端会先调用 GET /v1/images/:image_id/json 向服务器检查是否已经存在 JSON 信息。
//如果存在了 JSON 信息，docker 客户端就认为是已经存在了 layer 数据，不再向服务器 PUT layer 的 JSON 信息和文件了。
//如果不存在 JSON 信息，docker 客户端会先后执行 PUT /v1/images/:image_id/json 和 PUT /v1/images/:image_id/layer 。
func (this *ImageAPIController) GetImageJSON() {

	if this.GetSession("access") == "write" || this.GetSession("access") == "read" {
		//TODO 检查 imageID 的合法性
		imageId := string(this.Ctx.Input.Param(":image_id"))

		image := new(models.Image)
		has, err := image.GetPushed(imageId, true, true)
		if err != nil {
			beego.Error("[Search Image] " + imageId + " " + " search error: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Check the image error.\"}"))
			this.StopRun()
		}

		if has == true {
			this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
			this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Checksum", image.Checksum)
			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.Ctx.Output.Context.Output.Body([]byte(image.JSON))
			this.StopRun()
		} else {
			this.Ctx.Output.Context.Output.SetStatus(http.StatusNotFound)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"No image json.\"}"))
			this.StopRun()
		}

	} else {
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Unauthorized.\"}"))
		this.StopRun()
	}
}

//向数据库写入 Layer 的 JSON 数据
//TODO: 检查 JSON 是否合法
func (this *ImageAPIController) PutImageJson() {
	if this.GetSession("access") == "write" {
		//判断是否存在 image 的数据, 新建或更改 JSON 数据
		imageId := this.Ctx.Input.Param(":image_id")

		image := new(models.Image)
		has, err := image.Get(imageId)
		if err != nil {
			beego.Error("[Search Image] " + imageId + " " + " search error: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Select image record error.\"}"))
			this.StopRun()
		}

		//TODO: 检查 JSON 是否合法
		//TODO: 检查 JSON 的逻辑性是否合法
		json := string(this.Ctx.Input.CopyBody())

		if has == true {
			_, err := image.UpdateJSON(json)
			if err != nil {
				beego.Error("[Update Image] " + imageId + " " + " update json error: " + err.Error())
				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Update the image JSON data error.\"}"))
				this.StopRun()
			}
		} else {
			_, err := image.Insert(imageId, json)
			if err != nil {
				beego.Error("[Update Image] " + imageId + " " + " create image record error: " + err.Error())
				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Create the image record error.\"}"))
				this.StopRun()
			}
		}

		this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.Ctx.Output.Context.Output.Body([]byte(""))
	}
}

//向本地硬盘写入 Layer 的文件
func (this *ImageAPIController) PutImageLayer() {
	if this.GetSession("access") == "write" {
		//查询是否存在 image 的数据库记录
		imageId := string(this.Ctx.Input.Param(":image_id"))

		image := new(models.Image)
		has, err := image.Get(imageId)
		if err != nil {
			beego.Error("[Search Image] " + imageId + " " + " search error: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Select image record error.\"}"))
			this.StopRun()
		}

		//写入磁盘的时候，如果数据库中没有对应的 image 数据报错。
		if has == false {
			beego.Error("[Search Image] " + imageId + " " + " search has none.")
			this.Ctx.Output.Context.Output.SetStatus(http.StatusNotFound)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Select image record error.\"}"))
			this.StopRun()
		}

		//TODO 保存文件的磁盘路径调度

		//处理 Layer 文件保存的目录
		basePath := beego.AppConfig.String("docker::BasePath")
		repositoryPath := fmt.Sprintf("%v/images/%v", basePath, imageId)
		layerfile := fmt.Sprintf("%v/images/%v/layer", basePath, imageId)

		if !utils.IsDirExists(repositoryPath) {
			os.MkdirAll(repositoryPath, os.ModePerm)
		}

		//如果存在了文件就移除文件
		if _, err := os.Stat(layerfile); err == nil {
			os.Remove(layerfile)
		}

		//写入 Layer 文件
		//TODO 超大的文件占内存，影响并发的情况。
		data, _ := ioutil.ReadAll(this.Ctx.Request.Body)

		err = ioutil.WriteFile(layerfile, data, 0777)
		if err != nil {
			beego.Error("[Put Image Layer] " + imageId + " " + " 写入磁盘错误: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Write image error.\"}"))
			this.StopRun()
		}

		//更新 Image 记录的 Uploaded
		_, err = image.UpdateUploaded(true)
		if err != nil {
			beego.Error("[Put Image Layer] " + imageId + " " + " 更新 Image Uploaded 标志错误: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\": \"Update the image upload status error.\"}"))
			this.StopRun()
		}

		//更新 Image 的 Size
		_, err = image.UpdateSize(int64(len(data)))
		if err != nil {
			beego.Error("[Put Image Layer] " + imageId + " " + " 更新 Image Size 错误: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\": \"Update the image size error.\"}"))
			this.StopRun()
		}

		//成功则返回 200
		this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.Ctx.Output.Context.Output.Body([]byte(""))
	}
}

func (this *ImageAPIController) PutChecksum() {

	if this.GetSession("access") == "write" {

		beego.Trace("Cookie: " + this.Ctx.Input.Header("Cookie"))
		beego.Trace("X-Docker-Checksum: " + this.Ctx.Input.Header("X-Docker-Checksum"))
		beego.Trace("X-Docker-Checksum-Payload: " + this.Ctx.Input.Header("X-Docker-Checksum-Payload"))

		//将 checksum 的值保存到数据库

		imageId := string(this.Ctx.Input.Param(":image_id"))

		image := new(models.Image)
		has, err := image.Get(imageId)
		if err != nil {
			beego.Error("[Search Image] " + imageId + " " + " search error: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Select image record error.\"}"))
			this.StopRun()
		}

		//在 Checksumed 的时候找不到 Image 的数据就进行报错。
		if has == false {
			beego.Error("[Search Image] " + imageId + " " + " search has none.")
			this.Ctx.Output.Context.Output.SetStatus(http.StatusNotFound)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Select image record error.\"}"))
			this.StopRun()
		}

		//TODO 检查上传的 Layer 文件的 SHA256 和传上来的 Checksum 的值是否一致？

		//更新 Checksumed 的记录
		_, err = image.UpdateChecksumed(true)
		if err != nil {
			beego.Error("[Put Image Checksum] " + imageId + " " + " 更新 Image Checksumed 标志错误: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Update the image checksum error.\"}"))
			this.StopRun()
		}

		_, err = image.UpdateChecksum(this.Ctx.Input.Header("X-Docker-Checksum"))
		if err != nil {
			beego.Error("[Put Image Checksum] " + imageId + " " + " 更新 Image Checksum 错误: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Update the image checksum error.\"}"))
			this.StopRun()
		}

		_, err = image.UpdatePayload(this.Ctx.Input.Header("X-Docker-Checksum-Payload"))
		if err != nil {
			beego.Error("[Put Image Checksum] " + imageId + " " + " 更新 Image Payload 标志错误: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Update the image checksum error.\"}"))
			this.StopRun()
		}

		_, err = image.UpdateParentJSON()
		if err != nil {
			beego.Error("[Put Image Checksum] " + imageId + " " + " 更新 Image Parent 错误: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Update the image checksum error.\"}"))
			this.StopRun()
		}

		this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.Ctx.Output.Context.Output.Body([]byte(""))
	}
}

func (this *ImageAPIController) GetImageAncestry() {
	if this.GetSession("access") == "read" {
		imageId := string(this.Ctx.Input.Param(":image_id"))

		image := new(models.Image)
		has, err := image.Get(imageId)
		if err != nil {
			beego.Error("[Search Image] " + imageId + " " + " search error: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Select image record error.\"}"))
			this.StopRun()
		}

		if has == false {
			beego.Error("[Search Image] " + imageId + " " + " search has none.")
			this.Ctx.Output.Context.Output.SetStatus(http.StatusNotFound)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Select image record error.\"}"))
			this.StopRun()
		} else {
			this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.Ctx.Output.Context.Output.Body([]byte(image.ParentJSON))
		}
	}
}

func (this *ImageAPIController) GetImageLayer() {

	if this.GetSession("access") == "read" {
		imageId := string(this.Ctx.Input.Param(":image_id"))

		image := new(models.Image)
		has, err := image.Get(imageId)
		if err != nil {
			beego.Error("[Search Image] " + imageId + " " + " search error: " + err.Error())
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Select image record error.\"}"))
			this.StopRun()
		}

		if has == false {
			beego.Error("[Search Image] " + imageId + " " + " search has none.")
			this.Ctx.Output.Context.Output.SetStatus(http.StatusNotFound)
			this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Select image record error.\"}"))
			this.StopRun()
		} else {
			//TODO 根据 private 的情况处理是从 CDN 下载还是七牛、又拍下载。

			//处理 Layer 文件保存的目录
			basePath := beego.AppConfig.String("docker::BasePath")
			layerfile := fmt.Sprintf("%v/images/%v/layer", basePath, imageId)

			if _, err := os.Stat(layerfile); err != nil {
				beego.Error("[Get Image Layer] " + imageId + " " + " has error: " + err.Error())
				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Could not find image file.\"}"))
				this.StopRun()
			}

			file, err := ioutil.ReadFile(layerfile)
			if err != nil {
				beego.Error("[Get Image Layer] " + imageId + " " + " has error: " + err.Error())
				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.Ctx.Output.Context.Output.Body([]byte("{\"error\":\"Load layer file error.\"}"))
				this.StopRun()
			}

			this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.Ctx.Output.Context.Output.Body(file)

		}
	}
}
