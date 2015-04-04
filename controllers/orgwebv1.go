package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego"

	"github.com/containerops/wharf/models"
	"github.com/containerops/wharf/utils"
)

type OrganizationWebV1Controller struct {
	beego.Controller
}

func (this *OrganizationWebV1Controller) URLMapping() {

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

func (this *OrganizationWebV1Controller) GetJoinOrgs() {
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

func (this *OrganizationWebV1Controller) GetTeams() {
	user := new(models.User)
	teams := make([]models.Team, 0)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	for _, name := range user.Teams {
		team := new(models.Team)
		if exist, _, err := team.Has(strings.Split(name, "-")[0], strings.Split(name, "-")[1]); err != nil {
			this.JSONOut(http.StatusBadRequest, "Get Team error", nil)
			return
		} else if exist == false {
			this.JSONOut(http.StatusBadRequest, "Team invalid", nil)
			return
		}

		teams = append(teams, *team)
	}

	this.JSONOut(http.StatusOK, "", teams)
	return
}

func (this *OrganizationWebV1Controller) GetJoinTeams() {
	user := new(models.User)
	teams := make([]models.Team, 0)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	for _, name := range user.JoinTeams {
		team := new(models.Team)
		if exist, _, err := team.Has(strings.Split(name, "-")[0], strings.Split(name, "-")[1]); err != nil {
			this.JSONOut(http.StatusBadRequest, "Get organizations error", nil)
			return
		} else if exist == false {
			this.JSONOut(http.StatusBadRequest, "Team invalid", nil)
			return
		}

		teams = append(teams, *team)
	}

	this.JSONOut(http.StatusOK, "", teams)
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

	if exist, _, err := user.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == true {
		this.JSONOut(http.StatusBadRequest, "Namespace is occupation already by another user", nil)
		return
	}

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == true {
		this.JSONOut(http.StatusBadRequest, "Namespace is occupation already by another organization", nil)
		return
	}

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &org); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
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
	user.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := user.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "User save error", nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	user.Log(models.ACTION_ADD_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, org.Id, memo)
	org.Log(models.ACTION_ADD_ORG, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo)

	this.JSONOut(http.StatusOK, "Create organization successfully.", nil)
	return
}

func (this *OrganizationWebV1Controller) PutOrg() {
	user := new(models.User)
	org := new(models.Organization)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

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

func (this *OrganizationWebV1Controller) GetOrg() {
	org := new(models.Organization)

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

	this.JSONOut(http.StatusOK, "", org)
	return
}

func (this *OrganizationWebV1Controller) GetPublicRepos() {
	org := new(models.Organization)
	repos := make([]models.Repository, 0)

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

	for _, v := range org.Repositories {
		repo := new(models.Repository)

		if err := repo.Get(v); err != nil {
			this.JSONOut(http.StatusBadRequest, "Repository invalid", nil)
			return
		}

		if repo.Privated == false {
			repos = append(repos, *repo)
		}
	}

	this.JSONOut(http.StatusOK, "", repos)
	return
}

func (this *OrganizationWebV1Controller) GetPrivateRepos() {
	org := new(models.Organization)
	repos := make([]models.Repository, 0)

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

	for _, v := range org.Repositories {
		repo := new(models.Repository)

		if err := repo.Get(v); err != nil {
			this.JSONOut(http.StatusBadRequest, "Repository invalid", nil)
			return
		}

		if repo.Privated == true {
			repos = append(repos, *repo)
		}
	}

	this.JSONOut(http.StatusOK, "", repos)
	return
}

func (this *OrganizationWebV1Controller) GetOrgTeams() {
	org := new(models.Organization)
	teams := make([]models.Team, 0)

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

	for _, v := range org.Teams {
		team := new(models.Team)

		if exist, _, err := team.Has(strings.Split(v, "-")[0], strings.Split(v, "-")[1]); err != nil {
			this.JSONOut(http.StatusBadRequest, "Team invalid", nil)
			return
		} else if exist == false {
			this.JSONOut(http.StatusBadRequest, "Team invalid", nil)
			return
		}

		teams = append(teams, *team)
	}

	this.JSONOut(http.StatusOK, "", teams)
	return
}

func (this *OrganizationWebV1Controller) PostTeam() {
	user := new(models.User)
	org := new(models.Organization)
	team := new(models.Team)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

	if exist, _, err := team.Has(this.Ctx.Input.Param(":org"), this.Ctx.Input.Param(":team")); err != nil {
		this.JSONOut(http.StatusBadRequest, "Search team error", nil)
		return
	} else if exist == true {
		this.JSONOut(http.StatusBadRequest, "Team already exist", nil)
		return
	}

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	team.Id = fmt.Sprintf("%s-%s", this.Ctx.Input.Param(":org"), this.Ctx.Input.Param(":team"))
	team.Username = this.Ctx.Input.Param(":username")
	team.Users, team.Repositories = []string{this.Ctx.Input.Param(":username")}, []string{}
	team.Created = time.Now().UnixNano() / int64(time.Millisecond)
	team.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := team.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "Team save error", nil)
		return
	}

	org.Teams = append(org.Teams, team.Id)
	org.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := org.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "Organization save error", nil)
		return
	}

	user.Teams = append(user.Teams, team.Id)
	user.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := user.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "User save error", nil)
		return
	}

	this.JSONOut(http.StatusOK, "Team create successfully", nil)
	return
}

