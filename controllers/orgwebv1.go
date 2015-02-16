package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type OrganizationWebV1Controller struct {
	beego.Controller
}

func (u *OrganizationWebV1Controller) URLMapping() {
	u.Mapping("PostOrganization", u.PostOrganization)
	u.Mapping("PutOrganization", u.PutOrganization)
	u.Mapping("GetOrganizations", u.GetOrganizations)
	u.Mapping("GetOrganizationDetail", u.GetOrganizationDetail)
}

func (this *OrganizationWebV1Controller) Prepare() {
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *OrganizationWebV1Controller) PostOrganization() {

	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()

	} else {

		var org models.Organization

		if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &org); err != nil {

			beego.Error("[WEB API] Unmarshal organization data error:", err.Error())

			result := map[string]string{"message": err.Error()}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
		}

		beego.Debug("[WEB API] organization create: %s", string(this.Ctx.Input.CopyBody()))

		org.UUID = utils.GeneralToken(org.Organization)

		org.Username = user.Username

		if err := org.Save(); err != nil {
			beego.Error("[WEB API] Organization save error:", err.Error())

			result := map[string]string{"message": "Organization save error."}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()

			return
		}

		user.Organizations = append(user.Organizations, org.UUID)

		if err := user.Save(); err != nil {
			beego.Error("[WEB API] User save error:", err.Error())

			result := map[string]string{"message": "User save error."}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()

			return
		}

		user.Get(user.Username, user.Password)
		this.Ctx.Input.CruSession.Set("user", user)

		result := map[string]string{"message": "Create organization successfully."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
	}
}

func (this *OrganizationWebV1Controller) PutOrganization() {

	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()

	} else {

		var org models.Organization

		if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &org); err != nil {
			beego.Error("[WEB API] Unmarshal organization data error:", err.Error())

			result := map[string]string{"message": err.Error()}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
		}

		beego.Debug("[WEB API] organization update: %s", string(this.Ctx.Input.CopyBody()))

		if err := org.Save(); err != nil {
			beego.Error("[WEB API] Organization save error:", err.Error())

			result := map[string]string{"message": "Organization save error."}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
		}

		result := map[string]string{"message": "Update organization successfully."}
		this.Data["json"] = result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
	}
}

func (this *OrganizationWebV1Controller) GetOrganizations() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()

	} else {

		organizations := make([]models.Organization, len(user.Organizations))

		for i, UUID := range user.Organizations {
			if err := organizations[i].Get(UUID); err != nil {
				beego.Error("[WEB API] Get organizations error:", err.Error())

				result := map[string]string{"message": "Get organizations error."}
				this.Data["json"] = result

				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.ServeJson()
			}
		}

		this.Data["json"] = organizations

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
	}
}

func (this *OrganizationWebV1Controller) GetOrganizationDetail() {
	if _, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()

	} else {
		organization := new(models.Organization)

		if _, _, err := organization.Has(this.Ctx.Input.Param(":org")); err != nil {
			beego.Error("[WEB API] Get organizations error:", err.Error())

			result := map[string]string{"message": "Get organizations error."}
			this.Data["json"] = result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
		}

		this.Data["json"] = organization

		this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
		this.ServeJson()
	}
}
