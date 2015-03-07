package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"net/http"
	"time"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type OrganizationWebV1Controller struct {
	beego.Controller
}

func (this *OrganizationWebV1Controller) URLMapping() {
	this.Mapping("PostOrganization", this.PostOrganization)
	this.Mapping("PutOrganization", this.PutOrganization)
	this.Mapping("GetOrganizations", this.GetOrganizations)
	this.Mapping("GetOrganizationDetail", this.GetOrganizationDetail)
}

func (this *OrganizationWebV1Controller) Prepare() {
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *OrganizationWebV1Controller) PostOrganization() {
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)

	if exist != true {
		beego.Error("[WEB API V1] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	var org models.Organization

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &org); err != nil {
		beego.Error("[WEB API V1] Unmarshal organization data error:", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	beego.Debug("[WEB API V1] organization create: %s", string(this.Ctx.Input.CopyBody()))

	org.UUID = string(utils.GeneralKey(org.Organization))

	org.Username = user.Username
	org.Created = time.Now().UnixNano() / int64(time.Millisecond)
	org.Updated = org.Created
	if err := org.Save(); err != nil {
		beego.Error("[WEB API V1] Organization save error:", err.Error())

		result := map[string]string{"message": "Organization save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	user.Organizations = append(user.Organizations, org.UUID)
	user.Updated = org.Created
	if err := user.Save(); err != nil {
		beego.Error("[WEB API V1] User save error:", err.Error())

		result := map[string]string{"message": "User save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := user.Log(models.ACTION_ADD_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEB, org.UUID, memo); err != nil {
		beego.Error("[WEB API V1] Log Erro:", err.Error())
	}
	if err := org.Log(models.ACTION_ADD_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEB, user.UUID, memo); err != nil {
		beego.Error("[WEB API V1] Log Erro:", err.Error())
	}

	user.Get(user.Username, user.Password)
	this.Ctx.Input.CruSession.Set("user", user)

	result := map[string]string{"message": "Create organization successfully."}
	this.Data["json"] = result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *OrganizationWebV1Controller) PutOrganization() {

	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API V1] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return

	} else {

		var org models.Organization

		if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &org); err != nil {
			beego.Error("[WEB API V1] Unmarshal organization data error:", err.Error())

			result := map[string]string{"message": err.Error()}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			return
		}

		beego.Debug("[WEB API V1] organization update: %s", string(this.Ctx.Input.CopyBody()))

		org.Updated = time.Now().UnixNano() / int64(time.Millisecond)

		if err := org.Save(); err != nil {
			beego.Error("[WEB API V1] Organization save error:", err.Error())

			result := map[string]string{"message": "Organization save error."}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			return
		}

		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := user.Log(models.ACTION_UPDATE_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEB, org.UUID, memo); err != nil {
			beego.Error("[WEB API V1] Log Erro:", err.Error())
		}
		if err := org.Log(models.ACTION_UPDATE_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEB, user.UUID, memo); err != nil {
			beego.Error("[WEB API V1] Log Erro:", err.Error())
		}

		result := map[string]string{"message": "Update organization successfully."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
		return
	}
}

func (this *OrganizationWebV1Controller) GetOrganizations() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API V1] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return

	} else {

		organizations := make([]models.Organization, len(user.Organizations))

		for i, UUID := range user.Organizations {
			if err := organizations[i].Get(UUID); err != nil {
				beego.Error("[WEB API V1] Get organizations error:", err.Error())

				result := map[string]string{"message": "Get organizations error."}
				this.Data["json"] = result

				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.ServeJson()
			}
		}

		this.Data["json"] = organizations

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
		return
	}
}

func (this *OrganizationWebV1Controller) GetOrganizationDetail() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API V1] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	} else {
		organization := new(models.Organization)

		if _, _, err := organization.Has(this.Ctx.Input.Param(":org")); err != nil {
			beego.Error("[WEB API V1] Get organizations error:", err.Error())

			result := map[string]string{"message": "Get organizations error."}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			return
		}

		this.Data["json"] = organization

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
		return
	}
}

func (this *OrganizationWebV1Controller) GetOrganizationRepo() {

	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API V1] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	org := new(models.Organization)

	if err := org.Get(this.Ctx.Input.Param(":org")); err != nil {
		beego.Error("[WEB API V1] Load session failure")

		result := map[string]string{"message": "Organization load failure"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	repositories := make([]models.Repository, 0)

	for _, repositoryUUID := range org.Repositories {
		repository := new(models.Repository)
		if err := repository.Get(repositoryUUID); err != nil {
			continue
		}
		repositories = append(repositories, *repository)
	}

	this.Data["json"] = repositories

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}
