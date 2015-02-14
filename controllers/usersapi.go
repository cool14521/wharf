package controllers

import (
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type UsersAPIController struct {
	beego.Controller
}

func (u *UsersAPIController) URLMapping() {
	u.Mapping("PostUsers", u.PostUsers)
	u.Mapping("GetUsers", u.GetUsers)
}

func (this *UsersAPIController) Prepare() {
	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)

	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))
}

func (this *UsersAPIController) PostUsers() {
	result := map[string]string{"error": "Don't support create user from docker command."}
	this.Data["json"] = &result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
	this.ServeJson()
}

func (this *UsersAPIController) GetUsers() {
	if username, passwd, err := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization")); err != nil {

		beego.Error("[USER API] Decode Basic Auth Error:", err.Error())

		result := map[string]string{"error": "Decode authorization failure."}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.ServeJson()

	} else {

		user := new(models.User)

		if user.Get(username, passwd); err != nil {
			beego.Error("[USER API] Search user error: ", err.Error())

			result := map[string]string{"error": "User authorization failure."}
			this.Data["json"] = &result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
			this.ServeJson()
		}

		beego.Info("[User API]", username, "authorization successfully")

		result := map[string]string{"status": "User authorization successfully."}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
	}
}
