package controllers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/net/context"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
)

// InviteController provides methods to invite a user for the beta-test.
type InviteController struct {
}

const TAG = dbg.Tag("goodl/invite.go")

// GetViewData creates or removes beta-invite-keys
// parameter "do" can be either newKey or removeKey, with given "invitekey"-parameter to be removed.
func (InviteController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	dbg.D(TAG, "I'm at InviteController")
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	T := ctx.Value("T").(*Translater)
	do := r.FormValue("do")
	usr, _, _ := userManager.GetUserWithSession(r)
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to invite without logging in?"), "", nil, true)
	}
	viewPath = "views/showDataMessage.htm"
	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}
	if usr.Level() != 1337 {
		vd.WarningMessage = "Sorry, you have no permission to be here."
		return
	}
	if do == "newKey" {
		key, err := userManager.CreateNewInviteKey()
		if err == nil {
			vd.Data["Message"] = template.HTML(webfw.Config().WebUrl + "/de/odl/register/" + key)
		} else {
			vd.ErrorMessage = template.HTML(fmt.Sprintf("Error : ", err))
		}
	} else if do == "removeKey" {
		vd.StatusMessage = fmt.Sprintf("Result (nil is good) : ", userManager.RemoveInviteKey(r.FormValue("invitekey")))
	} else {
		vd.StatusMessage = "Please either do a request with ?do=newKey or ?do=removeKey&invitekey=***"
	}

	return
}
