package models

import (
	//	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/dockercn/wharf/utils"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	AuthPutRepository       = "DoAuthPutRepository"
	AuthPutRepositoryTag    = "DoAuthPutRepositoryTag"
	AuthPutRepositoryImage  = "DoAuthPutRepositoryImage"
	AuthGetRepositoryImages = "DoAuthGetRepositoryImages"
	AuthGetRepositoryTags   = "DoAuthGetRepositoryTags"

	//-----------------------------------------------------------
	AuthGetImageJSON     = "DoAuthGetImageJSON"
	AuthPutImageJSON     = "DoAuthPutImageJSON"
	AuthGetImageAncestry = "DoAuthGetImageAncestry"
	AuthGetImageLayer    = "DoAuthGetImageLayer"
	AuthPutImageLayer    = "DoAuthPutImageLayer"
	AuthPutChecksum      = "DoAuthPutChecksum"
)

type AuthModel struct {
}

func DoAuth(Ctx *context.Context, callFunc string) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	authModel := new(AuthModel)

	params := make([]reflect.Value, 1)
	params[0] = reflect.ValueOf(Ctx)
	result := reflect.ValueOf(authModel).MethodByName(callFunc).Call(params)

	IsAuth = result[0].Bool()

	strErrCode := strconv.FormatInt(result[1].Int(), 10)
	ErrCode, _ = strconv.Atoi(strErrCode)

	ErrInfo = result[2].Bytes()

	return
}

func (this *AuthModel) DoAuthGetRepositoryImages(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetRepositoryImages")

	IsAuth, ErrCode, ErrInfo = this.DoAuthBasic(Ctx)

	if !IsAuth {
		return
	}

	IsAuth, ErrCode, ErrInfo = this.DoAuthNamespace(Ctx)

	if !IsAuth {
		return
	}

	ErrCode = 0
	ErrInfo = nil
	IsAuth = true
	return

}

func (this *AuthModel) DoAuthPutRepository(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutRepository")

	IsAuth, ErrCode, ErrInfo = this.DoAuthBasic(Ctx)

	if !IsAuth {
		return
	}

	IsAuth, ErrCode, ErrInfo = this.DoAuthNamespace(Ctx)

	if !IsAuth {
		return
	}

	ErrCode = 0
	ErrInfo = nil
	IsAuth = true
	return

}

func (this *AuthModel) DoAuthGetRepositoryTags(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetRepositoryTags")

	IsAuth, ErrCode, ErrInfo = this.DoAuthToken(Ctx)

	if !IsAuth {
		return
	}

	if Ctx.Input.Session("access") != "read" {
		beego.Error("[API 用户] 读取 Repository Tag 时在 Session 中没有 read 的权限记录")
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"error\":\"没有更新 Repository Tag 数据的写权限\"}")
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil
	return

}

func (this *AuthModel) DoAuthPutRepositoryTag(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutRepositoryTag")

	IsAuth, ErrCode, ErrInfo = this.DoAuthToken(Ctx)

	if !IsAuth {
		return
	}

	if Ctx.Input.Session("access") != "write" {

		beego.Error("[API 用户] 更新 Repository 的 Tag 信息时在 Session 中没有 write 的权限记录")
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"error\":\"没有更新 Repository Tag 数据的写权限\"}")
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil
	return

}

func (this *AuthModel) DoAuthPutRepositoryImage(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutRepositoryImage")

	IsAuth, ErrCode, ErrInfo = this.DoAuthBasic(Ctx)

	if !IsAuth {
		return
	}

	IsAuth, ErrCode, ErrInfo = this.DoAuthNamespace(Ctx)

	if !IsAuth {
		return
	}

	if Ctx.Input.Session("access") != "write" {

		beego.Error("[API 用户] 更新 Repository 的 Tag 信息时在 Session 中没有 write 的权限记录")
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"error\":\"没有更新 Repository Tag 数据的写权限\"}")
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil
	return

}

