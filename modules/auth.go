package modules

import (
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
		beego.Error("[REGISTRY API V1] Docker HTTP API Without Authorization In Header")
		beego.Error(Ctx.Request.Header)
		return false, http.StatusUnauthorized, []byte("Docker HTTP API Without Authorization In Header")
	}

	if strings.Index(Ctx.Input.Header("Authorization"), "Basic") == -1 {
		beego.Error("[REGISTRY API V1] Docker HTTP API Without Authorization And Token In Header")
		beego.Error(Ctx.Request.Header)
		return false, http.StatusUnauthorized, []byte("Docker HTTP API Without Authorization And Token In Header")
	}

	if username, passwd, err := utils.DecodeBasicAuth(Ctx.Input.Header("Authorization")); err != nil {
		beego.Error("[REGISTRY API V1] Decode Header Authorization Basic Auth Error: ", err.Error())
		return false, http.StatusUnauthorized, []byte("Decode Header Authorization Basic Auth Error")
	} else {
		user := new(models.User)
		if err = user.Get(username, passwd); err != nil {
			beego.Error("[REGISTRY API V1] User Authorization Error: ", err.Error())

			return false, http.StatusUnauthorized, []byte("User Authorization Error")
		}

		return true, 0, nil
	}
}

func authNamespace(Ctx *context.Context) (Auth bool, NamespaceType bool, Code int, Message []byte, Read bool, Write bool) {
	namespace := string(Ctx.Input.Param(":namespace"))
	repository := string(Ctx.Input.Param(":repository"))

	org := new(models.Organization)
	isOrg, _, err := org.Has(namespace)
	if err != nil {
		beego.Error("[REGISTRY API V1] Search Organization Error: ", err.Error())
		return false, false, http.StatusForbidden, []byte("Search Organization Error"), false, false
	}

	user := new(models.User)
	isUser, _, err := user.Has(namespace)
	if err != nil {
		beego.Error("[REGISTRY API V1] Search User Error: ", err.Error())
		return false, false, http.StatusForbidden, []byte("Search User Error"), false, false
	}

	if !isUser && !isOrg {
		beego.Error("[REGISTRY API V1] Search Namespace Error")
		return false, false, http.StatusForbidden, []byte("Search Namespace Error"), false, false
	}

	Auth, NamespaceType, Code, Message, Read, Write = false, false, 0, nil, false, false

	if isOrg == true {

		for _, value := range user.Organizations {
			organization := new(models.Organization)

			if err := organization.Get(value); err != nil {
				beego.Error("[REGISTRY API V1] Search Organization Error")
				return false, false, http.StatusForbidden, []byte("Search Organization Error"), false, false
			}

			if namespace == organization.Organization {
				NamespaceType = true
			}
		}

		if NamespaceType == false {
			return false, false, http.StatusForbidden, []byte("User not in the organization"), false, false
		}

		for _, value := range user.Teams {
			team := new(models.Team)

			if err := team.Get(value); err != nil {
				beego.Error("[REGISTRY API V1] Search Team Error")
				return false, false, http.StatusForbidden, []byte("Search Team Error"), false, false
			}

			for _, v := range team.TeamPrivileges {
				privilege := new(models.Privilege)
				if err := privilege.Get(v); err != nil {
					return false, false, http.StatusForbidden, []byte("Search Team Privilege Error"), false, false
				}

				if privilege.Repository == repository {
					if privilege.Privilege == true {
						Read = true
						Write = true
					} else if Read == false {
						Read = true
						Write = false

					}
					return true, NamespaceType, Code, Message, Read, Write
				}
			}
		}
	} else {
		if user.Username == namespace {
			Auth = true
			Read = true
			Write = true
		} else {
			Auth, Code, Message = false, http.StatusUnauthorized, []byte("Unauthorized Namespace")
		}
	}
	return Auth, NamespaceType, Code, Message, Read, Write
}

func authToken(Ctx *context.Context) (bool, int, []byte) {
	if strings.Index(Ctx.Input.Header("Authorization"), "Token") == -1 {
		return false, http.StatusUnauthorized, []byte("No Basic Auth And Token In HTTP Header Authorization")
	}

	r, _ := regexp.Compile(`Token (?P<token>\w+)`)
	tokens := r.FindStringSubmatch(Ctx.Input.Header("Authorization"))
	_, token := tokens[0], tokens[1]

	t := Ctx.Input.Session("token")

	if token != t {
		return false, http.StatusUnauthorized, []byte("Unauthorized Token")
	}

	return true, 0, []byte("")
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

	Ctx.Input.CruSession.Set("access", "read")

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

	if pushed, err := image.Pushed(imageId); err != nil {
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
