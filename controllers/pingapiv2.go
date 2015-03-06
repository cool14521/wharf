package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type PingAPIV2Controller struct {
	beego.Controller
}

func (this *PingAPIV2Controller) URLMapping() {
	this.Mapping("GetPing", this.GetPing)
}

func (this *PingAPIV2Controller) Prepare() {

	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)

	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"\"", beego.AppConfig.String("docker::Endpoints")))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
}

func (this *PingAPIV2Controller) GetPing() {
	result := map[string]bool{"result": true}
	this.Data["json"] = &result

	if len(this.Ctx.Input.Header("Authorization")) == 0 {
		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)

		this.ServeJson()
		return
	}

	if username, passwd, err := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization")); err != nil {
		beego.Error("[REGISTRY API V2] Decode Basic Auth Error:", err.Error())

		result := map[string]string{"error": "Decode authorization failure."}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
		this.ServeJson()
		return
	} else {
		user := new(models.User)

		if err := user.Get(username, passwd); err != nil {
			beego.Error("[REGISTRY API V2] Search user error: ", err.Error())

			result := map[string]string{"error": "User authorization failure."}
			this.Data["json"] = &result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusUnauthorized)
			this.ServeJson()
			return
		}

		beego.Info("[REGISTRY API V2]", username, "authorization successfully")

		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := user.Log(models.ACTION_SIGNUP, models.LEVELINFORMATIONAL, models.TYPE_API, user.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V2] Log Erro:", err.Error())
		}

		result := map[string]string{"status": "User authorization successfully."}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
		return
	}
}
