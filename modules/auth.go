package modules

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/utils"
)

func authBasic(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

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
	user := new(models.User)
	err = user.Get(username, passwd)
	if err != nil {
		//查询用户数据失败，返回 401 错误
		beego.Error(fmt.Sprintf("[API 用户] 验证用户错误：%s", err.Error()))
		ErrCode = http.StatusUnauthorized
		ErrInfo = []byte(`{"错误":"验证用户错误"}`)
		IsAuth = false
		return
	}

	IsAuth = true
	ErrCode = 0
	ErrInfo = nil
	return

}

func authNamespace(Ctx *context.Context) (IsAuth bool, NamespaceType bool, ErrCode int, ErrInfo []byte, Read bool, Write bool) {

	//根据 Namespace 查询数据
	namespace := string(Ctx.Input.Param(":namespace"))
	repository := string(Ctx.Input.Param(":repository"))
	//	NamespaceType = false //默认用户为普通户用，不是组织

	//判断是用户还是组织

	org := new(models.Organization)
	orgIsHas, _, err := org.Has(namespace)
	if err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 查询组织 %s 时错误 %s", namespace, err.Error()))
		ErrCode = http.StatusForbidden
		ErrInfo = []byte(`{"错误":"查询组织报错。"}`)
		IsAuth = false
		return
	}

	user := new(models.User)
	userIsHas, _, err := user.Has(namespace)
	if err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 查询用户 %s 时错误 %s", namespace, err.Error()))
		ErrCode = http.StatusForbidden
		ErrInfo = []byte(`{"错误":"查询用户报错。"}`)
		IsAuth = false
		return
	}

	if !orgIsHas && !userIsHas {
		beego.Error(fmt.Sprintf("[API 用户] 没有查询到namespace: %s", namespace))
		ErrCode = http.StatusForbidden
		ErrInfo = []byte(`{"错误":"没有查询到namespace。"}`)
		IsAuth = false
		return
	}

	ErrCode = 0
	ErrInfo = nil
	IsAuth = true
	Read = false
	Write = false
	NamespaceType = false //默认用户为普通户用，不是组织

	if orgIsHas {
		//NamespaceType = true
		//判断用户所在NameSpace指定的组织
		//读取用户所有的组织
		for index, value := range user.Organizations {
			fmt.Printf("遍历第[%d]个切片\n", index)
			//查找
			organization := new(models.Organization)
			err = organization.Get(value)
			if err != nil {
				beego.Error(fmt.Sprintf("[API 用户] 验证用户所在组织namespace错误: %s", namespace))
				ErrCode = http.StatusForbidden
				ErrInfo = []byte(`{"错误":"验证用户所在组织namespace错误"}`)
				IsAuth = false
				return
			}
			if namespace == organization.Organization {
				//用户在namespace 指定的组织
				NamespaceType = true
			}
		}

		//用户不在指定的组织
		if !NamespaceType {
			beego.Error(fmt.Sprintf("[API 用户] 用户不在指定的组织: %s", namespace))
			ErrCode = http.StatusForbidden
			ErrInfo = []byte(`{"错误":"用户不在指定的组织"}`)
			IsAuth = false
			return
		}

		//判断用户所有Team 对于此仓库权限的集合
		for index, value := range user.Teams {
			fmt.Printf("遍历第[%d]个Team\n", index)
			team := new(models.Team)
			err = team.Get(value)
			if err != nil {
				beego.Error(fmt.Sprintf("[API 用户] 验证用户所在Team错误"))
				ErrCode = http.StatusForbidden
				ErrInfo = []byte(`{"错误":"验证用户所在Team错误"}`)
				IsAuth = false
				return
			}
			for Pindex, Pvalue := range team.TeamPrivileges {
				fmt.Printf("遍历第[%d]个TeamPrivileges\n", Pindex)
				privilege := new(models.Privilege)
				err = privilege.Get(Pvalue)

				if err != nil {
					beego.Error(fmt.Sprintf("[API 用户] 验证用户所在Team对应的仓库权限错误"))
					ErrCode = http.StatusForbidden
					ErrInfo = []byte(`{"错误":"验证用户所在Team对应的仓库权限错误"}`)
					IsAuth = false
					return
				}
				if privilege.Repository == repository {
					if privilege.Privilege {
						Read = true
						Write = true
					} else if Read == false {
						Read = true
						Write = false
					}

				}

			}

		}

	} else {
		ErrCode = 0
		ErrInfo = nil
		IsAuth = true
		Read = true
		Write = true
		NamespaceType = false //默认用户为普通户用，不是组织
	}

	return

}

func DoAuthToken(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

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

func AuthPutRepository(Ctx *context.Context) (bool, int, []byte) {

	if auth, code, message := authBasic(Ctx); auth == false {

		return auth, code, message

	}

	if auth, _, code, message, _, write := authNamespace(Ctx); auth == false {

		return auth, code, message

	} else if write == false {

		code = http.StatusForbidden
		message = []byte("Forbidden Push Repository")

		return auth, code, message

	} else {

		return true, 0, nil
	}
}

func DoAuthGetImageJSON(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetImageJSON")

	IsAuth, ErrCode, ErrInfo = DoAuthToken(Ctx)

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
	beego.Info("[API 用户] sign:::%s", sign)

	//TODO 检查 imageID 的合法性
	imageId := string(Ctx.Input.Param(":image_id"))

	image := new(models.Image)
	isPushed, err := image.IsPushed(imageId)

	if err != nil {
		beego.Error(fmt.Sprintf("[API 用户] 查询 Image %s 时报错 ", imageId, err.Error()))
		IsAuth = false
		ErrCode = http.StatusBadRequest
		ErrInfo = []byte("{\"错误\":\"搜索 Image 错误\"}")
		return
	}

	if !isPushed {
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

func DoAuthPutImageJSON(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutImageJSON")

	IsAuth, ErrCode, ErrInfo = DoAuthToken(Ctx)

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

func DoAuthPutImageLayer(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutImageLayer")

	IsAuth, ErrCode, ErrInfo = DoAuthToken(Ctx)

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

func DoAuthPutChecksum(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutChecksum")

	IsAuth, ErrCode, ErrInfo = DoAuthToken(Ctx)

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

func DoAuthPutRepositoryTag(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutRepositoryTag")

	IsAuth, ErrCode, ErrInfo = DoAuthToken(Ctx)

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

func DoAuthPutRepositoryImage(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthPutRepositoryImage")

	IsAuth, ErrCode, ErrInfo = authBasic(Ctx)

	if !IsAuth {
		return
	}

	IsAuth, _, ErrCode, ErrInfo, _, _ = authNamespace(Ctx)

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

func DoAuthGetRepositoryImages(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetRepositoryImages")

	IsAuth, ErrCode, ErrInfo = authBasic(Ctx)

	if !IsAuth {
		return
	}
	IsAuth, _, ErrCode, ErrInfo, _, _ = authNamespace(Ctx)

	if !IsAuth {
		return
	}

	ErrCode = 0
	ErrInfo = nil
	IsAuth = true
	return

}

func DoAuthGetRepositoryTags(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetRepositoryTags")

	IsAuth, ErrCode, ErrInfo = DoAuthToken(Ctx)

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

func DoAuthGetImageAncestry(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetImageAncestry")

	IsAuth, ErrCode, ErrInfo = DoAuthToken(Ctx)

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

func DoAuthGetImageLayer(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	beego.Error("执行DoAuthGetImageLayer")

	IsAuth, ErrCode, ErrInfo = DoAuthToken(Ctx)

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
