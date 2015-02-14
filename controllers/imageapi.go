package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/astaxie/beego"
	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type ImageAPIController struct {
	beego.Controller
}

func (i *ImageAPIController) URLMapping() {
	i.Mapping("GetImageJSON", i.GetImageJSON)
	i.Mapping("PutImageJSON", i.PutImageJSON)
	i.Mapping("PutImageLayer", i.PutImageLayer)
	i.Mapping("PutChecksum", i.PutChecksum)
	i.Mapping("GetImageAncestry", i.GetImageAncestry)
	i.Mapping("GetImageLayer", i.GetImageLayer)
}

func (this *ImageAPIController) Prepare() {
	beego.Debug("[Header]")
	beego.Debug(this.Ctx.Request.Header)

	//相应 docker api 命令的 Controller 屏蔽 beego 的 XSRF ，避免错误。
	this.EnableXSRF = false

	//设置 Response 的 Header 信息，在处理函数中可以覆盖
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))

}

//在 Push 的流程中，docker 客户端会先调用 GET /v1/images/:image_id/json 向服务器检查是否已经存在 JSON 信息。
//如果存在了 JSON 信息，docker 客户端就认为是已经存在了 layer 数据，不再向服务器 PUT layer 的 JSON 信息和文件了。
//如果不存在 JSON 信息，docker 客户端会先后执行 PUT /v1/images/:image_id/json 和 PUT /v1/images/:image_id/layer 。
func (this *ImageAPIController) GetImageJSON() {

	isAuth, errCode, errInfo := models.DoAuthGetImageJSON(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

	//初始化加密签名
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

//向数据库写入 Layer 的 JSON 数据
func (this *ImageAPIController) PutImageJSON() {

	//	beego.Error("Session Access:::", this.GetSession("access").(string))
	//	beego.Error("Ctx.Session Access:::", this.Ctx.Input.Session("access").(string))
	isAuth, errCode, errInfo := models.DoAuthPutImageJSON(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

	imageId := this.Ctx.Input.Param(":image_id")

	//初始化加密签名
	sign := ""
	if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
	}
	beego.Error("sign:::", sign)

	image := new(models.Image)

	//TODO: 检查 JSON 是否合法
	//TODO: 检查 JSON 的逻辑性是否合法
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

//向本地硬盘写入 Layer 的文件
func (this *ImageAPIController) PutImageLayer() {
	isAuth, errCode, errInfo := models.DoAuthPutImageLayer(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

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

	beego.Error(fmt.Sprintf("[API 用户] 上传Layer 大小"), len(data))

	if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 文件写入磁盘错误: %s ", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\":\"文件写入磁盘错误\"}"))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("Image [%s] 文件本地存储全路径: %s", imageId, layerfile))

	//更新 Image 的文件本地存储路径
	if err := image.PutLayer(imageId, layerfile, true, int64(len(data))); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] %s 更新 Image Layer 本地存储路径标志错误: %s ", imageId, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"错误\": \"更新 Image Layer 本地存储路径错误\"}"))
		this.StopRun()
	}

	//成功则返回 200
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))

}

func (this *ImageAPIController) PutChecksum() {
	isAuth, errCode, errInfo := models.DoAuthPutChecksum(this.Ctx)

	if !isAuth {
		this.Ctx.Output.Context.Output.SetStatus(errCode)
		this.Ctx.Output.Context.Output.Body(errInfo)
		this.StopRun()
	}

	beego.Debug("[Cookie] " + this.Ctx.Input.Header("Cookie"))
	beego.Debug("[X-Docker-Checksum] " + this.Ctx.Input.Header("X-Docker-Checksum"))
	beego.Debug("[X-Docker-Checksum-Payload] " + this.Ctx.Input.Header("X-Docker-Checksum-Payload"))

	//初始化加密签名
	sign := ""
	if len(string(this.Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(this.Ctx.Input.Header("X-Docker-Sign"))
	}
	beego.Debug("sign:::", sign)

	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	//TODO 检查上传的 Layer 文件的 SHA256 和传上来的 Checksum 的值是否一致？

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

func (this *ImageAPIController) GetImageAncestry() {
	isAuth, errCode, errInfo := models.DoAuthGetImageAncestry(this.Ctx)

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

func (this *ImageAPIController) GetImageLayer() {

	isAuth, errCode, errInfo := models.DoAuthGetImageLayer(this.Ctx)

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
	//读取的文件放在 HTTP Body 中返回给 Docker 命令
	//this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/x-tar")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Transfer-Encoding", "binary")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Length", string(int64(len(file))))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(file)

}