func (this *OrganizationWebV1Controller) PutTeam() {
	user := new(models.User)
	org := new(models.Organization)
	team := new(models.Team)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

	if exist, _, err := team.Has(this.Ctx.Input.Param(":org"), this.Ctx.Input.Param(":team")); err != nil {
		this.JSONOut(http.StatusBadRequest, "Search team error", nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Team not exist", nil)
		return
	}

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	team.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := team.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "Team save error", nil)
		return
	}

	this.JSONOut(http.StatusOK, "Team update successfully", nil)
	return
}

func (this *OrganizationWebV1Controller) GetTeam() {
	user := new(models.User)
	org := new(models.Organization)
	team := new(models.Team)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

	if exist, _, err := team.Has(this.Ctx.Input.Param(":org"), this.Ctx.Input.Param(":team")); err != nil {
		this.JSONOut(http.StatusBadRequest, "Search team error", nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Team not exist", nil)
		return
	}

	this.JSONOut(http.StatusOK, "", team)
	return
}

func (this *OrganizationWebV1Controller) PostMember() {
	user := new(models.User)
	org := new(models.Organization)
	team := new(models.Team)
	member := new(models.User)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

	if exist, _, err := team.Has(this.Ctx.Input.Param(":org"), this.Ctx.Input.Param(":team")); err != nil {
		this.JSONOut(http.StatusBadRequest, "Search team error", nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Team not exist", nil)
		return
	}

	if exist, _, err := user.Has(this.Ctx.Input.Param(":member")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	team.Users = append(team.Users, member.Username)
	team.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := team.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "Team save error", nil)
		return
	}

	has := false
	for _, k := range member.JoinOrganizations {
		if k == org.Name {
			has = true
		}
	}

	if has == false {
		member.JoinOrganizations = append(member.JoinOrganizations, org.Name)
	}

	member.JoinTeams = append(member.JoinTeams, team.Id)
	member.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := member.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "User save error", nil)
		return
	}

	this.JSONOut(http.StatusOK, "User added to the team", nil)
	return
}

func (this *OrganizationWebV1Controller) PutMember() {
	user := new(models.User)
	org := new(models.Organization)
	team := new(models.Team)
	member := new(models.User)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	if exist, _, err := org.Has(this.Ctx.Input.Param(":org")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization not exist", nil)
		return
	}

	if exist, _, err := team.Has(this.Ctx.Input.Param(":org"), this.Ctx.Input.Param(":team")); err != nil {
		this.JSONOut(http.StatusBadRequest, "Search team error", nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Team not exist", nil)
		return
	}

	if exist, _, err := user.Has(this.Ctx.Input.Param(":member")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "User not exist", nil)
		return
	}

	for i, v := range team.Users {
		if v == member.Username {
			team.Users = append(team.Users[:i], team.Users[i+1:]...)
			team.Updated = time.Now().UnixNano() / int64(time.Millisecond)

			if err := team.Save(); err != nil {
				this.JSONOut(http.StatusBadRequest, "Team save error", nil)
				return
			}
		}
	}

	for i, v := range user.JoinTeams {
		if v == team.Id {
			user.JoinTeams = append(user.JoinTeams[:i], user.JoinTeams[i+1:]...)
			user.Updated = time.Now().UnixNano() / int64(time.Millisecond)

			if err := user.Save(); err != nil {
				this.JSONOut(http.StatusBadRequest, "User save error", nil)
				return
			}
		}
	}

	for _, v := range org.Teams {
		t := new(models.Team)

		if exist, _, err := team.Has(strings.Split(v, "-")[0], strings.Split(v, "-")[1]); err != nil {
			this.JSONOut(http.StatusBadRequest, "Search team error", nil)
			return
		} else if exist == false {
			this.JSONOut(http.StatusBadRequest, "Team not exist", nil)
			return
		}

		for _, u := range t.Users {
			if u == member.Username {
				this.JSONOut(http.StatusOK, "User remove successfully.", nil)
				return
			}
		}
	}

	for i, v := range user.JoinOrganizations {
		if v == org.Name {
			user.JoinOrganizations = append(user.JoinOrganizations[:i], user.JoinOrganizations[i+1:]...)
			user.Updated = time.Now().UnixNano() / int64(time.Millisecond)

			if err := user.Save(); err != nil {
				this.JSONOut(http.StatusBadRequest, "User save error", nil)
				return
			}
		}
	}

	this.JSONOut(http.StatusOK, "User remove successfully.", nil)
	return
}
