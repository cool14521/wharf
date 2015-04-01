package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/astaxie/beego"

	"github.com/containerops/wharf/models"
	"github.com/containerops/wharf/utils"
)

type RepoAPIV1Controller struct {
	beego.Controller
}

func (this *RepoAPIV1Controller) URLMapping() {
	this.Mapping("PutTag", this.PutTag)
	this.Mapping("PutRepositoryImages", this.PutRepositoryImages)
	this.Mapping("GetRepositoryImages", this.GetRepositoryImages)
	this.Mapping("GetRepositoryTags", this.GetRepositoryTags)
	this.Mapping("PutRepository", this.PutRepository)
}

func (this *RepoAPIV1Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		this.Data["json"] = map[string]string{"message": message}
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *RepoAPIV1Controller) Prepare() {
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))
}

func (this *RepoAPIV1Controller) PutRepository() {
	username, _, _ := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization"))

	namespace := string(this.Ctx.Input.Param(":namespace"))
	repository := string(this.Ctx.Input.Param(":repo_name"))

	repo := new(models.Repository)

	if err := repo.Put(namespace, repository, string(this.Ctx.Input.CopyBody()), this.Ctx.Input.Header("User-Agent"), models.APIVERSION_V1); err != nil {
		this.JSONOut(http.StatusForbidden, err.Error(), nil)
		return
	}

	if this.Ctx.Input.Header("X-Docker-Token") == "true" {
		token := string(utils.GeneralKey(username))
		this.SetSession("token", token)
		this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Token", token)
		this.Ctx.Output.Context.ResponseWriter.Header().Set("WWW-Authenticate", token)
	}

	user := new(models.User)
	if _, _, err := user.Has(username); err != nil {
		this.JSONOut(http.StatusForbidden, err.Error(), nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	user.Log(models.ACTION_UPDATE_REPO, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.Id, memo)
	repo.Log(models.ACTION_UPDATE_REPO, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.Id, memo)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Endpoints", beego.AppConfig.String("docker::Endpoints"))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *RepoAPIV1Controller) PutTag() {
	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")

	tag := this.Ctx.Input.Param(":tag")

	r, _ := regexp.Compile(`"([[:alnum:]]+)"`)
	imageIds := r.FindStringSubmatch(string(this.Ctx.Input.CopyBody()))

	repo := new(models.Repository)
	if err := repo.PutTag(imageIds[1], namespace, repository, tag); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	repo.Log(models.ACTION_PUT_TAG, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.Id, memo)

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *RepoAPIV1Controller) PutRepositoryImages() {
	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")

	repo := new(models.Repository)

	if err := repo.PutImages(namespace, repository); err != nil {
		this.JSONOut(http.StatusBadRequest, "Update Uploaded flag error", nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	repo.Log(models.ACTION_PUT_REPO_IMAGES, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.Id, memo)

	org := new(models.Organization)
	isOrg, _, err := org.Has(namespace)
	if err != nil {
		this.JSONOut(http.StatusBadRequest, "Search Organization Error", nil)
		return
	}

	user := new(models.User)
	authUsername, _, _ := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization"))
	isUser, _, err := user.Has(authUsername)
	if err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	if !isUser && !isOrg {
		this.JSONOut(http.StatusBadRequest, "Search Namespace Error", nil)
		return
	}

	if isUser {
		user.Repositories = append(user.Repositories, repo.Id)
		user.Save()
	}
	if isOrg {
		org.Repositories = append(org.Repositories, repo.Id)
		org.Save()
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusNoContent)
	this.Ctx.Output.Context.Output.Body([]byte(""))

	this.ServeJson()
	return
}

func (this *RepoAPIV1Controller) GetRepositoryImages() {
	namespace := string(this.Ctx.Input.Param(":namespace"))
	repository := string(this.Ctx.Input.Param(":repo_name"))

	repo := new(models.Repository)

	if has, _, err := repo.Has(namespace, repository); err != nil {
		this.JSONOut(http.StatusBadRequest, "Read repository json error", nil)
		return
	} else if has == false {
		this.JSONOut(http.StatusBadRequest, "Read repository no found", nil)
		return
	}

	repo.Download += 1

	if err := repo.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	repo.Log(models.ACTION_GET_REPO, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.Id, memo)

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(repo.JSON))
	return
}

func (this *RepoAPIV1Controller) GetRepositoryTags() {
	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")

	repo := new(models.Repository)

	if has, _, err := repo.Has(namespace, repository); err != nil {
		this.JSONOut(http.StatusBadRequest, "Read repository json error", nil)
		return
	} else if has == false {
		this.JSONOut(http.StatusBadRequest, "Read repository no found", nil)
		return
	}

	tag := map[string]string{}

	for _, value := range repo.Tags {
		t := new(models.Tag)
		if err := t.GetById(value); err != nil {
			this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": fmt.Sprintf("%s/%s Tags is not exist", namespace, repository)})
			return
		}

		tag[t.Name] = t.ImageId
	}

	this.JSONOut(http.StatusOK, "", tag)
	return
}
