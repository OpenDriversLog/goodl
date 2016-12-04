// Package controllers provides many controllers for running the ODL-page.
package controllers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"encoding/json"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/addressManager"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"

	"golang.org/x/net/context"
)

const TAG = dbg.Tag("goodl/addressManagerController.go")

// AddressManagerController provides CRUD-functions to manage contacts & addresses.
type AddressManagerController struct {
}

// GetViewData CRUDs contacts / addresses depending on the "action"-parameter. Addresses can only be created by creating contacts.
// read : Return list of all contacts and/or addresses, depending on the parameters "contacts" or "addresses" to be 1
// getContactTemplate : returns an empty contact
// create : Creates the given contact, creating a GeoZone dependant on the given address
// update : Updates the given contact, updating the GeoZone dependant on the given address
// delete : Deletes the given contact and its GeoZone.
func (AddressManagerController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	dbg.D(TAG, "AddressManagerController called")
	T := ctx.Value("T").(*Translater)

	vd = webfw.ViewData{
		T:              T,
		NoStyleOnError: true,
	}
	viewPath = "views/showDataMessage.htm"

	usr, _, _ := userManager.GetUserWithSession(r)
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to adressManager without logging in?"), "", nil, true)
	}
	dbCon, err := userManager.GetLocationDb(usr.Id())
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get userdb :("), "", err, true)
	}
	defer dbCon.Close()
	var marshaled []byte

	if r.FormValue("action") == "read" {
		var res addressManager.JSONAddressManAnswer
		res, err = addressManager.JSONGetContactAddressCollection(r.FormValue("contacts") == "1", r.FormValue("addresses") == "1", dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get addressManager.JSONGetContactAddressCollection :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	if r.FormValue("action") == "getContactTemplate" {
		res, err := addressManager.JSONGetEmptyContact()

		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get addressManager.JSONGetEmptyContact :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	if r.FormValue("action") == "create" {
		var res addressManager.JSONInsertAddressManAnswer
		res, err = addressManager.JSONCreateContact(r.FormValue("contact"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get addressManager.JSONCreateContact :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "update" {
		var res addressManager.JSONUpdateAddressManAnswer
		res, err = addressManager.JSONUpdateContact(r.FormValue("contact"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get addressManager.JSONUpdateAnswer :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "delete" {
		var res models.JSONDeleteAnswer
		res, err = addressManager.JSONDeleteContact(r.FormValue("contact"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get addressManager.JSONDeleteContact :( ", err), "", err, true)
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
