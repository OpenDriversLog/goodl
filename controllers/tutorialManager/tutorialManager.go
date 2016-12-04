package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"encoding/json"
	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/tutorialManager"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"html/template"
)

const TAG = dbg.Tag("goodl/tutorialManager.go")

// Main magic happens in userManager/userManager.go
type TutorialManagerController struct {
}

func (TutorialManagerController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	dbg.D(TAG, "TutorialManagerController called")
	T := ctx.Value("T").(*Translater)
	usr, _, _ := userManager.GetUserWithSession(r)

	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to TutorialManagerController without logging in?"), "", nil, true)
	}
	dbCon, err := userManager.GetLocationDb(usr.Id())
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get userdb :("), "", err, true)
	}
	defer dbCon.Close()
	var marshaled []byte

	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}

	viewPath = "views/showDataMessage.htm"

	if r.FormValue("action") == "read" {
		var res tutorialManager.JSONTutorialAnswer
		res, err = tutorialManager.JSONGetTutorialInfo(dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get tutorialManager.JSONGetTutorial :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	} else if r.FormValue("action") == "update" {
		var res models.JSONUpdateAnswer
		res, err = tutorialManager.JSONUpdateTutorial(r.FormValue("tutorial"), usr, r, T, dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get tutorialManager.JSONUpdateAnswer :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	mString := string(marshaled)
	if mString == "" {
		marshaled, err = json.Marshal(models.GetBadJSONAnswer("Unknown action"))
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal models.GetBadJSONAnswer(\"Unknown action\") \n error: %v", err), "", err, true)
		}
		mString = string(marshaled)
	}
	vd.Data = map[string]interface{}{"Message": template.HTML(mString)}

	return
}
