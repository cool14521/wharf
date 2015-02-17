package controllers

import (
	"github.com/astaxie/beego"
)

type RepoW1Controller struct {
	beego.Controller
}

func (u *RepoW1Controller) URLMapping() {
}

func (this *RepoW1Controller) Prepare() {
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}
