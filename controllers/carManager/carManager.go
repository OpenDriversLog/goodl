package controllers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"encoding/json"

	"github.com/Compufreak345/dbg"
	carManager "github.com/OpenDriversLog/goodl-lib/jsonapi/carManager"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
)

const TAG = dbg.Tag("goodl/carManagerController.go")

// CarManagerController serves for updating car data.
type CarManagerController struct {
}
// GetViewData CRUDs car-data depending on the "action"-parameter.
// read : Return list of all cars
// create : Creates the given car
// update : Updates the given car
// delete : Deletes the given car
// getCarTemplate : Gets an empty car-object
func (CarManagerController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	dbg.D(TAG, "CarManagerController called")
	T := ctx.Value("T").(*Translater)

	vd = webfw.ViewData{
		T:              T,
		NoStyleOnError: true,
	}
	viewPath = "views/showDataMessage.htm"

	usr, _, _ := userManager.GetUserWithSession(r)
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to carManager without logging in?"), "", nil, true)
	}
	dbCon, err := userManager.GetLocationDb(usr.Id())
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get userdb :("), "", err, true)
	}
	defer dbCon.Close()
	var marshaled []byte

	if r.FormValue("action") == "read" {
		var res carManager.JSONCarsAnswer
		res, err = carManager.JSONGetCars(dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get carManager.JSONGetCars :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	if r.FormValue("action") == "getCarTemplate" {
		res, err := carManager.JSONGetEmptyCar()

		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get carManager.JSONGetEmptyCar :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	if r.FormValue("action") == "create" {
		var res models.JSONInsertAnswer
		res, err = carManager.JSONCreateCar(r.FormValue("car"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get carManager.JSONCreateCar :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "update" {
		var res models.JSONUpdateAnswer
		res, err = carManager.JSONUpdateCar(r.FormValue("car"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get carManager.JSONUpdateAnswer :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "delete" {
		var res models.JSONDeleteAnswer
		res, err = carManager.JSONDeleteCar(r.FormValue("car"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get carManager.JSONDeleteCar :( ", err), "", err, true)
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
