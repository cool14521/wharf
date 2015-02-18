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

func authToken(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

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
		beego.Error("[REGISTRY API V1] Without write privilege for update the repository's json")
		return auth, http.StatusForbidden, []byte("Forbidden Push Repository")
	}

	return true, 0, nil
}

func AuthPutRepositoryTag(Ctx *context.Context) (bool, int, []byte) {
	if auth, code, message := authToken(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "write" {
		beego.Error("[REGISTRY API V1] Without write privilege for update the repository's tag")
		return false, http.StatusUnauthorized, []byte("Without write privilege for update the repository's tag")
	}

	return true, 0, nil
}

func AuthPutRepositoryImage(Ctx *context.Context) (bool, int, []byte) {
	if auth, code, message := authBasic(Ctx); auth == false {
		return auth, code, message
	}

	if auth, _, code, message, _, _ := authNamespace(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "write" {
		beego.Error("[REGISTRY API V1] Without write privilege for update the repository")
		return false, http.StatusUnauthorized, []byte("Without write privilege for update the repository")
	}

	return true, 0, nil
}

func AuthGetRepositoryImages(Ctx *context.Context) (bool, int, []byte) {
	if auth, code, message := authBasic(Ctx); auth == false {
		return auth, code, message
	}

	if auth, _, code, message, _, _ := authNamespace(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "read" {
		beego.Error("[REGISTRY API V1] Without read privilege for repository json")
		return false, http.StatusUnauthorized, []byte("REGISTRY API V1] Without read privilege for repository json")
	}

	return true, 0, nil
}

func AuthGetRepositoryTags(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	if auth, code, message := authToken(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "read" {
		beego.Error("[REGISTRY API V1] Without read privilege for repository images")
		return false, http.StatusUnauthorized, []byte("Without read privilege for repository images")
	}

	return true, 0, nil
}

func AuthGetImageJSON(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {

	if auth, code, message := authToken(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "write" && Ctx.Input.Session("access") != "read" {
		beego.Error("[REGISTRY API V1] Without read/write privilege in user session")
		return false, http.StatusUnauthorized, []byte("Without read/write privilege in user session")
	}

	imageId := string(Ctx.Input.Param(":image_id"))

	image := new(models.Image)

	if pushed, err := image.IsPushed(imageId); err != nil {
		return false, http.StatusBadRequest, []byte("Search Image Error")
	} else if pushed == false {
		return false, http.StatusBadRequest, []byte("Search Image None")
	}

	return true, 0, nil
}

func AuthPutImageJSON(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {
	if auth, code, message := authToken(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "write" {
		beego.Error("[REGISTRY API V1] Without write image json privilege in user session")
		return false, http.StatusUnauthorized, []byte("Without write image json privilege in user session")
	}

	return true, 0, nil
}

func AuthPutImageLayer(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {
	if auth, code, message := authToken(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "write" {
		beego.Error("[REGISTRY API V1] Without write image privilege in user session")
		return false, http.StatusUnauthorized, []byte("Without write image privilege in user session")
	}

	return true, 0, nil
}

func AuthPutChecksum(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {
	if auth, code, message := authToken(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "write" {
		beego.Error("[REGISTRY API V1] Without write image checksum privilege in user session")
		return false, http.StatusUnauthorized, []byte("Without write image checksum privilege in user session")
	}

	return true, 0, nil
}

func AuthGetImageAncestry(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {
	if auth, code, message := authToken(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "read" {
		beego.Error("[REGISTRY API V1] Without read image ancestry privilege in user session")
		return false, http.StatusUnauthorized, []byte("Without read image ancestry privilege in user session")
	}

	return true, 0, nil
}

func AuthGetImageLayer(Ctx *context.Context) (IsAuth bool, ErrCode int, ErrInfo []byte) {
	if auth, code, message := authToken(Ctx); auth == false {
		return auth, code, message
	}

	if Ctx.Input.Session("access") != "read" {
		beego.Error("[REGISTRY API V1] Without read image layer privilege in user session")
		return false, http.StatusUnauthorized, []byte("Without read image layer privilege in user session")
	}

	return true, 0, nil
}
