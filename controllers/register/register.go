package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
)

// RegisterController handles the protection of the actiavtion of newly registered users.
type RegisterController struct {
}

const TAG = dbg.Tag("goodl/register.go")

// GetViewData activates an account if FormValue "mail" and "activateKey" are given and correct.
func (RegisterController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	dbg.D(TAG, "I'm at RegisterController")
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()

	T := ctx.Value("T").(*Translater)

	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}

	if r.FormValue("activateKey") != "" {
		err = userManager.ActivateUser(r.FormValue("mail"), r.FormValue("activateKey"))
		subdir := webfw.Config().SubDir
		if err != nil {
			vd.Redirect = subdir + "/" + ctx.Value("T").(*Translater).UrlLang + "/odl/login" + "?errorMessage=" + url.QueryEscape(T.T("ActivationFailed"))
			viewPath = "views/odl.htm"
			dbg.W(TAG, "Activation failed : ", err)
			err = nil // Don't trigger MVC Internal Server error
			return

		} else {

			vd.Redirect = subdir + "/" + ctx.Value("T").(*Translater).UrlLang + "/odl/login/" + r.FormValue("mail")+"?statusMessage=" + url.QueryEscape(T.T("ActivationSucessful"))
			viewPath = "views/odl.htm"

			return
		}
	}

	return
}
