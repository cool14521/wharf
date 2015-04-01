package controllers

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/astaxie/beego"
)

type FileController struct {
	beego.Controller
}

func (this *FileController) URLMapping() {
	this.Mapping("GetGPG", this.GetGPG)
}

func (this *FileController) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		this.Data["json"] = map[string]string{"message": message}
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *FileController) Prepare() {
	this.EnableXSRF = false
}

func (this *FileController) GetGPG() {
	gpgPath := beego.AppConfig.String("rocket::GPG")

	if _, err := os.Stat(gpgPath); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	file, err := ioutil.ReadFile(gpgPath)
	if err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Transfer-Encoding", "binary")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Length", string(int64(len(file))))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(file)
	return
}
