package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type RepoWebAPIV1Controller struct {
	beego.Controller
}

func (this *RepoWebAPIV1Controller) Prepare() {
	this.EnableXSRF = false

	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist {
		user.GetById(user.Id)
		this.Ctx.Input.CruSession.Set("user", user)
	}

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *RepoWebAPIV1Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		result := map[string]string{"message": message}
		this.Data["json"] = result
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *RepoWebAPIV1Controller) URLMapping() {
	this.Mapping("PostRepository", this.PostRepository)
}

func (this *RepoWebAPIV1Controller) PostRepository() {
	var user models.User
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {
		beego.Error("[WEB API V1] Load session failure")
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	var repo models.Repository

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		beego.Error("[WEB API V1] Unmarshal Repository create data error:", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else {
		beego.Debug("[WEB API V1] Repository create:", string(this.Ctx.Input.CopyBody()))
		if exist, _, err := repo.Has(repo.Namespace, repo.Repository); err != nil {
			beego.Error("[WEB API] Repository create error: ", err.Error())
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		} else if exist == true {
			beego.Error("[WEB API V1] Repository already exist:", fmt.Sprint(repo.Namespace, "/", repo.Repository))
			this.JSONOut(http.StatusBadRequest, "Repository already exist.", nil)
			return
		} else {
			repo.Id = string(utils.GeneralKey(fmt.Sprint(repo.Namespace, repo.Repository)))
			repo.Created = time.Now().UnixNano() / int64(time.Millisecond)
			repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

			if err := repo.Save(); err != nil {
				beego.Error("[WEB API V1] Repository save error:", err.Error())
				this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
				return
			}

			if user.Username == repo.Namespace {
				user.Repositories = append(user.Repositories, repo.Id)
				if err := user.Save(); err != nil {
					beego.Error("[WEB API V1] User save error:", err.Error())
					this.JSONOut(http.StatusBadRequest, err.Error(), nil)
					return
				}
				this.Ctx.Input.CruSession.Set("user", user)
			} else {
				org := new(models.Organization)
				if exist, _, _ := org.Has(repo.Namespace); exist == true {
					org.Repositories = append(org.Repositories, repo.Id)
					if err := org.Save(); err != nil {
						beego.Error("[WEB API V1] Organization save error:", err.Error())
						this.JSONOut(http.StatusBadRequest, "Organization save error.", nil)
						return
					}
				}
			}

			memo, _ := json.Marshal(this.Ctx.Input.Header)
			if err := repo.Log(models.ACTION_ADD_REPO, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, repo.Id, memo); err != nil {
				beego.Error("[WEB API V1] Log Erro:", err.Error())
			}
			if err := user.Log(models.ACTION_ADD_REPO, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo); err != nil {
				beego.Error("[WEB API V1] Log Erro:", err.Error())
			}

			this.JSONOut(http.StatusOK, "Repository create successfully!", nil)
			return
		}
	}
}

func (this *RepoWebAPIV1Controller) GetRepository() {
	repo := new(models.Repository)

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		beego.Error("[WEB API V1] Search repository error:", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		beego.Error("[WEB API V1] Search repository don't exist:", this.Ctx.Input.Param(":namespace"), '/', this.Ctx.Input.Param(":repository"))
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	this.JSONOut(http.StatusOK, "", repo)
	return
}

func (this *RepoWebAPIV1Controller) PutRepository() {
	repo := new(models.Repository)

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		beego.Error("[WEB API V1] Search repository error:", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		beego.Error("[WEB API V1] Search repository don't exist:", this.Ctx.Input.Param(":namespace"), '/', this.Ctx.Input.Param(":repository"))
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		beego.Error("[WEB API V1] Unmarshal Repository create data error:", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := repo.Save(); err != nil {
		beego.Error("[WEB API V1] Repository save error:", err.Error())
		this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
		return
	}

	this.JSONOut(http.StatusOK, "Repository update successfully!", nil)
	return
}

func (this *RepoWebAPIV1Controller) GetRepositories() {
	var user models.User
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {
		beego.Error("[WEB API V1] Load session failure")
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	repositories := make([]models.Repository, 0)
	for _, repoUUID := range user.Repositories {
		repo := new(models.Repository)
		if err := repo.Get(repoUUID); err != nil {
			beego.Error("[WEB API] Repository get error,err=", err.Error())
			continue
		}
		repositories = append(repositories, *repo)
	}

	beego.Debug("[WEB API V1] ", user.Username, " Repositories: ", repositories)
	this.JSONOut(http.StatusOK, "", user)
	return
}

func (this *RepoWebAPIV1Controller) GetCollaborators() {

}

func (this *RepoWebAPIV1Controller) PostCollaborator() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		beego.Error("[WEB API V1] Load session failure")
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	} else {
		repo := new(models.Repository)

		//Collaborator In Organization
		if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil || exist == false {
			beego.Error("[WEB API V1] Search repository error:", err.Error())
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		}

		if user.Username == this.Ctx.Input.Param("namespace") {
			u := new(models.User)

			if exist, _, err := u.Has(this.Ctx.Input.Param("namespace")); err != nil {

			} else if exist == false {

			} else {

			}
		} else {
			for _, v := range repo.Permissions {
				if v == this.Ctx.Input.Param(":collaborator") {
					beego.Error("[WEB API V1] Collaborator already in permissions")
					this.JSONOut(http.StatusBadRequest, "Collaborator already in permissions", nil)
					return
				}
			}

			repo.Permissions = append(repo.Permissions, this.Ctx.Input.Param(":collaborator"))

			if err := repo.Save(); err != nil {
				beego.Error("[WEB API V1] Repository save error:", err.Error())
				this.JSONOut(http.StatusBadRequest, err.Error(), nil)
				return
			}

			this.JSONOut(http.StatusOK, "Repository update successfully!", nil)
			return
		}
	}
}

func (this *RepoWebAPIV1Controller) PutCollaborator() {

}
