package filters

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/astaxie/beego/context"

	"github.com/containerops/wharf/models"
	"github.com/containerops/wharf/modules"
	"github.com/containerops/wharf/utils"
)

const (
	PERMISSION_WRITE = iota
	PERMISSION_READ
)

func FilterAuth(ctx *context.Context) {
	var namespace, repository string
	var permission int

	auth := true
	user := new(models.User)

	namespace = strings.Split(string(ctx.Input.Params[":splat"]), "/")[0]
	repository = strings.Split(string(ctx.Input.Params[":splat"]), "/")[1]

	//Get Permission
	permission = getPermission(ctx.Input.Method())

	//Check Authorization In Header
	if len(ctx.Input.Header("Authorization")) == 0 || strings.Index(ctx.Input.Header("Authorization"), "Basic") == -1 {
		auth = false
		goto AUTH
	}

	//Check Username, Password And Get User
	if username, passwd, err := utils.DecodeBasicAuth(ctx.Input.Header("Authorization")); err != nil {
		auth = false
		goto AUTH
	} else {
		if err := user.Get(username, passwd); err != nil {
			auth = false
			goto AUTH
		}
	}

	//Docker Registry V1 Image Don't Check User/Org Permission
	if isImageResource(ctx.Request.URL.String()) == true {
		goto AUTH
	}

	//Username != namespace
	if user.Username != namespace {
		u := new(models.User)
		if has, _, err := u.Has(namespace); err != nil {
			auth = false
			goto AUTH
		} else if has == false {
			//Org Repository Check
			auth = checkOrgRepositoryPermission(user, namespace, repository, permission)
		} else if has == true {
			//Different User and Public/Private Repository
			auth = checkRepositoriesPrivate(namespace, repository)
		}
	}

AUTH:
	if auth == false {
		result := map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeUnauthorized]}}

		data, _ := json.Marshal(result)

		ctx.Output.Context.Output.SetStatus(http.StatusNotFound)
		ctx.Output.Context.Output.Body(data)
		return
	}
}

func getPermission(method string) int {
	write := map[string]string{"POST": "POST", "PUT": "PUT", "DELETE": "DELETE"}
	read := map[string]string{"HEAD": "HEAD", "GET": "GET"}

	if _, ok := write[method]; ok == true {
		return PERMISSION_WRITE
	}

	if _, ok := read[method]; ok == true {
		return PERMISSION_READ
	}

	return PERMISSION_READ
}

func isImageResource(url string) bool {
	r := bytes.NewReader([]byte(url))
	result, _ := regexp.MatchReader("/v1/images/*", r)

	return result
}

func checkRepositoriesPrivate(namespace, repository string) bool {
	repo := new(models.Repository)

	if has, _, err := repo.Has(namespace, repository); err != nil || has == false {
		return false
	} else if has == true {
		if repo.Privated == true {
			return false
		} else {
			return true
		}
	}

	return false
}

func checkOrgRepositoryPermission(user *models.User, namespace, repository string, permission int) bool {
	owner := false

	//Check Org exists
	org := new(models.Organization)
	if has, _, _ := org.Has(namespace); has == false {
		return false
	}

	//Check Owner, don't care Join team
	for _, k := range user.Organizations {
		if org.Id == k {
			owner = true
		}
	}

	//Check Repository
	repo := new(models.Repository)
	if has, _, _ := repo.Has(namespace, repository); has == false {
		if owner == true {
			return true
		} else {
			return false
		}
	}

	if repo.Privated == false && permission == PERMISSION_READ {
		return true
	}

	//Loop Team
	for _, k := range user.JoinTeams {
		team := new(models.Team)

		if err := team.GetById(k); err != nil {
			return false
		}

		//TODO Check Team Perimssion

	}

	return false
}
