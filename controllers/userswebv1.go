package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"

	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"
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

func (this *UserWebAPIV1Controller) PostGravatar() {

	file, fileHeader, err := this.Ctx.Request.FormFile("file")
	if err != nil {
		beego.Error(fmt.Sprintf("[image] upload gravatar err,err=%s", err))

		result := map[string]string{"message": "Upload gravatar failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	prefix := strings.Split(fileHeader.Filename, ".")[0]
	suffix := strings.Split(fileHeader.Filename, ".")[1]
	if suffix != "png" && suffix != "jpg" && suffix != "jpeg" {
		result := map[string]string{"message": "gravatar must be .jpg,.jepg or .png", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	if _, err := os.Stat(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix)); err == nil {
		os.Remove(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix))
	}
	f, err := os.OpenFile(fmt.Sprintf("%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", fileHeader.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		result := map[string]string{"message": "Upload gravatar failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}
	io.Copy(f, file)
	f.Close()

	// decode jpeg into image.Image
	var img image.Image
	imageFile, err := os.Open(fmt.Sprintf("%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", fileHeader.Filename))
	if err != nil {
		result := map[string]string{"message": "Upload gravatar failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}
	switch suffix {
	case "png":
		img, err = png.Decode(imageFile)
	case "jpg":
		img, err = jpeg.Decode(imageFile)
	case "jpeg":
		img, err = jpeg.Decode(imageFile)
	}
	if err != nil {
		imageFile.Close()
		result := map[string]string{"message": "Upload gravatar resize failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}
	imageFile.Close()
	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(100, 100, img, resize.Lanczos3)

	out, err := os.Create(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix))
	if err != nil {
		result := map[string]string{"message": "Upload gravatar resize failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}
	defer out.Close()
	// write new image to file
	switch suffix {
	case "png":
		png.Encode(out, m)
	case "jpg":
		jpeg.Encode(out, m, nil)
	case "jpeg":
		jpeg.Encode(out, m, nil)
	}

	//删除用户自己上传的图片
	os.Remove(fmt.Sprintf("%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", fileHeader.Filename))

	result := map[string]string{"message": "Please click button to finish uploading gravatar", "url": fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix)}
	this.Data["json"] = &result
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	this.StopRun()
	return
}

func (this *UserWebAPIV1Controller) PutProfile() {

	var u map[string]interface{}
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &u); err != nil {
		beego.Error(fmt.Sprintf("[WEB API] JSON unmarshal failure: %s", err.Error()))
		result := map[string]string{"message": "Update User failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {

		beego.Error("[WEB API] Load session failure")

		result := map[string]string{"message": "Session load failure", "url": "/auth"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()

	}

	if strings.Contains(fmt.Sprint(u["gravatar"]), "resize") {

		suffix := strings.Split(fmt.Sprint(u["gravatar"]), ".")[1]
		gravatar := fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", user.Username, "_show.", suffix)
		if _, err := os.Stat(gravatar); err == nil {
			os.Remove(gravatar)
		}

		os.Rename(fmt.Sprint(u["gravatar"]), gravatar)
		u["gravatar"] = gravatar
	}

	user.Email = u["email"].(string)
	user.Fullname = u["fullname"].(string)
	user.Mobile = u["mobile"].(string)
	user.Gravatar = u["gravatar"].(string)
	user.Company = u["company"].(string)
	user.URL = u["url"].(string)

	if err := user.Save(); err != nil {
		result := map[string]string{"message": "User save failure"}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		this.StopRun()
	}

	this.Ctx.Input.CruSession.Set("user", user)

	result := map[string]string{"message": "Success!"}
	this.Data["json"] = &result
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	this.StopRun()
}
