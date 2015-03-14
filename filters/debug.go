package filters

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func FilterDebug(ctx *context.Context) {
	beego.Debug("[URL]")
	beego.Debug(ctx.Request.URL)
	beego.Debug("[Method]")
	beego.Debug(ctx.Input.Method())
	beego.Debug("[Headers]")
	beego.Debug(ctx.Input.Request.Header)
}
