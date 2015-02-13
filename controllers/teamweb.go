package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dockercn/wharf/models"
	"net/http"
)

type TeamWebController struct {
	beego.Controller
}

func (this *TeamWebController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *TeamWebController) PostTeam() {
	//获得用户提交的team相关信息
	var team models.Team

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &team); err != nil {
		beego.Error(fmt.Sprintf(`{"error":"%s"`, err.Error))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error)))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("[Web 用户] 用户增加团队: %s", string(this.Ctx.Input.CopyBody())))

	fmt.Printf("%#v", team)
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "Ok")))
	return
}
