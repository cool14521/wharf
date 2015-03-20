package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/satori/go.uuid"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type TeamWebV1Controller struct {
	beego.Controller
}

func (this *TeamWebV1Controller) URLMapping() {
	this.Mapping("PostTeam", this.PostTeam)
	this.Mapping("GetTeams", this.GetTeams)
}

func (this *TeamWebV1Controller) Prepare() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist {
		user.GetById(user.Id)
		this.Ctx.Input.CruSession.Set("user", user)
	}

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *TeamWebV1Controller) PostTeam() {
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)

	if exist != true {
		beego.Error("[WEB API V1] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	var team models.Team

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
		beego.Error("[WEB API V1] Unmarshal team data error.", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	beego.Info("[Web API V1] Add team successfully: ", string(this.Ctx.Input.CopyBody()))

	org := new(models.Organization)

	if exist, _, err := org.Has(team.Organization); err != nil {
		beego.Error("[WEB API V1] Organization singup error: ", err.Error())
		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	} else if exist == false {
		beego.Error("[WEB API V1] Organization don't exist:", user.Username)

		result := map[string]string{"message": "Organization don't exist."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	} else if exist == true {
	}

	team.Id = string(utils.GeneralKey(team.Name))
	team.Username = user.Username
	team.Users = append(team.Users, user.Id)

	if err := team.Save(); err != nil {
		beego.Error("[WEB API V1] Team save error:", err.Error())

		result := map[string]string{"message": "Team save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	user.Teams = append(user.Teams, team.Id)

	if err := user.Save(); err != nil {
		beego.Error("[WEB API V1] User save error:", err.Error())

		result := map[string]string{"message": "User save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	org.Teams = append(org.Teams, team.Name)

	if err := org.Save(); err != nil {
		beego.Error("[WEB API V1] team save error:", err.Error())

		result := map[string]string{"message": "team save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := team.Log(models.ACTION_ADD_TEAM, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo); err != nil {
		beego.Error("[WEB API V1] Log Erro:", err.Error())
	}
	if err := user.Log(models.ACTION_ADD_TEAM, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, team.Id, memo); err != nil {
		beego.Error("[WEB API V1] Log Erro:", err.Error())
	}

	//Reload User Data In Session
	user.Get(user.Username, user.Password)
	this.Ctx.Input.CruSession.Set("user", user)

	result := map[string]string{"message": "OK"}
	this.Data["json"] = result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *TeamWebV1Controller) GetTeams() {
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {

		beego.Error("[WEB API V1] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return

	}

	teams := make([]models.Team, 0)

	orgUUID := this.Ctx.Input.Param(":org")
	org := new(models.Organization)
	if err := org.Get(orgUUID); err != nil {
		beego.Error(fmt.Sprintf("[WEB API V1] Get organization error:: %s", err.Error()))
		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	for _, teamUUID := range user.Teams {
		team := new(models.Team)
		if err := team.Get(teamUUID); err != nil {
			beego.Error("[WEB API V1] team load failure,uuid=", teamUUID, err.Error())
			continue
		}

		if team.Organization != org.Name {
			continue
		}

		userObjects := make([]models.User, 0)
		for _, id := range team.Users {
			user := new(models.User)
			if err := user.GetById(id); err != nil {
				beego.Error("[WEB API] user load failure,uuid=", id, err.Error())
				continue
			}
			userObjects = append(userObjects, *user)
		}
		team.UserObjects = userObjects

		repositories := make([]models.Repository, 0)
		for _, privilegeUUID := range team.TeamPrivileges {
			privilege := new(models.Privilege)
			if err := privilege.Get(privilegeUUID); err != nil {
				continue
			}
			repository := new(models.Repository)
			if err := repository.Get(privilege.Repository); err != nil {
				continue
			}
			repository.Privilege = *privilege
			repositories = append(repositories, *repository)
		}

		team.RepositoryObjects = repositories
		teams = append(teams, *team)
	}
	this.Data["json"] = teams

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *TeamWebV1Controller) PostPrivilege() {
	var repo map[string]interface{}

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		beego.Error(fmt.Sprintf("[WEB API V1] Unmarshal Repository create data error:: %s", err.Error()))
		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	privilege := repo["privilege"].(bool)
	teamUUID := repo["teamUUID"].(string)
	repoUUID := repo["repoUUID"].(string)

	privilegeObj := new(models.Privilege)
	privilegeObj.Id = string(utils.GeneralKey(uuid.NewV4().String()))
	privilegeObj.Privilege = privilege
	privilegeObj.Team = teamUUID
	privilegeObj.Repository = repoUUID

	if err := privilegeObj.Save(); err != nil {

		beego.Error("[WEB API V1] Privilege save error:", err.Error())
		result := map[string]string{"message": "Privilege save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	team := new(models.Team)
	if err := team.Get(teamUUID); err != nil {
		beego.Error("[WEB API V1] Team get error:", err.Error())
		result := map[string]string{"message": "Team get error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}
	team.Repositories = append(team.Repositories, repoUUID)
	team.TeamPrivileges = append(team.TeamPrivileges, privilegeObj.Id)

	if err := team.Save(); err != nil {
		beego.Error("[WEB API V1] Team save error:", err.Error())
		result := map[string]string{"message": "Team save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := team.Log(models.ACTION_ADD_TEAM, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, team.Id, memo); err != nil {
		beego.Error("[WEB API V1] Log Erro:", err.Error())
	}

	result := map[string]string{"message": "Privilege create successfully!"}
	this.Data["json"] = result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *TeamWebV1Controller) GetTeam() {
	team := new(models.Team)

	if err := team.Get(this.Ctx.Input.Param(":uuid")); err != nil {
		beego.Error("[WEB API] Team get error:", err.Error())
		result := map[string]string{"message": "Team get error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	userObjects := make([]models.User, 0)
	for _, id := range team.Users {
		user := new(models.User)
		if err := user.GetById(id); err != nil {
			beego.Error("[WEB API] user load failure,uuid=", id, err.Error())
			continue
		}
		userObjects = append(userObjects, *user)
	}
	team.UserObjects = userObjects

	this.Data["json"] = team

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *TeamWebV1Controller) PutTeam() {

	_, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return

	}

	var team models.Team

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
		beego.Error("[WEB API] Unmarshal team data error.", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return

	}

	beego.Info("[Web API] start update team", string(this.Ctx.Input.CopyBody()))

	org := new(models.Organization)
	if exist, _, _ := org.Has(team.Organization); !exist {
		beego.Error("[WEB API] Organization load error.")

		result := map[string]string{"message": "Organization load error"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	teamOld := new(models.Team)
	if err := teamOld.Get(this.Ctx.Input.Param(":uuid")); err != nil {
		beego.Error("[WEB API]team load error.", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	} else {
		for _, id := range teamOld.Users {
			user := new(models.User)
			if err := user.GetById(id); err == nil {
				joinOrganizations := make([]string, 0)
				for _, joinOrganization := range user.JoinOrganizations {
					if joinOrganization == org.Id {
						continue
					}
					joinOrganizations = append(joinOrganizations, joinOrganization)
				}

				joinTeams := make([]string, 0)
				for _, joinTeam := range user.JoinTeams {
					if joinTeam == teamOld.Id {
						continue
					}
					joinTeams = append(joinTeams, joinTeam)
				}

				user.JoinOrganizations = joinOrganizations
				user.JoinTeams = joinTeams
				if err := user.Save(); err != nil {
					beego.Error("[WEB API V1] user save error:", err.Error())
				}
			} else {
				beego.Error("[WEB API]User found err,err:=", err.Error())
				continue
			}
		}
	}

	usersUUID := make([]string, 0)
	for _, username := range team.Users {
		user := new(models.User)
		if _, UUID, err := user.Has(username); len(string(UUID)) > 0 && err == nil {
			usersUUID = append(usersUUID, string(UUID))

			user.JoinOrganizations = append(user.JoinOrganizations, org.Id)
			user.JoinTeams = append(user.JoinTeams, team.Id)
			if len(user.Gravatar) == 0 {
				user.Gravatar = "/static/images/default_user.jpg"
			}
			if err := user.Save(); err != nil {
				beego.Error("[WEB API V1] user save error:", err.Error())
			}
		} else {
			beego.Error("[WEB API]User found err,err:=", err.Error())
			continue
		}
	}

	team.Id = teamOld.Id
	team.Username = teamOld.Username
	team.Users = usersUUID
	team.TeamPrivileges = teamOld.TeamPrivileges
	team.Repositories = teamOld.Repositories
	team.Memo = teamOld.Memo

	if err := team.Save(); err != nil {
		beego.Error("[WEB API]team save error.", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	result := map[string]string{"message": "OK"}
	this.Data["json"] = result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}
