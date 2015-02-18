package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type UserWebAPIV1Controller struct {
	beego.Controller
}

func (u *UserWebAPIV1Controller) URLMapping() {
	u.Mapping("GetProfile", u.GetProfile)
	u.Mapping("GetUser", u.GetUser)
	u.Mapping("Signup", u.Signup)
	u.Mapping("Signin", u.Signin)
}

func (this *UserWebAPIV1Controller) Prepare() {
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *UserWebAPIV1Controller) GetProfile() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()

	} else {

		this.Data["json"] = user

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
		this.StopRun()

	}
}

func (this *UserWebAPIV1Controller) GetUser() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()

	} else {

		user := new(models.User)

		if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {

			beego.Error("[WEB API] Search user error:", err.Error())

			result := map[string]string{"message": "Search user error"}
			this.Data["json"] = &result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			this.StopRun()

		} else if exist == false && err == nil {

			beego.Info("[WEB API] Search user none:", this.Ctx.Input.Param(":username"))

			result := map[string]string{"message": "Search user error"}
			this.Data["json"] = &result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			this.StopRun()

		} else {

			users := make([]models.User, 0)
			users = append(users, *user)

			this.Data["json"] = users
			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.ServeJson()
			this.StopRun()

		}
	}
}

func (this *UserWebAPIV1Controller) Signin() {
	var user models.User

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &user); err != nil {

		beego.Error("[WEB API] Unmarshal user signin data error:", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()

	} else {

		beego.Debug("[WEB API] User signin:", string(this.Ctx.Input.CopyBody()))

		if err := user.Get(user.Username, user.Password); err != nil {
			beego.Error("[WEB API] User singin error: ", err.Error())

			result := map[string]string{"message": err.Error()}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			this.StopRun()

		}

		if user.Gravatar == "" {
			user.Gravatar = "/static/images/default_user.jpg"
		}

		this.Ctx.Input.CruSession.Set("user", user)

		result := map[string]string{"message": "User Singin Successfully!"}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
		this.StopRun()

	}
}

func (this *UserWebAPIV1Controller) Signup() {
	var user models.User

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &user); err != nil {
		beego.Error("[WEB API] Unmarshal user signup data error:", err.Error())

		result := map[string]string{"message": err.Error()}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()

	} else {

		beego.Debug("[WEB API] User signup:", string(this.Ctx.Input.CopyBody()))

		if exist, _, err := user.Has(user.Username); err != nil {
			beego.Error("[WEB API] User singup error: ", err.Error())

			result := map[string]string{"message": err.Error()}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			this.StopRun()

		} else if exist == true {
			beego.Error("[WEB API] User already exist:", user.Username)

			result := map[string]string{"message": "User already exist."}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			this.StopRun()

		} else {
			user.UUID = string(utils.GeneralKey(user.Username))

			if err := user.Save(); err != nil {
				beego.Error("[WEB API] User save error:", err.Error())

				result := map[string]string{"message": "User save error."}
				this.Data["json"] = result

				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.ServeJson()
				this.StopRun()

			}

			result := map[string]string{"message": "User Singup Successfully!"}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			this.StopRun()

		}
	}
}
