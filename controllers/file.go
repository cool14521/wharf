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

func (this *FileController) Prepare() {
	this.EnableXSRF = false
}

func (this *FileController) GetGPG() {
	gpgPath := beego.AppConfig.String("rocket::GPG")

	if _, err := os.Stat(gpgPath); err != nil {
		beego.Error("[Rocket API] PGP file: ", err.Error())
		result := map[string]string{"Error": "PGP File State Error"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	file, err := ioutil.ReadFile(gpgPath)
	if err != nil {
		beego.Error("[Rocket API] PGP read file: ", err.Error())
		result := map[string]string{"Error": "PGP read file"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Transfer-Encoding", "binary")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Length", string(int64(len(file))))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(file)
	return
}
