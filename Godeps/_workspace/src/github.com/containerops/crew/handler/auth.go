package handler

import (
	"net/http"

	"github.com/Unknwon/macaron"
	"github.com/macaron-contrib/binding"

	"github.com/astaxie/beego/logs"

	"github.com/containerops/wrench/utils"
)

type UserSignup struct {
	Username string `from:"username" Binding:"Required"`
	Email    string `from:"email" Binding:"Required;Email"`
	Passwd   string `from:"passwd" Binding:"Required"`
	Confirm  string `from:"passwd_confirm" Binding:"Required"`
}

func (u *UserSignup) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	if err := utils.ValidatePassword(u.Passwd); err != nil {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"passwd"},
			Classification: "PasswdError",
			Message:        err.Error(),
		})
	}

	if u.Passwd != u.Confirm {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"passwd_confirm"},
			Classification: "PasswdConfirmError",
			Message:        "Confirm password is not equal password",
		})
	}

	return errs
}

func W1UserSignup(ctx *macaron.Context, log *logs.BeeLogger, user UserSignup) (int, []byte) {
	return http.StatusOK, []byte{}
}

func W1UserSignin(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte{}
}
