package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/modules"
	"github.com/dockercn/wharf/utils"
)

type ImageAPIV1Controller struct {
	beego.Controller
}

func (i *ImageAPIV1Controller) URLMapping() {
	i.Mapping("GetImageJSON", i.GetImageJSON)
	i.Mapping("PutImageJSON", i.PutImageJSON)
	i.Mapping("PutImageLayer", i.PutImageLayer)
	i.Mapping("PutChecksum", i.PutChecksum)
	i.Mapping("GetImageAncestry", i.GetImageAncestry)
	i.Mapping("GetImageLayer", i.GetImageLayer)
}

func (this *ImageAPIV1Controller) Prepare() {
	beego.Debug("[Header]")
	beego.Debug(this.Ctx.Request.Header)

	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))

}

func (this *ImageAPIV1Controller) GetImageJSON() {

	isAuth, errCode, errInfo := modules.DoAuthGetImageJSON(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

	sign := ""
	if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
	}
	beego.Info(fmt.Sprintf("[API 用户] sign:::%s", sign))

	imageId := string(this.Ctx.Input.Param(":image_id"))
	image := new(models.Image)

	var json []byte
	var checksum string
	var err error
	//获取 Image 的 JSON 和 Checksum 数据返回给 Docker 命令
	if json, err = image.GetJSON(imageId); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 查询 Image JSON %s 时报错 ", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"搜索 Image 的 JSON 数据错误\"}"))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("[%s JSON] %s", imageId, json))
	beego.Debug(fmt.Sprintf("[%s Checksum] %s", imageId, checksum))

	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Checksum", checksum)
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(json)
	this.StopRun()
}

func (this *ImageAPIV1Controller) PutImageJSON() {

	isAuth, errCode, errInfo := modules.DoAuthPutImageJSON(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

	imageId := this.Ctx.Input.Param(":image_id")

	sign := ""
	if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
	}
	beego.Error("sign:::", sign)

	image := new(models.Image)

	json := string(this.Ctx.Input.CopyBody())

	if err := image.PutJSON(imageId, json); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 向数据库写入 Image  [%s] 的 JSON [%s] 信息错误: %s"), imageId, json, err.Error())
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"向数据库写入 Image 的 JSON 数据错误\"}"))
		this.StopRun()
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))

}

func (this *ImageAPIV1Controller) PutImageLayer() {
	isAuth, errCode, errInfo := modules.DoAuthPutImageLayer(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

	imageId := string(this.Ctx.Input.Param(":image_id"))

	sign := ""
	if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
	}

	image := new(models.Image)

	basePath := beego.AppConfig.String("docker::BasePath")
	imagePath := fmt.Sprintf("%v/images/%v", basePath, imageId)
	layerfile := fmt.Sprintf("%v/images/%v/layer", basePath, imageId)

	if len(sign) > 0 {
		layerfile = fmt.Sprintf("%s-%s", layerfile, sign)
	}

	if !utils.IsDirExists(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}

	if _, err := os.Stat(layerfile); err == nil {
		os.Remove(layerfile)
	}

	data, _ := ioutil.ReadAll(this.Ctx.Request.Body)

	beego.Error(fmt.Sprintf("[API 用户] 上传Layer 大小"), len(data))

	if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 文件写入磁盘错误: %s ", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"文件写入磁盘错误\"}"))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("Image [%s] 文件本地存储全路径: %s", imageId, layerfile))

	if err := image.PutLayer(imageId, layerfile, true, int64(len(data))); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 更新 Image Layer 本地存储路径标志错误: %s ", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\": \"更新 Image Layer 本地存储路径错误\"}"))
		this.StopRun()
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))

}

func (this *ImageAPIV1Controller) PutChecksum() {
	isAuth, errCode, errInfo := modules.DoAuthPutChecksum(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

	beego.Debug("[Cookie] " + this.Ctx.Input.Header("Cookie"))
	beego.Debug("[X-Docker-Checksum] " + this.Ctx.Input.Header("X-Docker-Checksum"))
	beego.Debug("[X-Docker-Checksum-Payload] " + this.Ctx.Input.Header("X-Docker-Checksum-Payload"))

	sign := ""
	if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
	}

	beego.Debug("sign:::", sign)

	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	if err := image.PutChecksum(imageId, this.Ctx.Input.Header("X-Docker-Checksum"), true, this.Ctx.Input.Header("X-Docker-Checksum-Payload")); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 更新 Image Checksum错误: %s ", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Image Checksum错误\"}"))
		this.StopRun()
	}

	if err := image.PutAncestry(imageId); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 更新 Image Ancestry 错误: %s ", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"更新 Image Ancestry 错误\"}"))
		this.StopRun()
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))

}

func (this *ImageAPIV1Controller) GetImageAncestry() {
	isAuth, errCode, errInfo := modules.DoAuthGetImageAncestry(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	isHas, _, err := image.Has(imageId)

	if err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 读取 Ancestry 数据错误: %s ", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 Ancestry 数据错误\"}"))
		this.StopRun()
	}

	if !isHas {
		beego.Error(fmt.Sprintf("[API 用户] %s 读取 Ancestry 数据错误，没有找到Image", imageId))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(`{"错误":"读取 Ancestry 数据错误，没有找到Image"}`))
		this.StopRun()
	}
	beego.Debug(fmt.Sprintf("[%s Ancestry] %s", imageId, image.Ancestry))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(image.Ancestry))

}

func (this *ImageAPIV1Controller) GetImageLayer() {

	isAuth, errCode, errInfo := modules.DoAuthGetImageLayer(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	isHas, _, err := image.Has(imageId)
	if err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 读取 Layer 数据错误: %s ", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 Image Layer 数据错误\"}"))
		this.StopRun()

	}
	if !isHas {
		beego.Error(fmt.Sprintf("[API 用户] images 不存在 %s ", imageId))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(`{"错误":" Image 不存在"}`))
		this.StopRun()
	}

	layerfile := image.Path

	beego.Debug(fmt.Sprintf("[Image 本地存储路径] %s ", layerfile))

	if _, err := os.Stat(layerfile); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 读取 Layer 文件状态错误：%s", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 Image Layer 文件状态错误\"}"))
		this.StopRun()
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 读取 Layer 文件错误：%s", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"读取 Image Layer 文件错误\"}"))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("[Image 文件大小] %d", int64(len(file))))

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Transfer-Encoding", "binary")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Length", string(int64(len(file))))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(file)

}
