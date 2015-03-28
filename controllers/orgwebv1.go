package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/astaxie/beego"

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

func (this *OrganizationWebV1Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		result := map[string]string{"message": message}
		this.Data["json"] = result
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *OrganizationWebV1Controller) Prepare() {
	this.EnableXSRF = false

	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist {
		user.GetById(user.Id)
		this.Ctx.Input.CruSession.Set("user", user)
	}

	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *OrganizationWebV1Controller) PostOrganization() {
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)

	if exist != true {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	var org models.Organization

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &org); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	if exist, _, err := user.Has(org.Name); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == true {
		this.JSONOut(http.StatusBadRequest, "Namespace is occupation already by another user", nil)
		return
	}

	if exist, _, err := org.Has(org.Name); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == true {
		this.JSONOut(http.StatusBadRequest, "Namespace is occupation already by another organization", nil)
		return
	}

	org.Id = string(utils.GeneralKey(org.Name))
	org.Username = user.Username
	org.Created = time.Now().UnixNano() / int64(time.Millisecond)
	org.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := org.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "Organization save error", nil)
		return
	}

	user.Organizations = append(user.Organizations, org.Name)
	user.JoinOrganizations = append(user.JoinOrganizations, org.Name)
	user.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := user.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "User save error", nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	user.Log(models.ACTION_ADD_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, org.Id, memo)
	org.Log(models.ACTION_ADD_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo)

	//Reload user data in session.
	user.Get(user.Username, user.Password)
	this.Ctx.Input.CruSession.Set("user", user)

	this.JSONOut(http.StatusOK, "Create organization successfully.", nil)
	return
}

func (this *OrganizationWebV1Controller) PutOrganization() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	} else {
		var org models.Organization

		if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &org); err != nil {
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		}

		org.Updated = time.Now().UnixNano() / int64(time.Millisecond)

		if err := org.Save(); err != nil {
			this.JSONOut(http.StatusBadRequest, "Organization save error", nil)
			return
		}

		memo, _ := json.Marshal(this.Ctx.Input.Header)
		user.Log(models.ACTION_UPDATE_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, org.Id, memo)
		org.Log(models.ACTION_UPDATE_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo)

		this.JSONOut(http.StatusOK, "Update organization successfully", nil)
		return
	}
}

func (this *OrganizationWebV1Controller) GetOrganizations() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	} else {
		orgs := make([]models.Organization, 0)

		for _, name := range user.Organizations {
			org := new(models.Organization)
			if err := org.GetByName(name); err != nil {
				this.JSONOut(http.StatusBadRequest, "Get organizations error", nil)
				return
			}

			orgs = append(orgs, *org)
		}

		this.JSONOut(http.StatusOK, "", orgs)
		return
	}
}

func (this *OrganizationWebV1Controller) GetOrganizationDetail() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	} else {
		organization := new(models.Organization)

		if _, _, err := organization.Has(this.Ctx.Input.Param(":org")); err != nil {
			this.JSONOut(http.StatusBadRequest, "Get organizations error", nil)
			return
		}

		this.JSONOut(http.StatusOK, "", organization)
		return
	}
}

func (this *OrganizationWebV1Controller) GetOrganizationRepo() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	org := new(models.Organization)

	if err := org.GetByName(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, "Get organizations error", nil)
		return
	}

	repositories := make([]models.Repository, 0)

	for _, repositoryId := range org.Repositories {
		repository := new(models.Repository)
		if err := repository.Get(repositoryId); err != nil {
			continue
		}
		repositories = append(repositories, *repository)
	}

	this.JSONOut(http.StatusOK, "", repositories)
	return
}
