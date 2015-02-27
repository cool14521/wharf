package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type UserWebAPIV1Controller struct {
	beego.Controller
}

func (this *UserWebAPIV1Controller) URLMapping() {
	this.Mapping("GetProfile", this.GetProfile)
	this.Mapping("GetUser", this.GetUser)
	this.Mapping("Signup", this.Signup)
	this.Mapping("Signin", this.Signin)
	this.Mapping("GetNamespaces", this.GetNamespaces)
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

		//memo, _ := json.Marshal(this.Ctx.Input.Header)
		//user.Log(models.ACTION_SIGNIN, models.LEVELINFORMATIONAL, user.UUID, memo)

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
			user.Created = time.Now().Unix()

			if err := user.Save(); err != nil {
				beego.Error("[WEB API] User save error:", err.Error())
				result := map[string]string{"message": "User save error."}
				this.Data["json"] = result

				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.ServeJson()
				this.StopRun()
			}

			memo, _ := json.Marshal(this.Ctx.Input.Header)
			user.Log(models.ACTION_SIGNUP, models.LEVELINFORMATIONAL, user.UUID, memo)

			result := map[string]string{"message": "User Singup Successfully!"}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
			this.ServeJson()
			this.StopRun()
		}
	}
}

type Namespace struct {
	Namespace     string `json:"namespace"`     //仓库所有者的名字
	NamespaceType bool   `json:"namespacetype"` // false 为普通用户，true为组织
}

func (this *UserWebAPIV1Controller) GetNamespaces() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {
		beego.Error("[WEB API] Load session failure")
		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	} else {
		namespaces := make([]Namespace, 0)
		namespaceUser := Namespace{Namespace: user.Username, NamespaceType: false}
		namespaces = append(namespaces, namespaceUser)

		orgs, _ := user.Orgs(user.Username)

		for k, _ := range orgs {
			namespaceOrg := Namespace{Namespace: k, NamespaceType: true}
			namespaces = append(namespaces, namespaceOrg)
		}

		this.Data["json"] = namespaces

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
		this.StopRun()
	}
}
