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
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	var repo models.Repository

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else {
		if exist, _, err := repo.Has(repo.Namespace, repo.Repository); err != nil {
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		} else if exist == true {
			this.JSONOut(http.StatusBadRequest, "Repository already exist.", nil)
			return
		} else {
			repo.Id = string(utils.GeneralKey(fmt.Sprint(repo.Namespace, repo.Repository)))
			repo.Created = time.Now().UnixNano() / int64(time.Millisecond)
			repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

			if err := repo.Save(); err != nil {
				this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
				return
			}

			if user.Username == repo.Namespace {
				user.Repositories = append(user.Repositories, repo.Id)
				if err := user.Save(); err != nil {
					this.JSONOut(http.StatusBadRequest, err.Error(), nil)
					return
				}
				this.Ctx.Input.CruSession.Set("user", user)
			} else {
				org := new(models.Organization)
				if exist, _, _ := org.Has(repo.Namespace); exist == true {
					org.Repositories = append(org.Repositories, repo.Id)
					if err := org.Save(); err != nil {
						this.JSONOut(http.StatusBadRequest, "Organization save error.", nil)
						return
					}
				}
			}

			memo, _ := json.Marshal(this.Ctx.Input.Header)
			repo.Log(models.ACTION_ADD_REPO, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, repo.Id, memo)
			user.Log(models.ACTION_ADD_REPO, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo)

			this.JSONOut(http.StatusOK, "Repository create successfully!", nil)
			return
		}
	}
}

func (this *RepoWebAPIV1Controller) GetRepository() {
	repo := new(models.Repository)

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	this.JSONOut(http.StatusOK, "", repo)
	return
}

func (this *RepoWebAPIV1Controller) PutRepository() {
	repo := new(models.Repository)

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := repo.Save(); err != nil {
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
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	repositories := make([]models.Repository, 0)
	for _, repoUUID := range user.Repositories {
		repo := new(models.Repository)
		if err := repo.Get(repoUUID); err != nil {
			continue
		}
		repositories = append(repositories, *repo)
	}

	this.JSONOut(http.StatusOK, "", user)
	return
}

func (this *RepoWebAPIV1Controller) GetCollaborators() {

}

func (this *RepoWebAPIV1Controller) PostCollaborator() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	} else {
		repo := new(models.Repository)

		//Collaborator In Organization
		if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil || exist == false {
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		}

		if user.Username == this.Ctx.Input.Param("namespace") {
		} else {
			for _, v := range repo.Permissions {
				if v == this.Ctx.Input.Param(":collaborator") {
					this.JSONOut(http.StatusBadRequest, "Collaborator already in permissions", nil)
					return
				}
			}

			repo.Permissions = append(repo.Permissions, this.Ctx.Input.Param(":collaborator"))

			if err := repo.Save(); err != nil {
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
