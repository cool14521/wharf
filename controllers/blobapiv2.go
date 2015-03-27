package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/modules"
	"github.com/dockercn/wharf/utils"
)

type BlobAPIV2Controller struct {
	beego.Controller
}

func (this *BlobAPIV2Controller) URLMapping() {
	this.Mapping("HeadDigest", this.HeadDigest)
	this.Mapping("PostBlobs", this.PostBlobs)
	this.Mapping("PutBlobs", this.PutBlobs)
	this.Mapping("GetBlobs", this.GetBlobs)
}

func (this *BlobAPIV2Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		result := map[string]string{"message": message}
		this.Data["json"] = result
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *BlobAPIV2Controller) Prepare() {
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

//Has image return 200; other return 404
func (this *BlobAPIV2Controller) HeadDigest() {
	image := new(models.Image)
	digest := strings.Split(this.Ctx.Input.Param(":digest"), ":")[1]

	beego.Debug("[REGISTRY API V2] Tarsum.v1+sha256: ", digest)

	if has, _, _ := image.HasTarsum(digest); has == false {
		this.JSONOut(http.StatusNotFound, "", map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeUnauthorized]}})
		return
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *BlobAPIV2Controller) PostBlobs() {
	uuid := utils.GeneralKey(fmt.Sprintf("%s/%s", this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repo_name")))
	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/uploads/%s", beego.AppConfig.String("docker::Endpoints"), this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repo_name"), uuid)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Location", random)
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Range", "bytes=0-0")
	this.Ctx.Output.Context.Output.SetStatus(http.StatusAccepted)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *BlobAPIV2Controller) PutBlobs() {
	var digest string

	this.Ctx.Input.Bind(&digest, "digest")
	beego.Debug("[REGISTRY API V2] Digest: ", digest)

	basePath := beego.AppConfig.String("docker::BasePath")
	imagePath := fmt.Sprintf("%v/uuid/%v", basePath, strings.Split(digest, ":")[1])
	layerfile := fmt.Sprintf("%v/uuid/%v/layer", basePath, strings.Split(digest, ":")[1])

	if !utils.IsDirExists(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}

	if _, err := os.Stat(layerfile); err == nil {
		os.Remove(layerfile)
	}

	data, _ := ioutil.ReadAll(this.Ctx.Request.Body)

	if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
		this.JSONOut(http.StatusBadRequest, "", map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeBlobUploadUnknown]}})
		return
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusCreated)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *BlobAPIV2Controller) GetBlobs() {
	image := new(models.Image)
	digest := strings.Split(this.Ctx.Input.Param(":digest"), ":")[1]

	beego.Debug("[REGISTRY API V2] Tarsum.v1+sha256: ", digest)

	if has, _, _ := image.HasTarsum(digest); has == false {
		this.JSONOut(http.StatusBadRequest, "", map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeUnauthorized]}})
		return
	}

	layerfile := image.Path

	if _, err := os.Stat(layerfile); err != nil {
		this.JSONOut(http.StatusBadRequest, "", map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeBlobUnknown]}})
		return
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		this.JSONOut(http.StatusBadRequest, "", map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeBlobUnknown]}})
		return
	}

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Transfer-Encoding", "binary")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Length", string(int64(len(file))))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(file)
	return
}
