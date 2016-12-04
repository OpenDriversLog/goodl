package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/net/context"

	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"

	"github.com/Compufreak345/dbg"

	"github.com/OpenDriversLog/goodl-lib/models"
	betaMan "github.com/OpenDriversLog/goodl/controllers/betaMan"
	tools "github.com/OpenDriversLog/goodl/utils/tools"
)

// SignupBetaController is responsible to allow users to apply for the closed beta-test.
// !!!TO BE DISCUSSED!!!
type SignupBetaController struct {
}

const TAG = dbg.Tag("goodl/signup-beta.go")

// GetViewData allows users to apply for the closed beta-test.
func (SignupBetaController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	dbg.D(TAG, "I'm at SignupBetaController")
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), false)
		}
	}()

	T := ctx.Value("T").(*Translater)
	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}

	viewPath = "views/showDataMessage.htm"

	var res models.JSONInsertAnswer

	// r.ParseForm()
	strBU := r.FormValue("betaUser")
	dbg.D(TAG, "got formvalue beta", strBU)

	if strBU == "" {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "email address required ", err), "", err, true)
	} else {
		dbCon, err := betaMan.GetBetaDbCon()
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Failed to connect to BetaUserDb ", err), "", err, true)
		}

		// res, err = betaMan.JSONInsertBetaUser(m, dbCon)
		res, err = betaMan.JSONInsertBetaUser(strBU, dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Failed to signup user to Beta ", err), "", err, true)
		}
	}

	tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

	marshaled, err := json.Marshal(res)
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
