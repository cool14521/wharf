package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/utils"
)

type BlobAPIV2Controller struct {
	beego.Controller
}

func (this *BlobAPIV2Controller) URLMapping() {
}

func (this *BlobAPIV2Controller) Prepare() {
	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)
	beego.Debug(this.Ctx.Request.URL)

	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

//Has image return 200; other return 404
func (this *BlobAPIV2Controller) HeadDigest() {
	this.Ctx.Output.Context.Output.SetStatus(http.StatusNotFound)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *BlobAPIV2Controller) PostBlobs() {
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Location", "https://containerops.me/v2/genedna/ubuntu/blobs/uploads/bWhQZUhtdWlodnFyaWU5bXlJbHd5NEx1Mkc5MzUydUo")
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Range", "bytes=0-0")
	this.Ctx.Output.Context.Output.SetStatus(http.StatusAccepted)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *BlobAPIV2Controller) PutBlobs() {
	data, _ := ioutil.ReadAll(this.Ctx.Request.Body)

	if err := ioutil.WriteFile(fmt.Sprintf("/tmp/%s", utils.GeneralKey("abc")), data, 0777); err != nil {
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusCreated)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}
