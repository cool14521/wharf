package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
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

func (this *ImageAPIV1Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		this.Data["json"] = map[string]string{"message": message}
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *ImageAPIV1Controller) Prepare() {
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))
}

func (this *ImageAPIV1Controller) GetImageJSON() {
	imageId := string(this.Ctx.Input.Param(":image_id"))
	image := new(models.Image)

	var json []byte
	var checksum []byte
	var err error

	if json, err = image.GetJSON(imageId); err != nil {
		this.JSONOut(http.StatusBadRequest, "Search Image JSON Error", nil)
		return
	}

	if checksum, err = image.GetChecksum(imageId); err != nil {
		this.JSONOut(http.StatusBadRequest, "Search Image Checksum Error", nil)
		return
	} else {
		this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Checksum", string(checksum))
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(json)
	return
}

func (this *ImageAPIV1Controller) PutImageJSON() {
	imageId := this.Ctx.Input.Param(":image_id")

	image := new(models.Image)

	j := string(this.Ctx.Input.CopyBody())

	if err := image.PutJSON(imageId, j, models.APIVERSION_V1); err != nil {
		this.JSONOut(http.StatusBadRequest, "Put Image JSON Error", nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	image.Log(models.ACTION_PUT_IMAGES_JSON, models.LEVELINFORMATIONAL, models.TYPE_APIV1, image.Id, memo)

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *ImageAPIV1Controller) PutImageLayer() {
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

	if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
		this.JSONOut(http.StatusBadRequest, "Put Image Layer File Error", nil)
		return
	}

	if err := image.PutLayer(imageId, layerfile, true, int64(len(data))); err != nil {
		this.JSONOut(http.StatusBadRequest, "Put Image Layer File Data Error", nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	image.Log(models.ACTION_PUT_IMAGES_LAYER, models.LEVELINFORMATIONAL, models.TYPE_APIV1, image.Id, memo)

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *ImageAPIV1Controller) PutChecksum() {
	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	checksum := strings.Split(this.Ctx.Input.Header("X-Docker-Checksum"), ":")[1]
	payload := strings.Split(this.Ctx.Input.Header("X-Docker-Checksum-Payload"), ":")[1]

	if err := image.PutChecksum(imageId, checksum, true, payload); err != nil {
		this.JSONOut(http.StatusBadRequest, "Put Image Checksum & Payload Error", nil)
		return
	}

	if err := image.PutAncestry(imageId); err != nil {
		this.JSONOut(http.StatusBadRequest, "Put Image Ancestry Error", nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	image.Log(models.ACTION_PUT_IMAGES_CHECKSUM, models.LEVELINFORMATIONAL, models.TYPE_APIV1, image.Id, memo)

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *ImageAPIV1Controller) GetImageAncestry() {
	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	if has, _, err := image.Has(imageId); err != nil {
		this.JSONOut(http.StatusBadRequest, "Read Image Ancestry Error", nil)
		return
	} else if has == false {
		this.JSONOut(http.StatusBadRequest, "Read Image None", nil)
		return
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(image.Ancestry))
	return
}

func (this *ImageAPIV1Controller) GetImageLayer() {
	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	if has, _, err := image.Has(imageId); err != nil {
		this.JSONOut(http.StatusBadRequest, "Read Image Layer file Error", nil)
		return
	} else if has == false {
		this.JSONOut(http.StatusBadRequest, "Read Image None", nil)
		return
	}

	layerfile := image.Path

	if _, err := os.Stat(layerfile); err != nil {
		this.JSONOut(http.StatusBadRequest, "Read Image file state error", nil)
		return
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		this.JSONOut(http.StatusBadRequest, "Read Image file error", nil)
		return
	}

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Transfer-Encoding", "binary")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Length", string(int64(len(file))))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(file)
	return
}
