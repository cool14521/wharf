package controllers

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/nfnt/resize"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

type UserWebAPIV1Controller struct {
	beego.Controller
}

func (this *UserWebAPIV1Controller) URLMapping() {
	this.Mapping("Signin", this.Signin)
	this.Mapping("Signup", this.Signup)
	this.Mapping("GetUsers", this.GetUsers)
	this.Mapping("GetUser", this.GetUser)
	this.Mapping("GetNamespaces", this.GetNamespaces)
	this.Mapping("PostGravatar", this.PostGravatar)
	this.Mapping("PutPassword", this.PutPassword)
	this.Mapping("PutProfile", this.PutProfile)
}

func (this *UserWebAPIV1Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		result := map[string]string{"message": message}
		this.Data["json"] = result
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *UserWebAPIV1Controller) Prepare() {
	this.EnableXSRF = false

	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		user.GetById(user.Id)
		this.Ctx.Input.CruSession.Set("user", user)
	}

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *UserWebAPIV1Controller) Signin() {
	var user models.User

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &user); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else {
		if err := user.Get(user.Username, user.Password); err != nil {
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		}

		memo, _ := json.Marshal(this.Ctx.Input.Header)
		user.Log(models.ACTION_SIGNIN, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo)

		this.Ctx.Input.CruSession.Set("user", user)

		this.JSONOut(http.StatusOK, "User singin successfully!", nil)
		return
	}
}

func (this *UserWebAPIV1Controller) Signup() {
	var user models.User
	var org models.Organization

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &user); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else {
		if exist, _, err := org.Has(user.Username); err != nil {
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		} else if exist == true {
			this.JSONOut(http.StatusBadRequest, "Namespace is occupation already by organization.", nil)
			return
		}

		if exist, _, err := user.Has(user.Username); err != nil {
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		} else if exist == true {
			this.JSONOut(http.StatusBadRequest, "User already exist.", nil)
			return
		} else {
			user.Id = string(utils.GeneralKey(user.Username))
			user.Created = time.Now().UnixNano() / int64(time.Millisecond)
			user.Gravatar = "/static/images/default-user-icon-profile.png"

			if err := user.Save(); err != nil {
				this.JSONOut(http.StatusBadRequest, err.Error(), nil)
				return
			}

			memo, _ := json.Marshal(this.Ctx.Input.Header)
			user.Log(models.ACTION_SIGNUP, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo)

			this.JSONOut(http.StatusOK, "User singup successfully!", nil)
			return
		}
	}
}

func (this *UserWebAPIV1Controller) GetUsers() {
	user := new(models.User)
	users := user.All()

	this.JSONOut(http.StatusOK, "", users)
	return
}

func (this *UserWebAPIV1Controller) GetUser() {
	user := new(models.User)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":username")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false && err == nil {
		this.JSONOut(http.StatusBadRequest, "Search user error", nil)
		return
	}

	user.Password = "xxxxxx"
	this.JSONOut(http.StatusOK, "", user)
	return
}

func (this *UserWebAPIV1Controller) GetNamespaces() {
	if user, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	} else if user.Username != this.Ctx.Input.Param(":username") {
		this.JSONOut(http.StatusBadRequest, "Invalid user session", nil)
		return
	} else {
		namespaces := make([]string, 0)

		namespaces = append(namespaces, user.Username)
		namespaces = append(namespaces, user.Organizations...)

		this.JSONOut(http.StatusOK, "", namespaces)
		return
	}
}

func (this *UserWebAPIV1Controller) PostGravatar() {
	var user models.User

	if u, exist := this.Ctx.Input.CruSession.Get("user").(models.User); exist == false {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	} else if u.Username != this.Ctx.Input.Param(":username") {
		this.JSONOut(http.StatusBadRequest, "Invalid user session", nil)
		return
	}

	file, fileHeader, err := this.Ctx.Request.FormFile("file")
	if err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	prefix := strings.Split(fileHeader.Filename, ".")[0]
	suffix := strings.Split(fileHeader.Filename, ".")[1]
	if suffix != "png" && suffix != "jpg" && suffix != "jpeg" {
		this.JSONOut(http.StatusBadRequest, "gravatar must be .jpg,.jepg or .png", nil)
		return
	}

	if _, err := os.Stat(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("gravatar"), "/", prefix, "_resize.", suffix)); err == nil {
		os.Remove(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("gravatar"), "/", prefix, "_resize.", suffix))
	}

	f, err := os.OpenFile(fmt.Sprintf("%s%s%s", beego.AppConfig.String("gravatar"), "/", fileHeader.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	io.Copy(f, file)
	f.Close()

	// decode jpeg into image.Image
	var img image.Image
	imageFile, err := os.Open(fmt.Sprintf("%s%s%s", beego.AppConfig.String("gravatar"), "/", fileHeader.Filename))
	if err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
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
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	imageFile.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(100, 100, img, resize.Lanczos3)

	out, err := os.Create(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("gravatar"), "/", prefix, "_resize.", suffix))
	if err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
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

	os.Remove(fmt.Sprintf("%s%s%s", beego.AppConfig.String("gravatar"), "/", fileHeader.Filename))

	url := fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("gravatar"), "/", prefix, "_resize.", suffix)

	user.Gravatar = url
	if err := user.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "User save failure", nil)
		return
	}

	this.Ctx.Input.CruSession.Set("user", user)

	this.JSONOut(http.StatusOK, "", map[string]string{"message": "Please click button to finish uploading gravatar", "url": url})
	return
}

func (this *UserWebAPIV1Controller) PutProfile() {
	var p map[string]interface{}
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &p); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist != true {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	} else if user.Username != this.Ctx.Input.Param(":username") {
		this.JSONOut(http.StatusBadRequest, "Invalid user session", nil)
		return
	}

	if strings.Contains(fmt.Sprint(p["gravatar"]), "resize") {
		suffix := strings.Split(fmt.Sprint(p["gravatar"]), ".")[1]
		gravatar := fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("gravatar"), "/", user.Username, "_gravatar.", suffix)
		if _, err := os.Stat(gravatar); err == nil {
			os.Remove(gravatar)
		}

		os.Rename(fmt.Sprint(p["gravatar"]), gravatar)
		p["gravatar"] = gravatar
	}

	user.Email, user.Fullname, user.Mobile = p["email"].(string), p["fullname"].(string), p["mobile"].(string)
	user.Gravatar, user.Company, user.URL = p["gravatar"].(string), p["company"].(string), p["url"].(string)

	if err := user.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "User save failure", nil)
		return
	}

	this.Ctx.Input.CruSession.Set("user", user)

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	user.Log(models.ACTION_UPDATE_PROFILE, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo)

	this.JSONOut(http.StatusOK, "Update Profile Successfully!", nil)
	return
}

func (this *UserWebAPIV1Controller) PutPassword() {
	var p map[string]interface{}
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &p); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, exist := this.Ctx.Input.CruSession.Get("user").(models.User)
	if exist == false {
		this.JSONOut(http.StatusBadRequest, "", map[string]string{"message": "Session load failure", "url": "/auth"})
		return
	} else if user.Username != this.Ctx.Input.Param(":username") {
		this.JSONOut(http.StatusBadRequest, "Invalid user session", nil)
		return
	} else if p["oldPassword"].(string) != user.Password {
		this.JSONOut(http.StatusBadRequest, "account and password not match", nil)
		return
	}

	user.Password = p["newPassword"].(string)
	if err := user.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	user.Log(models.ACTION_UPDATE_PASSWORD, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo)

	this.JSONOut(http.StatusOK, "Update password success!", nil)
	return
}
