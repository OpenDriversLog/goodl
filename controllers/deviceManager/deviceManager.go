package controllers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"encoding/json"

	"github.com/Compufreak345/dbg"
	deviceManager "github.com/OpenDriversLog/goodl-lib/jsonapi/deviceManager"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
	S "github.com/OpenDriversLog/goodl-lib/models/SQLite"
)

const TAG = dbg.Tag("goodl/deviceManagerController.go")

// DeviceManagerController serves to manage devices (e.g. OBD car plugs)
type DeviceManagerController struct {
}
// GetViewData CRUDs device-data depending on the "action"-parameter.
// read : Return list of all devices
// create : Creates the given device
// update : Updates the given device
// delete : Deletes the given device
// getDeviceTemplate : Gets an empty device-object
func (DeviceManagerController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	dbg.D(TAG, "DeviceManagerController called")
	T := ctx.Value("T").(*Translater)

	vd = webfw.ViewData{
		T:              T,
		NoStyleOnError: true,
	}
	viewPath = "views/showDataMessage.htm"

	usr, _, _ := userManager.GetUserWithSession(r)
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to deviceManager without logging in?"), "", nil, true)
	}
	dbCon, err := userManager.GetLocationDb(usr.Id())
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get userdb :("), "", err, true)
	}
	defer dbCon.Close()
	var marshaled []byte

	if r.FormValue("action") == "read" {
		var res deviceManager.JSONDevicesAnswer
		res, err = deviceManager.JSONGetDevices(dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get deviceManager.JSONGetDevices :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	if r.FormValue("action") == "getDeviceTemplate" {
		res, err := deviceManager.JSONGetEmptyDevice()

		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get deviceManager.JSONGetEmptyDevice :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	if r.FormValue("action") == "create" {
		var res models.JSONInsertAnswer
		res, err = deviceManager.JSONCreateDevice(r.FormValue("device"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get deviceManager.JSONCreateDevice :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)
		devs,err := deviceManager.GetDevicesByWhere(dbCon,"_deviceId=?",res.LastKey)
		if err != nil {
			dbg.WTF(TAG,"We just created an device but get an error for finding it in the database?",res.LastKey, err)
		} else if len(devs)==0 {
			dbg.WTF(TAG,"We just created an device but can't find it in the database?", res.LastKey)
		} else {
			dev := devs[0]
			if dev.Guid != "" { // Update Key in User-Database to point to current user.
				userManager.UpdateKey(&userManager.Key {
					UserId:S.NInt64(usr.Id()),
					GUID:string(dev.Guid),
				})
			}
		}
		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "update" {
		var res models.JSONUpdateAnswer
		res, err = deviceManager.JSONUpdateDevice(r.FormValue("device"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get deviceManager.JSONUpdateAnswer :( ", err), "", err, true)
		}
		devs,err := deviceManager.GetDevicesByWhere(dbCon,"_deviceId=?",res.Id)
		if err != nil {
			dbg.WTF(TAG,"We just tried to update an device but get an error for finding it in the database?",res.Id, err)
		} else if len(devs)==0 {
			dbg.WTF(TAG,"We just tried to update an device but can't find it in the database?", res.Id)
		} else {
			dev := devs[0]
			if dev.Guid != "" { // Update Device in User-Database to point to current user.
				userManager.UpdateKey(&userManager.Key {
					UserId:S.NInt64(usr.Id()),
					GUID:string(dev.Guid),
				})
			}
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "delete" {
		var res deviceManager.JSONDeleteDeviceAnswer
		res, err = deviceManager.JSONDeleteDevice(r.FormValue("device"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get deviceManager.JSONDeleteDevice :( ", err), "", err, true)
		}
		if res.Guid != "" { // Remove assignment to the current user for this device.
			userManager.UpdateKey(&userManager.Key {
				UserId:S.NInt64(-1),
				GUID:string(res.Guid),
			})
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
