package controllers

import (
	"encoding/json"
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

func (this *ImageAPIV1Controller) URLMapping() {
	this.Mapping("GetImageJSON", this.GetImageJSON)
	this.Mapping("PutImageJSON", this.PutImageJSON)
	this.Mapping("PutImageLayer", this.PutImageLayer)
	this.Mapping("PutChecksum", this.PutChecksum)
	this.Mapping("GetImageAncestry", this.GetImageAncestry)
	this.Mapping("GetImageLayer", this.GetImageLayer)
}

func (this *ImageAPIV1Controller) Prepare() {
	beego.Debug("[Header]")
	beego.Debug(this.Ctx.Request.Header)
	beego.Debug(this.Ctx.Request.URL)

	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))

}

func (this *ImageAPIV1Controller) GetImageJSON() {
	if auth, code, message := modules.AuthGetImageJSON(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	imageId := string(this.Ctx.Input.Param(":image_id"))
	image := new(models.Image)

	var json []byte
	var checksum []byte
	var err error

	if json, err = image.GetJSON(imageId); err != nil {
		beego.Error("[REGISTRY API V1] Search Image JSON Error: ", err.Error())
		result := map[string]string{"Error": "Search Image JSON Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	if checksum, err = image.GetChecksum(imageId); err != nil {
		beego.Error("[REGISTRY API V1] Search Image Checksum Error: ", err.Error())
		result := map[string]string{"Error": "Search Image Checksum Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	} else {
		this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Checksum", string(checksum))
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(json)
	this.StopRun()
}

func (this *ImageAPIV1Controller) PutImageJSON() {

	beego.Error("进入PutImageJSON")
	if auth, code, message := modules.AuthPutImageJSON(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result
		beego.Error("进入PutImageJSON1")
		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	imageId := this.Ctx.Input.Param(":image_id")

	image := new(models.Image)

	j := string(this.Ctx.Input.CopyBody())

	if err := image.PutJSON(imageId, j); err != nil {
		beego.Error("[REGISTRY API V1] Put Image JSON Error: ", err.Error())
		result := map[string]string{"Error": "Put Image JSON Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := image.Log(models.ACTION_PUT_IMAGES_JSON, models.LEVELINFORMATIONAL, models.TYPE_API, image.UUID, memo); err != nil {
		beego.Error("[REGISTRY API V1] Log Error:", err.Error())
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	this.StopRun()
}

func (this *ImageAPIV1Controller) PutImageLayer() {
	beego.Error("进入PutImageLayer")
	if auth, code, message := modules.AuthPutImageLayer(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result
		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}
	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	basePath := beego.AppConfig.String("docker::BasePath")
	imagePath := fmt.Sprintf("%v/images/%v", basePath, imageId)
	layerfile := fmt.Sprintf("%v/images/%v/layer", basePath, imageId)

	if !utils.IsDirExists(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}

	if _, err := os.Stat(layerfile); err == nil {
		os.Remove(layerfile)
	}

	data, _ := ioutil.ReadAll(this.Ctx.Request.Body)
	beego.Error("进入PutImageLayer3")
	if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
		beego.Error("[REGISTRY API V1] Put Image Layer File Error: ", err.Error())
		result := map[string]string{"Error": "Put Image Layer File Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	if err := image.PutLayer(imageId, layerfile, true, int64(len(data))); err != nil {
		beego.Error("[REGISTRY API V1] Put Image Layer File Data Error: ", err.Error())
		result := map[string]string{"Error": "Put Image Layer File Data Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := image.Log(models.ACTION_PUT_IMAGES_LAYER, models.LEVELINFORMATIONAL, models.TYPE_API, image.UUID, memo); err != nil {
		beego.Error("[REGISTRY API V1] Log Error:", err.Error())
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	this.StopRun()
}

func (this *ImageAPIV1Controller) PutChecksum() {
	if auth, code, message := modules.AuthPutChecksum(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	if err := image.PutChecksum(imageId, this.Ctx.Input.Header("X-Docker-Checksum"), true, this.Ctx.Input.Header("X-Docker-Checksum-Payload")); err != nil {
		beego.Error("[REGISTRY API V1] Put Image Checksum & Payload Error: ", err.Error())
		result := map[string]string{"Error": "Put Image Checksum & Payload Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	if err := image.PutAncestry(imageId); err != nil {
		beego.Error("[REGISTRY API V1] Put Image Ancestry Error: ", err.Error())
		result := map[string]string{"Error": "Put Image Ancestry Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := image.Log(models.ACTION_PUT_IMAGES_CHECKSUM, models.LEVELINFORMATIONAL, models.TYPE_API, image.UUID, memo); err != nil {
		beego.Error("[REGISTRY API V1] Log Error:", err.Error())
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	this.StopRun()
}

func (this *ImageAPIV1Controller) GetImageAncestry() {
	if auth, code, message := modules.AuthGetImageAncestry(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	if has, _, err := image.Has(imageId); err != nil {
		beego.Error("[REGISTRY API V1] Read Image Ancestry Error: ", err.Error())
		result := map[string]string{"Error": "Read Image Ancestry Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	} else if has == false {
		beego.Error("[REGISTRY API V1] Read Image None: ", err.Error())
		result := map[string]string{"Error": "Read Image None"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(image.Ancestry))
	this.StopRun()
}

func (this *ImageAPIV1Controller) GetImageLayer() {
	if auth, code, message := modules.AuthGetImageLayer(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	if has, _, err := image.Has(imageId); err != nil {
		beego.Error("[REGISTRY API V1] Read Image Layer File Status Error: ", err.Error())
		result := map[string]string{"Error": "Read Image Layer file Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	} else if has == false {
		beego.Error("[REGISTRY API V1] Read Image None Error")
		result := map[string]string{"Error": "Read Image None"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	layerfile := image.Path

	if _, err := os.Stat(layerfile); err != nil {
		beego.Error("[REGISTRY API V1] Read Image file state error: ", err.Error())
		result := map[string]string{"Error": "Read Image file state error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		beego.Error("[REGISTRY API V1] Read Image file error: ", err.Error())
		result := map[string]string{"Error": "Read Image file error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Transfer-Encoding", "binary")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Length", string(int64(len(file))))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(file)
	this.StopRun()
}