func (this *AuthModel) DoAuthBasic(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	if len(Ctx.Input.Header("Authorization")) == 0 {
		//不存在 Authorization 信息返回错误信息
		beego.Error("[API 用户] Docker 命令访问 HTTP API 的 Header 中没有 Authorization 信息: ")
		beego.Error(Ctx.Request.Header)

		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"错误\":\"没有找到 Authorization 的认证信息\"}")
		IsAuth = false
		return
	}
	isBasicIndex := strings.Index(Ctx.Input.Header("Authorization"), "Basic")

	if isBasicIndex == -1 {
		beego.Error("[API 用户] Docker 命令访问 HTTP API 的 Header 中没有 Basic Auth 和 Token 的信息 ")
		beego.Error(Ctx.Request.Header)

		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"错误\":\"没有找到 Authorization 的认证信息\"}")
		IsAuth = false
		return

	}
	//解码 Basic Auth 进行用户的判断
	username, passwd, err := utils.DecodeBasicAuth(Ctx.Input.Header("Authorization"))

	if err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 解码 Header Authorization 的 Basic Auth [%s] 时遇到错误： %s", Ctx.Input.Header("Authorization"), err.Error()))
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte([]byte("{\"错误\":\"解码 Authorization 的 Basic Auth 信息错误\"}"))
		IsAuth = false
		return

	}

	//根据解码的数据，在数据库中查询用户
	user := new(User)
	has, err := user.Get(username, passwd)
	if err != nil {
		//查询用户数据失败，返回 401 错误
		beego.Error(fmt.Sprintf("[API 用户] 在数据库中查询用户数据遇到错误：%s", err.Error()))
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"错误\":\"在数据库中查询用户数据时出现数据库错误\"}")
		IsAuth = false
		return
	}

	if has == false {
		//查询用户数据失败，返回 401 错误
		beego.Error(fmt.Sprintf("[API 用户] 在数据库中查询用户数据没有发现用户：%s", username))
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"错误\":\"在数据库中查询用户数据时没有发现用户\"}")
		IsAuth = false
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil
	return

}

func (this *AuthModel) DoAuthNamespace(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	//根据 Namespace 查询组织数据
	namespace := string(Ctx.Input.Param(":namespace"))
	org := new(Organization)
	if _, _, err := org.Has(namespace); err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 查询组织名称 %s 时错误 %s", namespace, err.Error()))
		ErrCode = http.StatusForbidden
		ErrInfo = []byte("{\"错误\":\"查询组织数据报错。\"}")
		IsAuth = false
		return
	}

	ErrCode = 0
	ErrInfo = nil
	IsAuth = true
	return

}

func (this *AuthModel) DoAuthToken(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	//非 Basic Auth ，检查 Token
	if strings.Index(Ctx.Input.Header("Authorization"), "Token") == -1 {
		beego.Error("[API 用户] Docker 命令访问 HTTP API 的 Header 中没有 Basic Auth 和 Token 的信息 ")
		beego.Error(Ctx.Request.Header)
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"错误\":\"在 HTTP Header Authorization 中没有找到 Basic Auth 和 Token 信息\"}")
		return
	}

	//使用正则获取 Token 的值
	r, _ := regexp.Compile(`Token (?P<token>\w+)`)
	tokens := r.FindStringSubmatch(Ctx.Input.Header("Authorization"))
	_, token := tokens[0], tokens[1]

	beego.Debug("[Token in Header] " + token)

	t := Ctx.Input.Session("token")

	//用 Header 中的 Token 和 Session 中得 Token 值进行比较，不相等返回错误退出执行
	if token != t {
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"错误\":\"HTTP Header 中的 Token 和 Session 的 Token 不同\"}")
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil
	return
}

func AuthPrepare(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {
	if len(Ctx.Input.Header("Authorization")) == 0 {
		//不存在 Authorization 信息返回错误信息
		beego.Error("[API 用户] Docker 命令访问 HTTP API 的 Header 中没有 Authorization 信息: ")
		beego.Error(Ctx.Request.Header)

		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"错误\":\"没有找到 Authorization 的认证信息\"}")
		IsAuth = false
		return
	}

	//检查是否 Basic Auth
	isBasicIndex := strings.Index(Ctx.Input.Header("Authorization"), "Basic")
	isTokenIndex := strings.Index(Ctx.Input.Header("Authorization"), "Token")
	if isBasicIndex == -1 && isTokenIndex == -1 {
		beego.Error("[API 用户] Docker 命令访问 HTTP API 的 Header 中没有 Basic Auth 和 Token 的信息 ")
		beego.Error(Ctx.Request.Header)

		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"错误\":\"没有找到 Authorization 的认证信息\"}")
		IsAuth = false
		return

	} else if isBasicIndex > -1 {
		//解码 Basic Auth 进行用户的判断
		username, passwd, err := utils.DecodeBasicAuth(Ctx.Input.Header("Authorization"))

		if err != nil {
			beego.Error(fmt.Sprintf("[API 用户] 解码 Header Authorization 的 Basic Auth [%s] 时遇到错误： %s", Ctx.Input.Header("Authorization"), err.Error()))
			ErrCode = http.StatusUnauthorized
			ErrInfo = []byte([]byte("{\"错误\":\"解码 Authorization 的 Basic Auth 信息错误\"}"))
			IsAuth = false
			return

		}

		//根据解码的数据，在数据库中查询用户
		user := new(User)
		has, err := user.Get(username, passwd)
		if err != nil {
			//查询用户数据失败，返回 401 错误
			beego.Error(fmt.Sprintf("[API 用户] 在数据库中查询用户数据遇到错误：%s", err.Error()))
			ErrCode = http.StatusUnauthorized
			ErrInfo = []byte("{\"错误\":\"在数据库中查询用户数据时出现数据库错误\"}")
			IsAuth = false
			return
		}

		if has == true {

			//根据 Namespace 查询组织数据
			namespace := string(Ctx.Input.Param(":namespace"))
			org := new(Organization)
			if _, _, err := org.Has(namespace); err != nil {
				beego.Error(fmt.Sprintf("[API 用户] 查询组织名称 %s 时错误 %s", namespace, err.Error()))
				ErrCode = http.StatusForbidden
				ErrInfo = []byte("{\"错误\":\"查询组织数据报错。\"}")
				IsAuth = false
				return
			}

		} else {
			//没有查询到用户数据，返回 401 错误
			beego.Error(fmt.Sprintf("[API 用户] 没有查询到用户：%s ", username))
			ErrCode = http.StatusUnauthorized
			ErrInfo = []byte("{\"错误\":\"没有查询到用户\"}")
			IsAuth = false
			return
		}

	}
	ErrCode = 0
	ErrInfo = nil
	IsAuth = true
	return

}

