package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"encoding/json"
	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/syncMan"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
	"html/template"
	"io/ioutil"
)

// SyncMan is responsible for managing synchronisations with address books etc.
type SyncManController struct {
}

const TAG = dbg.Tag("goodl/syncMan.go")

// GetViewData provides CRUD for synchronisations, as well as triggering an update of synchronized data.
func (SyncManController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	dbg.D(TAG, "I'm at SyncManController")
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), false)
		}
	}()
	viewPath = "views/showDataMessage.htm"
	T := ctx.Value("T").(*Translater)
	action := r.FormValue("action")
	mdl := webfw.Model{}
	vd = webfw.ViewData{
		Model: &mdl,
		T:     T,
		Data:  make(map[string]interface{}),
	}
	usr, _, _ := userManager.GetUserWithSession(r)
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to syncMan without logging in?"), "", nil, true)
	}
	dbCon, err := userManager.GetLocationDb(usr.Id())
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get userdb :("), "", err, true)
	}
	defer dbCon.Close()
	fPath := webfw.Config().RootDir + "/DONTADDTOGIT/client_secret.json"
	clientSecretFilecontent, err := ioutil.ReadFile(fPath)
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Error reading clientsecretfile %s : ", fPath, err), "", err, true)
	}

	var marshaled []byte
	if action == "read" {
		var res syncMan.JSONSyncManAnswer
		res, err = syncMan.JSONGetSyncs(dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get addressManager.JSONGetContactAddressCollection :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	} else if action == "autoRefresh" {
		var res syncMan.JSONRefreshSyncManAnswer
		res, err = syncMan.JSONAutoRefresh(clientSecretFilecontent,usr.Id(), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get syncMan.JSONRefresh :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	} else if action == "refresh" {
		var res3 syncMan.JSONRefreshSyncManAnswer

		res3, err = syncMan.JSONRefreshSync(r.FormValue("sync"), clientSecretFilecontent,usr.Id(), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get syncMan.JSONRefresh :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res3.ErrorMessage, res3.Errors, T)
		marshaled, err = json.Marshal(res3)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res3, err), "", err, true)
		}
	} else if action == "create" {
		var res1 syncMan.JSONTokenSyncManAnswer
		fPath := webfw.Config().RootDir + "/DONTADDTOGIT/client_secret.json"
		clientSecretFilecontent, err := ioutil.ReadFile(fPath)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Error reading clientsecretfile %s : ", fPath, err), "", err, true)
		}
		res1.Success = true
		res1.Id = -1
		if r.FormValue("googleAuthCode") != "" { // this is a google auth - request refresh token
			res1, err = syncMan.JSONCreateRefreshToken(clientSecretFilecontent, r.FormValue("googleAuthCode"), dbCon)
			if err != nil {
				return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get syncMan.JSONCreateRefreshToken :( ", err), "", err, true)
			}
			tools.TranslateErrors(&res1.ErrorMessage, res1.Errors, T)
		}
		if res1.Success {
			res2, err := syncMan.JSONCreateSync(r.FormValue("sync"), res1.Id, dbCon)
			if err != nil {
				dbg.E(TAG, "JSONCreateSync was not succesful : ", err)
				if err != nil {
					return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res1, err), "", err, true)
				}
			}
			marshaled, err = json.Marshal(res2)
		} else {
			dbg.E(TAG, "Creating refresh token was not succesful : %+v", res1)
			marshaled, err = json.Marshal(res1)
			if err != nil {
				return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res1, err), "", err, true)
			}
		}

	} else if action == "update" {
		var res1 syncMan.JSONTokenSyncManAnswer
		fPath := webfw.Config().RootDir + "/DONTADDTOGIT/client_secret.json"
		clientSecretFilecontent, err := ioutil.ReadFile(fPath)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Error reading clientsecretfile %s : ", fPath, err), "", err, true)
		}
		res1.Success = true
		res1.Id = -1
		if r.FormValue("googleAuthCode") != "" { // this is a google auth - request refresh token
			res1, err = syncMan.JSONCreateRefreshToken(clientSecretFilecontent, r.FormValue("googleAuthCode"), dbCon)
			if err != nil {
				return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get syncMan.JSONCreateRefreshToken :( ", err), "", err, true)
			}
			tools.TranslateErrors(&res1.ErrorMessage, res1.Errors, T)
		}
		if res1.Success {
			res2, err := syncMan.JSONUpdateSync(r.FormValue("sync"), res1.Id, dbCon)
			if err != nil {
				dbg.E(TAG, "JSONUpdateSync was not succesful : ", err)
				return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res1, err), "", err, true)
			}
			marshaled, err = json.Marshal(res2)
		} else {
			dbg.E(TAG, "Creating refresh token was not succesful : %+v", res1)
			marshaled, err = json.Marshal(res1)
			if err != nil {
				return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res1, err), "", err, true)
			}
		}

	} else if r.FormValue("action") == "delete" {
		var res models.JSONDeleteAnswer
		res, err = syncMan.JSONDeleteSync(r.FormValue("sync"), clientSecretFilecontent, dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get syncMan.JSONDeleteSync :( ", err), "", err, true)
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
