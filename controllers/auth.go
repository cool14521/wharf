package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dockercn/wharf/models"
	"log"
	"net/http"
)

type Result struct {
	Success bool
	Message string
	Url     string
}

type AuthController struct {
	beego.Controller
}

func (this *AuthController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))

	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *AuthController) Get() {
	this.TplNames = "auth.html"

	this.Data["description"] = ""
	this.Data["author"] = ""

	this.Render()
}

func (this *AuthController) Signup() {
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	var result Result

	//form的attr name:username,email,password,password_confirm
	user := new(models.User)
	if err := user.Put(this.GetString("username"), this.GetString("password"), this.GetString("email")); err != nil {
		result = Result{Success: false, Message: fmt.Sprint(err), Url: "/auth"}
		this.Data["json"] = &result
		this.ServeJson()
		return
	}
	log.Println("数据存储成功")
}
