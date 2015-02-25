package controllers

import (
	"encoding/json"
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
}

func (this *TeamWebV1Controller) Prepare() {
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *TeamWebV1Controller) PostTeam() {

	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()

	}

	var team models.Team

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
		beego.Error("[WEB API] Unmarshal team data error.", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()

	}

	beego.Info("[Web API] Add team successfully: ", string(this.Ctx.Input.CopyBody()))

	team.UUID = utils.GeneralToken(team.Team)
	team.Username = user.Username

	if err := team.Save(); err != nil {
		beego.Error("[WEB API] Team save error:", err.Error())

		result := map[string]string{"message": "Team save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	org := new(models.Organization)

	if exist, _, _ := org.Has(team.Organization); exist {
		org.Teams = append(org.Teams, team.UUID)
	}

	if err := org.Save(); err != nil {
		beego.Error("[WEB API] team save error:", err.Error())

		result := map[string]string{"message": "team save error."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	result := map[string]string{"message": "OK"}
	this.Data["json"] = result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	this.StopRun()
}
