package handler

import (
	"net/http"

	"github.com/Unknwon/macaron"

	"github.com/astaxie/beego/logs"
)

func W1PostToken(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}
