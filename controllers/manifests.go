package controllers

import (
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/modules"
)

type ManifestsAPIV2Controller struct {
	beego.Controller
}

func (this *ManifestsAPIV2Controller) URLMapping() {
}

func (this *ManifestsAPIV2Controller) Prepare() {
	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)
	beego.Debug(this.Ctx.Request.URL)

	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *ManifestsAPIV2Controller) PutManifests() {
	if auth, _, _ := modules.AuthManifests(this.Ctx); auth == false {
		result := map[string][]V2ErrorDescriptor{"errors": []V2ErrorDescriptor{V2ErrorDescriptors[APIV2ErrorCodeUnauthorized]}}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.ServeJson()
		return
	}

	manifest, _ := ioutil.ReadAll(this.Ctx.Request.Body)

	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")
	tag := this.Ctx.Input.Param(":tag")

	repo := new(models.Repository)

	if err := repo.Put(namespace, repository, "", this.Ctx.Input.Header("User-Agent"), models.APIVERSION_V2); err != nil {
		result := map[string][]V2ErrorDescriptor{"errors": []V2ErrorDescriptor{V2ErrorDescriptors[APIV2ErrorCodeManifestInvalid]}}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	if err := repo.PutManifests(string(manifest), namespace, repository, tag); err != nil {
		result := map[string][]V2ErrorDescriptor{"errors": []V2ErrorDescriptor{V2ErrorDescriptors[APIV2ErrorCodeManifestInvalid]}}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	beego.Debug("[REGISTRY API V2] Manifests Body: ", string(manifest))

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}
