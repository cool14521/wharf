package handler

import (
	"net/http"

	"github.com/Unknwon/macaron"

	"github.com/astaxie/beego/logs"
)

func W1PostOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1PutOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1GetOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1DeleteOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}
