package controllers

import (
	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (this *MainController) Prepare() {
}

func (this *MainController) Get() {
	this.Layout = "default.html"
	this.TplNames = "index.html"
	this.LayoutSections = make(map[string]string)
	this.LayoutSections["HtmlHead"] = "index/head.html"
	this.LayoutSections["Header"] = "header.html"
	this.LayoutSections["Footer"] = "footer.html"
	this.Data["Title"] = "Hub"
	this.Render()
}
