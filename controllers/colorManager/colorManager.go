package controllers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"encoding/json"

	"github.com/Compufreak345/dbg"
	colorManager "github.com/OpenDriversLog/goodl-lib/jsonapi/colorManager"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
)

const TAG = dbg.Tag("goodl/colorManagerController.go")

// ColorManagerController serves to manage colors (or to be more exact arrays of 3 corresponding colors)
type ColorManagerController struct {
}

// GetViewData CRUDs color-data depending on the "action"-parameter.
// read : Return list of all colors
// create : Creates the given color
// update : Updates the given color
// delete : Deletes the given color
// getColorTemplate : Gets an empty color-object
func (ColorManagerController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	dbg.D(TAG, "ColorManagerController called")
	T := ctx.Value("T").(*Translater)

	vd = webfw.ViewData{
		T:              T,
		NoStyleOnError: true,
	}
	viewPath = "views/showDataMessage.htm"

	usr, _, _ := userManager.GetUserWithSession(r)
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to colorManager without logging in?"), "", nil, true)
	}
	dbCon, err := userManager.GetLocationDb(usr.Id())
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get userdb :("), "", err, true)
	}
	defer dbCon.Close()
	var marshaled []byte

	if r.FormValue("action") == "read" {
		var res colorManager.JSONColorsAnswer
		res, err = colorManager.JSONGetColors(dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get colorManager.JSONGetColors :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	if r.FormValue("action") == "getColorTemplate" {
		res, err := colorManager.JSONGetEmptyColor()

		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get colorManager.JSONGetEmptyColor :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	if r.FormValue("action") == "create" {
		var res models.JSONInsertAnswer
		res, err = colorManager.JSONCreateColor(r.FormValue("color"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get colorManager.JSONCreateColor :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "update" {
		var res models.JSONUpdateAnswer
		res, err = colorManager.JSONUpdateColor(r.FormValue("color"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get colorManager.JSONUpdateAnswer :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "delete" {
		var res models.JSONDeleteAnswer
		res, err = colorManager.JSONDeleteColor(r.FormValue("color"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get colorManager.JSONDeleteColor :( ", err), "", err, true)
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
