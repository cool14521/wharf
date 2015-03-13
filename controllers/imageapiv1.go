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
		beego.Error("[REGISTRY API V1] Search Image JSON Error: ", err.Error())
		result := map[string]string{"Error": "Search Image JSON Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	if checksum, err = image.GetChecksum(imageId); err != nil {
		beego.Error("[REGISTRY API V1] Search Image Checksum Error: ", err.Error())
		result := map[string]string{"Error": "Search Image Checksum Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
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
		beego.Error("[REGISTRY API V1] Put Image JSON Error: ", err.Error())
		result := map[string]string{"Error": "Put Image JSON Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := image.Log(models.ACTION_PUT_IMAGES_JSON, models.LEVELINFORMATIONAL, models.TYPE_APIV1, image.UUID, memo); err != nil {
		beego.Error("[REGISTRY API V1] Log Error:", err.Error())
	}

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
		beego.Error("[REGISTRY API V1] Put Image Layer File Error: ", err.Error())
		result := map[string]string{"Error": "Put Image Layer File Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	if err := image.PutLayer(imageId, layerfile, true, int64(len(data))); err != nil {
		beego.Error("[REGISTRY API V1] Put Image Layer File Data Error: ", err.Error())
		result := map[string]string{"Error": "Put Image Layer File Data Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := image.Log(models.ACTION_PUT_IMAGES_LAYER, models.LEVELINFORMATIONAL, models.TYPE_APIV1, image.UUID, memo); err != nil {
		beego.Error("[REGISTRY API V1] Log Error:", err.Error())
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *ImageAPIV1Controller) PutChecksum() {
	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	checksum := strings.Split(this.Ctx.Input.Header("X-Docker-Checksum"), ":")[1]
	payload := strings.Split(this.Ctx.Input.Header("X-Docker-Checksum-Payload"), ":")[1]

	beego.Debug("[REGISTRY API V1] Image Checksum : ", checksum)
	beego.Debug("[REGISTRY API V1] Image Payload: ", payload)

	if err := image.PutChecksum(imageId, checksum, true, payload); err != nil {
		beego.Error("[REGISTRY API V1] Put Image Checksum & Payload Error: ", err.Error())
		result := map[string]string{"Error": "Put Image Checksum & Payload Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	if err := image.PutAncestry(imageId); err != nil {
		beego.Error("[REGISTRY API V1] Put Image Ancestry Error: ", err.Error())
		result := map[string]string{"Error": "Put Image Ancestry Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := image.Log(models.ACTION_PUT_IMAGES_CHECKSUM, models.LEVELINFORMATIONAL, models.TYPE_APIV1, image.UUID, memo); err != nil {
		beego.Error("[REGISTRY API V1] Log Error:", err.Error())
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *ImageAPIV1Controller) GetImageAncestry() {
	imageId := string(this.Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	if has, _, err := image.Has(imageId); err != nil {
		beego.Error("[REGISTRY API V1] Read Image Ancestry Error: ", err.Error())
		result := map[string]string{"Error": "Read Image Ancestry Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	} else if has == false {
		beego.Error("[REGISTRY API V1] Read Image None: ", err.Error())
		result := map[string]string{"Error": "Read Image None"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
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
		beego.Error("[REGISTRY API V1] Read Image Layer File Status Error: ", err.Error())
		result := map[string]string{"Error": "Read Image Layer file Error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	} else if has == false {
		beego.Error("[REGISTRY API V1] Read Image None Error")
		result := map[string]string{"Error": "Read Image None"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	layerfile := image.Path

	if _, err := os.Stat(layerfile); err != nil {
		beego.Error("[REGISTRY API V1] Read Image file state error: ", err.Error())
		result := map[string]string{"Error": "Read Image file state error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		beego.Error("[REGISTRY API V1] Read Image file error: ", err.Error())
		result := map[string]string{"Error": "Read Image file error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Transfer-Encoding", "binary")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Length", string(int64(len(file))))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(file)
	return
}
