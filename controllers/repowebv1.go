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

func (this *RepoWebAPIV1Controller) URLMapping() {
	this.Mapping("PostRepository", this.PostRepository)
}

func (this *RepoWebAPIV1Controller) PostRepository() {
	var user models.User
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {
		beego.Error("[WEB API V1] Load session failure")
		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	var repo models.Repository

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		beego.Error("[WEB API V1] Unmarshal Repository create data error:", err.Error())
		result := map[string]string{"message": err.Error()}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	} else {
		beego.Debug("[WEB API V1] Repository create:", string(this.Ctx.Input.CopyBody()))
		if exist, _, err := repo.Has(repo.Namespace, repo.Repository); err != nil {
			beego.Error("[WEB API] Repository create error: ", err.Error())
			result := map[string]string{"message": err.Error()}
			this.Data["json"] = &result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			return
		} else if exist == true {
			beego.Error("[WEB API V1] Repository already exist:", fmt.Sprint(repo.Namespace, "/", repo.Repository))

			result := map[string]string{"message": "Repository already exist."}
			this.Data["json"] = &result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			return
		} else {
			repo.Id = string(utils.GeneralKey(fmt.Sprint(repo.Namespace, repo.Repository)))
			repo.Created = time.Now().UnixNano() / int64(time.Millisecond)
			repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

			if err := repo.Save(); err != nil {
				beego.Error("[WEB API V1] Repository save error:", err.Error())
				result := map[string]string{"message": "Repository save error."}
				this.Data["json"] = &result

				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.ServeJson()
				return
			}

			if user.Username == repo.Namespace {
				user.Repositories = append(user.Repositories, repo.Id)
				if err := user.Save(); err != nil {
					beego.Error("[WEB API V1] User save error:", err.Error())
					result := map[string]string{"message": "User save error."}
					this.Data["json"] = &result

					this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
					this.ServeJson()
					return
				}
				this.Ctx.Input.CruSession.Set("user", user)
			} else {
				org := new(models.Organization)
				if exist, _, _ := org.Has(repo.Namespace); exist == true {
					org.Repositories = append(org.Repositories, repo.Id)
					if err := org.Save(); err != nil {
						beego.Error("[WEB API V1] Organization save error:", err.Error())
						result := map[string]string{"message": "Organization save error."}
						this.Data["json"] = &result

						this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
						this.ServeJson()
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

			result := map[string]string{"message": "Repository create successfully!"}
			this.Data["json"] = &result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.ServeJson()
			return
		}
	}
}

func (this *RepoWebAPIV1Controller) GetRepository() {
	repo := new(models.Repository)

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		beego.Error("[WEB API V1] Search repository error:", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	} else if exist == false {
		beego.Error("[WEB API V1] Search repository don't exist:", this.Ctx.Input.Param(":namespace"), '/', this.Ctx.Input.Param(":repository"))

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	this.Data["json"] = &repo

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *RepoWebAPIV1Controller) PutRepository() {
	repo := new(models.Repository)

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		beego.Error("[WEB API V1] Search repository error:", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	} else if exist == false {
		beego.Error("[WEB API V1] Search repository don't exist:", this.Ctx.Input.Param(":namespace"), '/', this.Ctx.Input.Param(":repository"))

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		beego.Error("[WEB API V1] Unmarshal Repository create data error:", err.Error())
		result := map[string]string{"message": err.Error()}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := repo.Save(); err != nil {
		beego.Error("[WEB API V1] Repository save error:", err.Error())
		result := map[string]string{"message": "Repository save error."}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	result := map[string]string{"message": "Repository update successfully!"}
	this.Data["json"] = &result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *RepoWebAPIV1Controller) GetRepositories() {
	var user models.User
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {
		beego.Error("[WEB API V1] Load session failure")
		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	repositories := make([]models.Repository, 0)
	for _, repoUUID := range user.Repositories {
		repo := new(models.Repository)
		if err := repo.Get(repoUUID); err != nil {
			beego.Error("[WEB API] Repository get error,err=", err.Error())
			continue
		} //else if repo.NamespaceType {
		//continue
		//}
		repositories = append(repositories, *repo)
	}

	beego.Debug("[WEB API V1] ", user.Username, " Repositories: ", repositories)

	//Repositories object to User
	this.Data["json"] = &user

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}
