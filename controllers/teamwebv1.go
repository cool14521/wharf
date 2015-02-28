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
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

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
		this.StopRun()

	}

	var team models.Team

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
		beego.Error("[WEB API V1] Unmarshal team data error.", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()

	}

	beego.Info("[Web API V1] Add team successfully: ", string(this.Ctx.Input.CopyBody()))

	team.UUID = string(utils.GeneralKey(team.Team))
	team.Username = user.Username

	org := new(models.Organization)

	if exist, _, _ := org.Has(team.Organization); exist {
		org.Teams = append(org.Teams, team.UUID)
	}

	if err := org.Save(); err != nil {
		beego.Error("[WEB API V1] team save error:", err.Error())

		result := map[string]string{"message": "team save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	usersUUID := make([]string, 0)
	for _, username := range team.Users {
		user := new(models.User)
		if _, UUID, _ := user.Has(username); len(string(UUID)) > 0 {
			usersUUID = append(usersUUID, string(UUID))

			user.JoinOrganizations = append(user.JoinOrganizations, org.UUID)
			user.JoinTeams = append(user.JoinTeams, team.UUID)

			if err := user.Save(); err != nil {
				beego.Error("[WEB API V1] user save error:", err.Error())
			}
		}

		continue
	}

	team.Users = usersUUID

	if err := team.Save(); err != nil {
		beego.Error("[WEB API V1] Team save error:", err.Error())

		result := map[string]string{"message": "Team save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	user.Teams = append(user.Teams, team.UUID)

	if err := user.Save(); err != nil {
		beego.Error("[WEB API V1] User save error:", err.Error())

		result := map[string]string{"message": "User save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := team.Log(models.ACTION_ADD_TEAM, models.LEVELINFORMATIONAL, models.TYPE_WEB, user.UUID, memo); err != nil {
		beego.Error("[WEB API V1] Log Erro:", err.Error())
	}
	if err := user.Log(models.ACTION_ADD_TEAM, models.LEVELINFORMATIONAL, models.TYPE_WEB, team.UUID, memo); err != nil {
		beego.Error("[WEB API V1] Log Erro:", err.Error())
	}

	user.Get(user.Username, user.Password)
	this.Ctx.Input.CruSession.Set("user", user)

	result := map[string]string{"message": "OK"}
	this.Data["json"] = result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	this.StopRun()
}

func (this *TeamWebV1Controller) GetTeams() {
	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {

		beego.Error("[WEB API V1] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()

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
		this.StopRun()
	}

	for _, teamUUID := range user.Teams {
		team := new(models.Team)
		if err := team.Get(teamUUID); err != nil {
			beego.Error("[WEB API V1] team load failure,uuid=", teamUUID, err.Error())
			continue
		}

		if team.Organization != org.Organization {
			continue
		}

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
	this.StopRun()
}

func (this *TeamWebV1Controller) PostPrivilege() {
	var repo map[string]interface{}

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		beego.Error(fmt.Sprintf("[WEB API V1] Unmarshal Repository create data error:: %s", err.Error()))
		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	privilege := repo["privilege"].(bool)
	teamUUID := repo["teamUUID"].(string)
	repoUUID := repo["repoUUID"].(string)

	privilegeObj := new(models.Privilege)
	privilegeObj.UUID = string(utils.GeneralKey(uuid.NewV4().String()))
	privilegeObj.Privilege = privilege
	privilegeObj.Team = teamUUID
	privilegeObj.Repository = repoUUID

	if err := privilegeObj.Save(); err != nil {

		beego.Error("[WEB API V1] Privilege save error:", err.Error())
		result := map[string]string{"message": "Privilege save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	team := new(models.Team)
	if err := team.Get(teamUUID); err != nil {
		beego.Error("[WEB API V1] Team get error:", err.Error())
		result := map[string]string{"message": "Team get error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}
	team.Repositories = append(team.Repositories, repoUUID)
	team.TeamPrivileges = append(team.TeamPrivileges, privilegeObj.UUID)

	if err := team.Save(); err != nil {
		beego.Error("[WEB API V1] Team save error:", err.Error())
		result := map[string]string{"message": "Team save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	if err := team.Log(models.ACTION_ADD_TEAM, models.LEVELINFORMATIONAL, models.TYPE_WEB, team.UUID, memo); err != nil {
		beego.Error("[WEB API V1] Log Erro:", err.Error())
	}

	result := map[string]string{"message": "Privilege create successfully!"}
	this.Data["json"] = result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	this.StopRun()
}
