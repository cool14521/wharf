package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/modules"
	"github.com/dockercn/wharf/utils"
)

type PingAPIV2Controller struct {
	beego.Controller
}

func (this *PingAPIV2Controller) URLMapping() {
	this.Mapping("GetPing", this.GetPing)
}

func (this *PingAPIV2Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		this.Data["json"] = map[string]string{"message": message}
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *PingAPIV2Controller) Prepare() {
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"\"", beego.AppConfig.String("docker::Endpoints")))
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
}

func (this *PingAPIV2Controller) GetPing() {
	if len(this.Ctx.Input.Header("Authorization")) == 0 {
		this.JSONOut(http.StatusUnauthorized, "", map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeUnauthorized]}})
		return
	}

	if username, passwd, err := utils.DecodeBasicAuth(this.Ctx.Input.Header("Authorization")); err != nil {
		this.JSONOut(http.StatusUnauthorized, "", map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeUnauthorized]}})
		return
	} else {
		user := new(models.User)

		if err := user.Get(username, passwd); err != nil {
			this.JSONOut(http.StatusUnauthorized, "", map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeUnauthorized]}})
			return
		}

		memo, _ := json.Marshal(this.Ctx.Input.Header)
		user.Log(models.ACTION_SIGNUP, models.LEVELINFORMATIONAL, models.TYPE_APIV2, user.Id, memo)

		this.JSONOut(http.StatusOK, "", "User authorization successfully.")
		return
	}
}
