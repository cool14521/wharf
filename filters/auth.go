package filters

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/modules"
	"github.com/dockercn/wharf/utils"
)

func FilterAuth(ctx *context.Context) {
	auth := true
	user := new(models.User)

	//Check Authorization In Header
	if len(ctx.Input.Header("Authorization")) == 0 || strings.Index(ctx.Input.Header("Authorization"), "Basic") == -1 {
		beego.Debug("[Docker Registry AP] Header Authorization Error!")
		auth = false
	}

	//Check Username And Password
	if username, passwd, err := utils.DecodeBasicAuth(ctx.Input.Header("Authorization")); err != nil {
		beego.Debug("[Docker Registry AP] DecodeBasicAuth Error!")
		auth = false
	} else {
		if err := user.Get(username, passwd); err != nil {
			beego.Debug("[Docker Registry AP] Username And Password Check Error!")
			auth = false
		}
	}

	if auth == false {
		result := map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeUnauthorized]}}

		data, _ := json.Marshal(result)

		ctx.Output.Context.Output.SetStatus(http.StatusNotFound)
		ctx.Output.Context.Output.Body(data)
		return
	}

}
