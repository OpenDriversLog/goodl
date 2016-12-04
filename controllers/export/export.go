package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/export"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
	"html/template"
	"net/http"
	"strconv"
	//"github.com/OpenDriversLog/goodl/utils/dbManHelper"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/goodl/utils/notificationChecker"
)

// ExportController provides an interface for logbook-export
type ExportController struct {
}

const TAG = dbg.Tag("goodl/export")

// GetViewData returns a view for exporting a drivers logbook in different formats (currently only pdf)
// parameters : carId - ID of the car to export, minTime & maxTime - timestamps for the timeframe to export.
// format - currently only pdf
func (ExportController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	dbg.D(TAG, "I'm at ExportController")
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), false)
		}
	}()
	// Message we know that we started
	dbg.D(TAG, "JsonApiController called")

	T := ctx.Value("T").(*Translater)

	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}
	viewPath = "views/showDataMessage.htm"

	// get current logged in user
	usr, _, _ := userManager.GetUserWithSession(r)

	// Just for security reasons (should not happen but does not hurt) :
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to exportController without logging in?"), "", nil, true)
	}

	database, err := userManager.GetLocationDb(usr.Id())
	if err != nil {
		dbg.E(TAG,"Error getting user-database : ", err)
		return
	}
	defer database.Close()

	nots,err := notificationChecker.GetActiveNotificationsForUser(usr.Id(),database,false)
	if err != nil {
		dbg.E(TAG,"Error getting active notifications for user %d : ",usr.Id(), err)
		return
	}
	var marshaled []byte

	sCarId := r.FormValue("carId")
	sMinTime := r.FormValue("minTime")
	sMaxTime := r.FormValue("maxTime")
	if sMaxTime == "" || sMinTime == "" || sCarId == "" {
		return webfw.GetErrorViewData(TAG, 500, dbg.W(TAG, "Wrong export call"), "", errors.New("Unknown request"), true)
	}

	maxTime, err := strconv.ParseInt(r.FormValue("maxTime"), 10, 64)
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.W(TAG, "Could not parse maxTime"), "", errors.New("Unknown request"), true)
	}

	minTime, err := strconv.ParseInt(r.FormValue("minTime"), 10, 64)
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.W(TAG, "Could not parse minTime"), "", errors.New("Unknown request"), true)
	}

	carId, err := strconv.ParseInt(r.FormValue("carId"), 10, 64)
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.W(TAG, "Could not parse carId"), "", errors.New("Unknown request"), true)
	}
	res, err := export.JSONExport(r.FormValue("format"), userManager.GetUserWorkDir(usr.Id()), minTime, maxTime, carId,nots,T, database)
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get export.JSONExportAnswer :( ", err), "", err, true)
	}

	tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

	marshaled, err = json.Marshal(res)
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
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
