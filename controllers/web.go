package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/shurcooL/go/github_flavored_markdown"
)

type WebController struct {
	beego.Controller
}

func (this *WebController) URLMapping() {
	this.Mapping("GetIndex", this.GetIndex)
	this.Mapping("GetAuth", this.GetAuth)
	this.Mapping("GetDashboard", this.GetDashboard)
	this.Mapping("GetSetting", this.GetSetting)
	this.Mapping("GetRepository", this.GetRepository)
	this.Mapping("GetAdmin", this.GetAdmin)
	this.Mapping("GetAdminAuth", this.GetAdminAuth)
	this.Mapping("GetSignout", this.GetSignout)
}

func (this *WebController) Prepare() {
	this.EnableXSRF = false

	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *WebController) GetIndex() {
	this.TplNames = "index.html"
	this.Render()
	return
}

func (this *WebController) GetAuth() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		this.TplNames = "auth.html"
		this.Render()

		return
	} else {
		this.Ctx.Redirect(http.StatusMovedPermanently, "/dashboard")
	}
}

func (this *WebController) GetDashboard() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		beego.Error("[WEB API] Load session failure")
		this.Ctx.Redirect(http.StatusMovedPermanently, "/auth")
		return
	} else {
		this.TplNames = "dashboard.html"
		this.Data["username"] = user.Username

		this.Render()
		return
	}
}

func (this *WebController) GetSetting() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		beego.Error("[WEB API] Load session failure")
		this.Ctx.Redirect(http.StatusMovedPermanently, "/auth")

		return
	} else {
		this.TplNames = "setting.html"
		this.Data["username"] = user.Username

		this.Render()
		return
	}
}

func (this *WebController) GetRepository() {
	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repository")

	repo := new(models.Repository)
	if exist, _, _ := repo.Has(namespace, repository); exist {
		user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
		if repo.Privated {
			if !exist == true {
				this.Abort("404")
				return
				// } else if repo.NamespaceType {
				// 	userHasOrg := false

				// 	org := new(models.Organization)
				// 	_, orgUUID, _ := org.Has(namespace)
				// 	for _, userOrgUUID := range user.Organizations {
				// 		if userOrgUUID == string(orgUUID) {
				// 			userHasOrg = true
				// 			return
				// 		}
				// 	}
				// 	if !userHasOrg {
				// 		this.Abort("404")
				// 		return
				// 	}
				// 	this.Data["username"] = user.Username
				// 	this.Data["privated"] = repo.Privated
				// 	this.Data["namespace"] = repo.Namespace
				// 	this.Data["repository"] = repo.Repository
				// 	this.Data["created"] = repo.Created
				// 	this.Data["short"] = repo.Short
				// 	this.Data["description"] = string(github_flavored_markdown.Markdown([]byte(repo.Description)))
				// 	this.Data["download"] = repo.Download
				// 	this.Data["comments"] = len(repo.Comments)
				// 	this.Data["starts"] = len(repo.Starts)

				// 	this.TplNames = "repository.html"
				// 	this.Render()
				// 	return

			} else {
				if user.Username != namespace {
					this.Abort("404")
					return
				}
				this.Data["username"] = user.Username
				this.Data["privated"] = repo.Privated
				this.Data["namespace"] = repo.Namespace
				this.Data["repository"] = repo.Repository
				this.Data["created"] = repo.Created
				this.Data["short"] = repo.Short
				this.Data["description"] = string(github_flavored_markdown.Markdown([]byte(repo.Description)))
				this.Data["download"] = repo.Download
				this.Data["comments"] = len(repo.Comments)
				this.Data["starts"] = len(repo.Starts)

				this.TplNames = "repository.html"
				this.Render()
				return
			}
		} else {
			this.Data["username"] = user.Username
			this.Data["privated"] = repo.Privated
			this.Data["namespace"] = repo.Namespace
			this.Data["repository"] = repo.Repository
			this.Data["created"] = repo.Created
			this.Data["short"] = repo.Short
			this.Data["description"] = string(github_flavored_markdown.Markdown([]byte(repo.Description)))
			this.Data["download"] = repo.Download
			this.Data["comments"] = len(repo.Comments)
			this.Data["starts"] = len(repo.Starts)

			this.TplNames = "repository.html"
			this.Render()
			return
		}
	} else {
		this.Abort("404")
		return
	}
	return
}

func (this *WebController) GetAdmin() {
	this.TplNames = "admin.html"

	this.Data["username"] = "genedna"

	this.Render()
	return
}

func (this *WebController) GetAdminAuth() {
	this.TplNames = "admin-auth.html"

	this.Render()
	return
}

func (this *WebController) GetSignout() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		this.TplNames = "auth.html"
		this.Render()

		return
	} else {
		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := user.Log(models.ACTION_SINGOUT, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo); err != nil {
			beego.Error("[WEB] Log Erro:", err.Error())
		}

		this.Ctx.Input.CruSession.Delete("user")
		this.Ctx.Redirect(http.StatusMovedPermanently, "/auth")
	}
}
