package controllers

import (
	"github.com/astaxie/beego"
	"net/http"
)

type StaticController struct {
	beego.Controller
}

func (i *StaticController) URLMapping() {
	i.Mapping("GetFavicon", i.GetFavicon)
}

func (this *StaticController) Prepare() {

}

func (this *StaticController) GetFavicon() {
	this.Redirect("/static/images/favicon.ico", http.StatusTemporaryRedirect)
}
