package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
)

type StaticController struct {
	beego.Controller
}

func (i *StaticController) URLMapping() {
	i.Mapping("GetFavicon", i.GetFavicon)
}

func (this *StaticController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Headers]")
	beego.Debug(this.Ctx.Input.Request.Header)
}

func (this *StaticController) GetFavicon() {
	this.Redirect("/static/images/favicon.ico", http.StatusTemporaryRedirect)
}
