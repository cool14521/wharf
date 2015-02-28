package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"net/http"
	"time"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type RepoWebAPIV1Controller struct {
	beego.Controller
}

func (this *RepoWebAPIV1Controller) URLMapping() {
	this.Mapping("PostRepository", this.PostRepository)
}

func (this *RepoWebAPIV1Controller) PostRepository() {

	var user models.User
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {
		beego.Error("[WEB API] Load session failure")
		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	var repo models.Repository

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		beego.Error("[WEB API] Unmarshal Repository create data error:", err.Error())
		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	} else {
		beego.Debug("[WEB API] Repository create:", string(this.Ctx.Input.CopyBody()))
		if exist, _, err := repo.Has(repo.Namespace, repo.Repository); err != nil {
			beego.Error("[WEB API] Repository create error: ", err.Error())
			result := map[string]string{"message": err.Error()}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			this.StopRun()
		} else if exist == true {
			beego.Error("[WEB API] Repository already exist:", fmt.Sprint(repo.Namespace, "/", repo.Repository))

			result := map[string]string{"message": "Repository already exist."}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			this.StopRun()
		} else {
			repo.UUID = string(utils.GeneralKey(fmt.Sprint(repo.Namespace, repo.Repository)))
			repo.Created = time.Now().Unix()

			if err := repo.Save(); err != nil {
				beego.Error("[WEB API] Repository save error:", err.Error())
				result := map[string]string{"message": "Repository save error."}
				this.Data["json"] = result

				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.ServeJson()
				this.StopRun()
			}

			if repo.NamespaceType {
				org := new(models.Organization)
				if exist, _, _ := org.Has(repo.Namespace); exist {
					org.Repositories = append(org.Repositories, repo.UUID)
					if err := org.Save(); err != nil {
						beego.Error("[WEB API] Repository save error:", err.Error())
						result := map[string]string{"message": "Repository save error."}
						this.Data["json"] = result

						this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
						this.ServeJson()
						this.StopRun()
					}
				}
			} else {
				user.Repositories = append(user.Repositories, repo.UUID)
				if err := user.Save(); err != nil {
					beego.Error("[WEB API] Repository save error:", err.Error())
					result := map[string]string{"message": "Repository save error."}
					this.Data["json"] = result

					this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
					this.ServeJson()
					this.StopRun()
				}
				this.Ctx.Input.CruSession.Set("user", user)
			}

			memo, _ := json.Marshal(this.Ctx.Input.Header)
			if err := repo.Log(models.ACTION_ADD_REPO, models.LEVELINFORMATIONAL, models.TYPE_WEB, repo.UUID, memo); err != nil {
				beego.Error("[WEB API] Log Erro:", err.Error())
			}
			if err := user.Log(models.ACTION_ADD_REPO, models.LEVELINFORMATIONAL, models.TYPE_WEB, user.UUID, memo); err != nil {
				beego.Error("[WEB API] Log Erro:", err.Error())
			}

			result := map[string]string{"message": "Repository create successfully!"}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.ServeJson()
			this.StopRun()
		}
	}
}
