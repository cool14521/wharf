package handler

import (
	"net/http"

	"github.com/Unknwon/macaron"

	"github.com/astaxie/beego/logs"
)

func W1PostTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1GetTeams(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1PutTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1GetTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1DeleteTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1GetTeamUsers(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1PutTeamUser(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1DeleteTeamUser(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}