//------------------------------------------------------------

func (this *AuthModel) DoAuthGetImageJSON(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetImageJSON")

	IsAuth, ErrCode, ErrInfo = this.DoAuthToken(Ctx)

	if !IsAuth {
		return
	}

	if Ctx.Input.Session("access") != "write" && Ctx.Input.Session("access") != "read" {

		beego.Error("[API 用户] 查询 Image 时在 Session 中没有 write 或 read 的权限记录")
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"error\":\"没有访问 Image 数据的权限\"}")
		return
	}

	//初始化加密签名
	sign := ""
	if len(string(Ctx.Input.Header("X-Docker-Sign"))) > 0 {
		sign = string(Ctx.Input.Header("X-Docker-Sign"))
	}

	//TODO 检查 imageID 的合法性
	imageId := string(Ctx.Input.Param(":image_id"))

	image := new(Image)
	has, err := image.GetPushed(imageId, sign, true, true)

	if err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 查询 Image %s 时报错 ", imageId, err.Error()))
		IsAuth = false
		ErrCode = http.StatusBadRequest
		ErrInfo = []byte("{\"错误\":\"搜索 Image 错误\"}")
		return
	}

	if has != true {
		beego.Error(fmt.Sprintf("[API 用户] 没有查询到 Image ：%s ", imageId))
		IsAuth = false
		ErrCode = http.StatusNotFound
		ErrInfo = []byte("{\"error\":\"没有找到 Image 数据\"}")
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil
	return
}

func (this *AuthModel) DoAuthPutImageJSON(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutImageJSON")

	IsAuth, ErrCode, ErrInfo = this.DoAuthToken(Ctx)

	if !IsAuth {
		return
	}

	if Ctx.Input.Session("access") != "write" {

		beego.Error("[API 用户] 查询 Image 时在 Session 中没有 write 的权限记录")
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"error\":\"没有访问 Image 数据的权限\"}")
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil

	return
}

func (this *AuthModel) DoAuthGetImageAncestry(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetImageAncestry")

	IsAuth, ErrCode, ErrInfo = this.DoAuthToken(Ctx)

	if !IsAuth {
		return
	}

	if Ctx.Input.Session("access") != "read" {
		beego.Error("[API 用户] 读取 Image Ancestry 时在 Session 中没有 read 的权限记录")
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"error\":\"没有读取 Image Ancestry 的权限\"}")
		return
	}

	ErrCode = 0
	ErrInfo = nil
	IsAuth = true
	return

}

func (this *AuthModel) DoAuthGetImageLayer(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetImageLayer")

	IsAuth, ErrCode, ErrInfo = this.DoAuthToken(Ctx)

	if !IsAuth {
		return
	}

	if Ctx.Input.Session("access") != "read" {
		beego.Error("[API 用户] 读取 Image Layer 文件时在 Session 中没有 read 的权限记录")
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"error\":\"没有读取 Image Layer 文件的权限\"}")
		return
	}

	ErrCode = 0
	ErrInfo = nil
	IsAuth = true
	return

}

func (this *AuthModel) DoAuthPutImageLayer(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutImageLayer")

	IsAuth, ErrCode, ErrInfo = this.DoAuthToken(Ctx)

	if !IsAuth {
		return
	}

	if Ctx.Input.Session("access") != "write" {

		beego.Error("[API 用户] 写入 Image Layer 时在 Session 中没有 write 的权限记录")
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"error\":\"没有写入 Image Layer 文件的权限\"}")
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil

	return
}

func (this *AuthModel) DoAuthPutChecksum(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutChecksum")

	IsAuth, ErrCode, ErrInfo = this.DoAuthToken(Ctx)

	if !IsAuth {
		return
	}

	if Ctx.Input.Session("access") != "write" {

		beego.Error("[API 用户] 写入 Image Checksum 时在 Session 中没有 write 的权限记录")
		IsAuth = false
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte("{\"error\":\"没有写入 Image Checksum 的权限\"}")
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil

	return
}
