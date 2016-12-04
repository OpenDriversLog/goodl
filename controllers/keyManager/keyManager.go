package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"strconv"
	"html/template"
)

const TAG = dbg.Tag("goodl/keyManagerController.go")
// KeyManController is responsible for generating, deleting & reading device-keys.
type KeyManController struct {
}
// GetViewData provides or manages keys depending on the "action"-parameter.
// read : Return list of all keys (without secret key, as it is not stored)
// generate : Creates a new device-key
// delete : Deletes the given device-key
func (KeyManController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() { // Error handling, if this getviewdata panics (should not happen)
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	dbg.D(TAG, "KeyManagerController called")

	T := ctx.Value("T").(*Translater)

	user, _, _ := userManager.GetUserWithSession(r)
	if user.Level() != 1337 {

		vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,403,dbg.WTF(TAG, "Somebody attempted to access KeyManController without having permissions", err, user),"Insufficient permissions",nil,true)
		return
	}

	vd = webfw.ViewData{
		T:              T,
		NoStyleOnError: true,
	}
	viewPath = "views/showDataMessage.htm"
	var marshaled []byte

	if r.FormValue("action") == "read" {
		var res userManager.JSONKeysAnswer
		res, err = userManager.JSONGetKeys()
		if err != nil {
			vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,500,dbg.E(TAG, "Could not get keyManager.JSONGetKeys :( ", err),"internal server error",nil,true)
			return
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,500,dbg.E(TAG, "unable to marshal result: %v \n error: %v", err),"internal server error",nil,true)
			return
		}
	}

	if r.FormValue("action") == "generate" {

		var res userManager.JSONKeysAnswer
		if r.FormValue("cnt")=="" {

			vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,500,dbg.E(TAG, "Tried to create keys, but no cnt was given"),"Please provide cnt for the number of keys to generate",nil,true)
			return
		}
		var cnt int64
		cnt, err = strconv.ParseInt(r.FormValue("cnt"),10,64)
		if err != nil {
			if r.FormValue("cnt")=="" {
				vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,500,dbg.E(TAG, "Tried to create keys, but cnt was not perseable", err),"internal server error",nil,true)
				return
			}
		}
		res, err = userManager.JSONGenerateKeys(cnt)
		if err != nil {
			vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,500,dbg.E(TAG, "Could not get keyManager.JSONCreateKey :(", err),"internal server error",nil,true)
			return
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,500,dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err),"internal server error",nil,true)
			return
		}
	}
	if r.FormValue("action") == "delete" {
		var res userManager.JSONKeyDeleteAnswer
		res, err = userManager.JSONDeleteKey(r.FormValue("key"))
		if err != nil {
			vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,500,dbg.E(TAG, "Error at keyManager.JSONDeleteKey :(", res, err),"internal server error",nil,true)
			return
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,500,dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err),"internal server error",nil,true)
			return
		}
	}
	mString := string(marshaled)
	if mString == "" {
		marshaled, err = json.Marshal(models.GetBadJSONAnswer("Unknown action"))
		if err != nil {
			vd, viewPath, vShared, err = webfw.GetErrorViewData(TAG,500,dbg.E(TAG, "unable to marshal result: %v \n error: %v", marshaled, err),"internal server error",nil,true)
			return
		}
		mString = string(marshaled)
	}
	vd.Data = map[string]interface{}{"Message": template.HTML(mString)}

	return
}
