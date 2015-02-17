package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
)

type TeamWebV1Controller struct {
	beego.Controller
}

func (u *TeamWebV1Controller) URLMapping() {
	u.Mapping("PostTeam", u.PostTeam)
}

func (this *TeamWebV1Controller) Prepare() {
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *TeamWebV1Controller) PostTeam() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()

	} else {
		var team models.Team

		if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
			beego.Error("[WEB API] Unmarshal team data error.", err.Error())

			result := map[string]string{"message": err.Error()}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()

		} else {
			beego.Info("[Web API] Add team successfully: ", string(this.Ctx.Input.CopyBody()))

			result := map[string]string{"message": "OK"}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.ServeJson()

		}
	}
}
