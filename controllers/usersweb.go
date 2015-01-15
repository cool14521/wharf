package controllers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dockercn/docker-bucket/models"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type UsersWebController struct {
	beego.Controller
}

func (this *UsersWebController) Prepare() {
	beego.Debug(fmt.Sprintf("[%s] %s | %s", this.Ctx.Input.Host(), this.Ctx.Input.Request.Method, this.Ctx.Input.Request.RequestURI))
	beego.Debug("[Header] ")
	beego.Debug(this.Ctx.Request.Header)
}

func (this *UsersWebController) PostGravatar() {
	//从请求中读取图片信息，图片保存在相应
	file, fileHeader, err := this.Ctx.Request.FormFile("file")
	if err != nil {
		beego.Error(fmt.Sprintf("[image] 处理上传头像错误,err=%s", err))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"图片上传处理失败\"}"))
		return
	}

	//读取文件后缀名，如果不是图片，则返回错误
	prefix := strings.Split(fileHeader.Filename, ".")[0]
	suffix := strings.Split(fileHeader.Filename, ".")[1]
	if suffix != "png" && suffix != "jpg" && suffix != "jpeg" {
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"文件的扩展名必须是jpg、jpeg或者png!\"}"))
		return
	}
	//删除名称重复的文件
	if _, err := os.Stat(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix)); err == nil {
		os.Remove(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix))
	}
	f, err := os.OpenFile(fmt.Sprintf("%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", fileHeader.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		//处理文件错误
		beego.Error(fmt.Sprintf("[image] 处理上传头像错误,err=%s", err))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"图片上传处理失败\"}"))
		return
	}
	io.Copy(f, file)
	f.Close()

	// decode jpeg into image.Image
	var img image.Image
	imageFile, err := os.Open(fmt.Sprintf("%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", fileHeader.Filename))
	if err != nil {
		beego.Error(fmt.Sprintf("[image] 上传图片预失败,err=%s", err))
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
		beego.Error(fmt.Sprintf("[image] 裁剪图片失败,err=%s", err))
		imageFile.Close()
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"图片上传处理失败\"}"))
		return
	}
	imageFile.Close()
	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(100, 100, img, resize.Lanczos3)

	out, err := os.Create(fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix))
	if err != nil {
		beego.Error(fmt.Sprintf("[image] 裁剪图片失败,err=%s", err))
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

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"请点击Update profile完成图片上传！\",\"url\":\"" + fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", prefix, "_resize.", suffix) + "\"}"))
	return
}

func (this *UsersWebController) GetProfile() {
	//加载session
	user, ok := this.GetSession("user").(models.User)
	if !ok {
		beego.Error(fmt.Sprintf("[WEB 用户] session加载失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"session加载失败\",\"url\":\"/auth\"}"))
		return
	}
	user2json, err := json.Marshal(user)
	if err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] session解码json失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"session解码json失败\",\"url\":\"/auth\"}"))
		return
	}
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body(user2json)
	return
}

func (this *UsersWebController) PutProfile() {
	//获得用户提交的信息
	var u map[string]interface{}
	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &u); err != nil {
		beego.Error(fmt.Sprintf("[WEB 用户] 解码用户注册发送的 JSON 数据失败: %s", err.Error()))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"更新用户数据失败\"}"))
		return
	}
	//加载session
	user, ok := this.GetSession("user").(models.User)
	if !ok {
		beego.Error(fmt.Sprintf("[WEB 用户] session加载失败"))
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"更新用户数据失败\"}"))
		return
	}
	//处理用户上传头像（判断头像是否更新，如果更新，删掉以前头像，然后重新命名新头像）
	if strings.Contains(fmt.Sprint(u["gravatar"]), "resize") {
		//包含resize，则认为用户上传新头像
		suffix := strings.Split(fmt.Sprint(u["gravatar"]), ".")[1]
		gravatar := fmt.Sprintf("%s%s%s%s%s", beego.AppConfig.String("docker::Gravatar"), "/", user.Username, "_show.", suffix)
		if _, err := os.Stat(gravatar); err == nil {
			//文件存在，计算old_image文件的MD5的值
			bytes, _ := ioutil.ReadFile(gravatar)
			old2MD5 := md5.Sum(bytes)

			//计算new_image文件的MD5的值
			bytes, _ = ioutil.ReadFile(fmt.Sprint(u["gravatar"]))
			newMD5 := md5.Sum(bytes)

			//两个文件的MD5值相同，删除resize文件，返回错误，用户上传重复头像
			if old2MD5 == newMD5 {
				os.Remove(fmt.Sprint(u["gravatar"]))
				u["gravatar"] = gravatar
				this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
				this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
				this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"上传头像与原头像相同！\"}"))
				return
			}
			//删除掉用户之前的头像文件
			os.Remove(gravatar)
			//将新文件重新命名
			os.Rename(fmt.Sprint(u["gravatar"]), gravatar)
			u["gravatar"] = gravatar
		} else {
			os.Rename(fmt.Sprint(u["gravatar"]), gravatar)
			u["gravatar"] = gravatar
		}
	}
	//根据session更新user
	if success, err := (&user).Update(u); err != nil || !success {
		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
		this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"更新用户数据失败\"}"))
		return
	}
	//更新session中的user
	this.SetSession("user", user)
	//处理返回值
	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	this.Ctx.Output.Context.Output.Body([]byte("{\"message\":\"更新用户数据成功\"}"))
	return
}
