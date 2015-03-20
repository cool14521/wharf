package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type UserAPIV1Controller struct {
	beego.Controller
}

func (this *UserAPIV1Controller) URLMapping() {
	this.Mapping("PostUsers", this.PostUsers)
	this.Mapping("GetUsers", this.GetUsers)
}

func (this *UserAPIV1Controller) Prepare() {
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("docker::Standalone"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("docker::Version"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Config", beego.AppConfig.String("docker::Config"))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Encrypt", beego.AppConfig.String("docker::Encrypt"))
}

func (this *UserAPIV1Controller) PostUsers() {
	result := map[string]string{"error": "Don't support create user from docker command."}
	this.Data["json"] = &result

	this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
	this.ServeJson()
	return
}

//There is nothing in request body, just authorization through Basic Authorization.
func (this *UserAPIV1Controller) GetUsers() {
	if username, passwd, err := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization")); err != nil {

		beego.Error("[REGISTRY API V1] Decode Basic Auth Error:", err.Error())

		result := map[string]string{"error": "Decode authorization failure."}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.ServeJson()
		return
	} else {
		user := new(models.User)

		if err := user.Get(username, passwd); err != nil {
			beego.Error("[REGISTRY API V1] Search user error: ", err.Error())

			result := map[string]string{"error": "User authorization failure."}
			this.Data["json"] = &result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
			this.ServeJson()
			return
		}

		beego.Info("[REGISTRY API V1]", username, "authorization successfully")

		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := user.Log(models.ACTION_SIGNUP, models.LEVELINFORMATIONAL, models.TYPE_APIV1, user.Id, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Erro:", err.Error())
		}

		result := map[string]string{"status": "User authorization successfully."}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
		return
	}
}
