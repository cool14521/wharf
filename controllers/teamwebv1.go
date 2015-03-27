package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type TeamWebV1Controller struct {
	beego.Controller
}

func (this *TeamWebV1Controller) URLMapping() {
	this.Mapping("PostTeam", this.PostTeam)
	this.Mapping("GetTeams", this.GetTeams)
	this.Mapping("GetTeam", this.GetTeam)
}

func (this *TeamWebV1Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		result := map[string]string{"message": message}
		this.Data["json"] = result
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *TeamWebV1Controller) Prepare() {
	this.EnableXSRF = false

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
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	var team models.Team

	beego.Trace("[WEB API V1] Create a team:", string(this.Ctx.Input.CopyBody()))

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
		beego.Error("[WEB API V1] Unmarshal team data error.", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	beego.Info("[Web API V1] Create team: ", string(this.Ctx.Input.CopyBody()))

	org := new(models.Organization)

	if exist, _, err := org.Has(team.Organization); err != nil {
		beego.Error("[WEB API V1] Organization singup error: ", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		beego.Error("[WEB API V1] Organization don't exist:", user.Username)
		this.JSONOut(http.StatusBadRequest, "Organization don't exist", nil)
		return
	}

	team.Id = string(utils.GeneralKey(team.Name))
	team.Username = user.Username
	team.Users = append(team.Users, user.Username)

	if err := team.Save(); err != nil {
		beego.Error("[WEB API V1] Team save error:", err.Error())
		this.JSONOut(http.StatusBadRequest, "Team save error", nil)
		return
	}

	user.Teams = append(user.Teams, team.Name)
	user.JoinTeams = append(user.JoinTeams, team.Name)

	if err := user.Save(); err != nil {
		beego.Error("[WEB API V1] User save error:", err.Error())
		this.JSONOut(http.StatusBadRequest, "User save error", nil)
		return
	}

	beego.Trace("[WEB API V1] User teams:", user.Teams)

	org.Teams = append(org.Teams, team.Name)

	if err := org.Save(); err != nil {
		beego.Error("[WEB API V1] Org save error:", err.Error())
		this.JSONOut(http.StatusBadRequest, "Org save error", nil)
		return
	}

	beego.Trace("[WEB API V1] Org teams:", org.Teams)

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

	this.JSONOut(http.StatusOK, "Team Create Successfully!", nil)
	return
}

func (this *TeamWebV1Controller) GetTeams() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		beego.Error("[WEB API V1] Load session failure")
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	teams := make([]models.Team, 0)

	org := new(models.Organization)
	if err := org.GetByName(this.Ctx.Input.Param(":org")); err != nil {
		beego.Error(fmt.Sprintf("[WEB API V1] Get organization error:: %s", err.Error()))
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	for _, name := range org.Teams {
		beego.Trace("[WEB API V1] Team Name:", name)
		team := new(models.Team)
		if err := team.GetByName(org.Name, name); err != nil {
			beego.Error("[WEB API V1] team load failure:", err.Error())
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		}

		teams = append(teams, *team)
	}

	this.JSONOut(http.StatusOK, "", teams)
	return
}

func (this *TeamWebV1Controller) PostPrivilege() {

}

func (this *TeamWebV1Controller) GetTeam() {
	team := new(models.Team)

	if err := team.GetByName(this.Ctx.Input.Param(":org"), this.Ctx.Input.Param(":team")); err != nil {
		beego.Error("[WEB API V1] Team get error:", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	this.Data["json"] = &team

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *TeamWebV1Controller) PutTeamAddMember() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		beego.Error("[WEB API V1] Load session failure")
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	org := new(models.Organization)
	if err := org.GetByName(this.Ctx.Input.Param(":org")); err != nil {
		beego.Error(fmt.Sprintf("[WEB API V1] Get organization error: %s", err.Error()))
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	team := new(models.Team)
	if err := team.GetByName(this.Ctx.Input.Param(":org"), this.Ctx.Input.Param(":team")); err != nil {
		beego.Error("[WEB API V1] Get team error:", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	beego.Trace("[WEB API V1] Add user:", this.Ctx.Input.Param(":username"))

	user := new(models.User)
	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		beego.Error("[WEB API V1] Search user error:", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		beego.Error("[WEB API V1] Search user none:", this.Ctx.Input.Param(":username"))
		this.JSONOut(http.StatusBadRequest, "User not found", nil)
		return
	}

	exist := false
	for _, u := range team.Users {
		if u == this.Ctx.Input.Param(":username") {
			exist = true
		}
	}

	if exist == true {
		beego.Error("[WEB API V1] User already in team", this.Ctx.Input.Param(":username"))
		this.JSONOut(http.StatusBadRequest, "User already in team", nil)
		return
	} else {
		team.Users = append(team.Users, this.Ctx.Input.Param(":username"))

		if err := team.Save(); err != nil {
			beego.Error("[WEB API]team save error.", err.Error())
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		}

		this.JSONOut(http.StatusOK, "", user)
		return
	}
}

func (this *TeamWebV1Controller) PutTeamRemoveMember() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		beego.Error("[WEB API] Load session failure")
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	org := new(models.Organization)
	if err := org.GetByName(this.Ctx.Input.Param(":org")); err != nil {
		beego.Error(fmt.Sprintf("[WEB API V1] Get organization error:: %s", err.Error()))
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	team := new(models.Team)
	if err := team.GetByName(this.Ctx.Input.Param(":org"), this.Ctx.Input.Param(":team")); err != nil {
		beego.Error("[WEB API V1] Get team error:", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	beego.Trace("[WEB API V1] Add user:", this.Ctx.Input.Param(":username"))

	user := new(models.User)
	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		beego.Error("[WEB API V1] Search user error:", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		beego.Error("[WEB API V1] Search user none:", this.Ctx.Input.Param(":username"))
		this.JSONOut(http.StatusBadRequest, "User not found", nil)
		return
	}

	for k, v := range team.Users {
		if v == this.Ctx.Input.Param(":username") {
			team.Users = append(team.Users[:k], team.Users[k+1:]...)

			if err := team.Save(); err != nil {
				beego.Error("[WEB API V1] Team save error.", err.Error())
				this.JSONOut(http.StatusBadRequest, err.Error(), nil)
				return
			}

			this.JSONOut(http.StatusOK, "", user)
			return
		}
	}

	beego.Error("[WEB API V1] Could not found user.")

	this.JSONOut(http.StatusBadRequest, "Could not found user", nil)
	return

}

func (this *TeamWebV1Controller) PutTeam() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		beego.Error("[WEB API] Load session failure")
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	}

	team := new(models.Team)
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
		beego.Error("[WEB API] Unmarshal team data error.", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := team.Save(); err != nil {
		beego.Error("[WEB API]team save error.", err.Error())
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	this.JSONOut(http.StatusOK, "Team update successfully!", nil)
	return
}
