package controllers

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/modules"
	"github.com/dockercn/wharf/utils"
)

type RepoAPIV1Controller struct {
	beego.Controller
}

func (r *RepoAPIV1Controller) URLMapping() {
	r.Mapping("PutTag", r.PutTag)
	r.Mapping("PutRepositoryImages", r.PutRepositoryImages)
	r.Mapping("GetRepositoryImages", r.GetRepositoryImages)
	r.Mapping("GetRepositoryTags", r.GetRepositoryTags)
	r.Mapping("PutRepository", r.PutRepository)
}

func (this *RepoAPIV1Controller) Prepare() {
	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)

	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))

}

func (this *RepoAPIV1Controller) PutRepository() {
	if auth, code, message := modules.AuthPutRepository(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	username, passwd, _ := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization"))

	namespace := string(this.Ctx.Input.Param(":namespace"))
	repository := string(this.Ctx.Input.Param(":repo_name"))

	repo := new(models.Repository)

	if err := repo.Put(namespace, repository, string(this.Ctx.Input.CopyBody()), this.Ctx.Input.Header("User-Agent")); err != nil {
		beego.Error("[REGISTRY API V1] Put repository error: %s", err.Error())

		result := map[string]string{"Error": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
		this.ServeJson()
		this.StopRun()
	}

	if this.Ctx.Input.Header("X-Docker-Token") == "true" {
		token := utils.GeneralToken(username + passwd)
		this.SetSession("token", token)
		this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Token", token)
		this.Ctx.Output.Context.ResponseWriter.Header().Set("WWW-Authenticate", token)
	}

	this.SetSession("username", username)
	this.SetSession("namespace", namespace)
	this.SetSession("repository", repository)
	this.SetSession("access", "write")

	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Endpoints", beego.AppConfig.String("docker::Endpoints"))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))

	this.StopRun()

}

func (this *RepoAPIV1Controller) PutTag() {
	if auth, code, message := modules.AuthPutRepositoryTag(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")

	tag := this.Ctx.Input.Param(":tag")

	r, _ := regexp.Compile(`"([[:alnum:]]+)"`)
	imageIds := r.FindStringSubmatch(string(this.Ctx.Input.CopyBody()))

	repo := new(models.Repository)
	if err := repo.PutTag(imageIds[1], namespace, repository, tag); err != nil {
		beego.Error("[REGISTRY API V1] Put repository tag error: %s", err.Error())

		result := map[string]string{"Error": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusForbidden)
		this.ServeJson()
		this.StopRun()
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	this.StopRun()
}

func (this *RepoAPIV1Controller) PutRepositoryImages() {

	if auth, code, message := modules.AuthPutRepositoryImage(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")

	repo := new(models.Repository)

	if err := repo.PutImages(namespace, repository); err != nil {
		beego.Error("[REGISTRY API V1] Update Uploaded flag error: %s", namespace, repository, err.Error())

		result := map[string]string{"message": "Update Uploaded flag error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusNoContent)
	this.Ctx.Output.Context.Output.Body([]byte(""))

	this.ServeJson()
	this.StopRun()
}

func (this *RepoAPIV1Controller) GetRepositoryImages() {

	if auth, code, message := modules.AuthPutRepositoryImage(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	username, passwd, _ := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization"))

	namespace := string(this.Ctx.Input.Param(":namespace"))
	repository := string(this.Ctx.Input.Param(":repo_name"))

	repo := new(models.Repository)

	if has, _, err := repo.Has(namespace, repository); err != nil {
		beego.Error("[REGISTRY API V1] Read repository json error: %s", namespace, repository, err.Error())

		result := map[string]string{"message": "Read repository json error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	} else if has == false {
		beego.Error("[REGISTRY API V1] Read repository no found", namespace, repository)

		result := map[string]string{"message": "Read repository no found"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	if this.Ctx.Input.Header("X-Docker-Token") == "true" {
		token := utils.GeneralToken(username + passwd)
		this.Ctx.Input.CruSession.Set("token", token)

		this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Token", token)
		this.Ctx.Output.Context.ResponseWriter.Header().Set("WWW-Authenticate", token)
	}

	this.Ctx.Input.CruSession.Set("namespace", namespace)
	this.Ctx.Input.CruSession.Set("repository", repository)
	this.Ctx.Input.CruSession.Set("access", "read")
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(repo.JSON))
	this.StopRun()
}

func (this *RepoAPIV1Controller) GetRepositoryTags() {

	if auth, code, message := modules.AuthGetRepositoryTags(this.Ctx); auth == false {
		result := map[string]string{"message": string(message)}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(code)
		this.ServeJson()
		this.StopRun()
	}

	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")

	repo := new(models.Repository)

	if has, _, err := repo.Has(namespace, repository); err != nil {
		beego.Error("[REGISTRY API V1] Read repository json error: %s", namespace, repository, err.Error())

		result := map[string]string{"message": "Read repository json error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	} else if has == false {
		beego.Error("[REGISTRY API V1] Read repository no found", namespace, repository)

		result := map[string]string{"message": "Read repository no found"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	tag := map[string]string{}

	for _, value := range repo.Tags {

		t := new(models.Tag)
		if err := t.GetByUUID(value); err != nil {
			beego.Error(fmt.Sprintf("[REGISTRY API V1]  %s/%s Tags is not exist", namespace, repository))

			result := map[string]string{"message": fmt.Sprintf("%s/%s Tags is not exist", namespace, repository)}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			this.StopRun()
		}

		tag[t.Name] = t.ImageId
	}

	this.Data["json"] = tag

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	this.StopRun()
}
