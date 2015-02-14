package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type OrganizationWebController struct {
	beego.Controller
}

func (this *OrganizationWebController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *OrganizationWebController) PostOrganization() {
	user, ok := this.Ctx.Input.CruSession.Get("user").(models.User)
	if !ok {
		beego.Error(fmt.Sprintf("[WEB 用户] session加载失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "session加载失败，无法新建组织")))
		return
	}

	var org models.Organization

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &org); err != nil {
		beego.Error(fmt.Sprintf(`{"error":"%s"`, err.Error))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error)))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("[Web 用户] 用户增加组织: %s", string(this.Ctx.Input.CopyBody())))

	//生成UUID
	org.UUID = utils.GeneralToken(org.Organization)

	//关联user
	org.Username = user.Username
	if err := org.Save(); err != nil {
		beego.Error(fmt.Sprintf(`{"error":"%s"`, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error())))
		return
	}

	//更新user中organzation slice中的值
	user.Organizations = append(user.Organizations, org.UUID)

	//保存user
	if err := user.Save(); err != nil {
		beego.Error(fmt.Sprintf(`{"error":"%s"`, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error())))
		return
	}

	//获取最新User
	user.Get(user.Username, user.Password)

	this.Ctx.Input.CruSession.Set("user", user)

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "新建组织成功！")))
	return
}

func (this *OrganizationWebController) PutOrganization() {
	//权限控制（操作权限，仓库是否存在）

	//更新仓库
	var org models.Organization

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &org); err != nil {
		beego.Error(fmt.Sprintf(`{"error":"%s"`, err.Error))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error)))
		this.StopRun()
	}

	beego.Debug(fmt.Sprintf("[Web 用户] 用户更新组织: %s", string(this.Ctx.Input.CopyBody())))

	if err := org.Save(); err != nil {
		beego.Error(fmt.Sprintf(`{"error":"%s"`, err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, err.Error())))
		return
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "更新组织成功！")))
	return
}

func (this *OrganizationWebController) GetOrganizations() {
	//获取session中的user
	user, ok := this.Ctx.Input.CruSession.Get("user").(models.User)
	if !ok {
		beego.Error(fmt.Sprintf("[WEB 用户] session加载失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "session加载失败")))
		return
	}
	organizations := make([]models.Organization, len(user.Organizations))
	for i, UUID := range user.Organizations {
		if err := organizations[i].Get(UUID); err != nil {
			beego.Error(fmt.Sprintf("[WEB 用户] 获取组织列表失败，err=", err))
			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "[WEB 用户] 获取组织列表失败")))
			return
		}
	}
	organizations4Json, err := json.Marshal(organizations)
	if err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] 组织Json序列化失败，err=", err))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "[WEB 用户] 组织Json序列化失败")))
		return
	}
	fmt.Println(string(organizations4Json))
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(organizations4Json)
	return

}

func (this *OrganizationWebController) GetOrganizationDetail() {
	organization := new(models.Organization)

	if _, _, err := organization.Has(this.Ctx.Input.Param(":orgName")); err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] 获取用户组织信息失败，err=", err))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "获取用户组织信息失败")))
		return
	}

	organization4Json, err := json.Marshal(organization)
	if err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] 组织Json序列化失败，err=", err))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte(fmt.Sprintf(`{"message":"%s"}`, "[WEB 用户] 组织Json序列化失败")))
		return
	}
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(organization4Json)
}
