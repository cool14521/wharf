package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/astaxie/beego"

	"github.com/containerops/wharf/models"
	"github.com/containerops/wharf/utils"
)

type OrganizationWebV1Controller struct {
	beego.Controller
}

func (this *OrganizationWebV1Controller) URLMapping() {
	this.Mapping("GetOrgs", this.GetOrgs)
	this.Mapping("GetJoins", this.GetJoins)
	this.Mapping("PostOrg", this.PostOrg)
	//	this.Mapping("PutOrganization", this.PutOrganization)
	//	this.Mapping("GetOrganizationRepositories", this.GetRepositories)
	//	this.Mapping("GetTeams", this.GetTeams)
	//	this.Mapping("PostTeam", this.PostTeam)
	//	this.Mapping("PutTeam", this.PutTeam)
	//	this.Mapping("GetTeam", this.GetTeam)
	//	this.Mapping("PutTeamAddMember", this.PutTeamAction)
}

func (this *OrganizationWebV1Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		this.Data["json"] = map[string]string{"message": message}
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *OrganizationWebV1Controller) Prepare() {
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *OrganizationWebV1Controller) GetOrgs() {
	user := new(models.User)
	orgs := make([]models.Organization, 0)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

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

func (this *OrganizationWebV1Controller) GetJoins() {
	user := new(models.User)
	orgs := make([]models.Organization, 0)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	for _, name := range user.JoinOrganizations {
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

func (this *OrganizationWebV1Controller) PostOrg() {
	user := new(models.User)
	org := new(models.Organization)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

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

func (this *OrganizationWebV1Controller) GetTeamsGetOrganizationDetail() {
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

func (this *OrganizationWebV1Controller) GetRepositories() {
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
