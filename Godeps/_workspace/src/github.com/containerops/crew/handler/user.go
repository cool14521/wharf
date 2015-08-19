package handler

import (
	"net/http"

	"github.com/Unknwon/macaron"

	"github.com/astaxie/beego/logs"
)

func W1GetUserList(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1PutUserProfile(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1GetUserProfile(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1PostUserGravatar(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1PutUserPasswd(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1GetUserOrganizations(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1GetUserTeams(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}
