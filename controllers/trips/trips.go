package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/deviceManager"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/tripMan"
	tripManModels "github.com/OpenDriversLog/goodl-lib/jsonapi/tripMan/models"

	"github.com/OpenDriversLog/goodl-lib/models"

	"github.com/OpenDriversLog/goodl/utils/tools"

	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/goodl/utils/notificationChecker"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/notificationManager"
)

// Tag for logmessages identifying the file
const TAG = dbg.Tag("goodl/trips.go")

type TripsController struct {
}

// func GetViewData returns a view either displaying a html-view or ajax-data
func (TripsController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() { // Error handling, if this getviewdata panics (should not happen)
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	// Message we know that we started
	dbg.D(TAG, "TripsController called")

	T := ctx.Value("T").(*Translater)

	// If we want to use our own model use the previous
	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}

	var marshaled []byte

	// get current logged in user
	usr, _, _ := userManager.GetUserWithSession(r)
	// Just for security reasons (should not happen but does not hurt) :
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to trips without logging in?"), "", nil, true)
	}

	dbCon, err := userManager.GetLocationDb(usr.Id())
	defer dbCon.Close()

	ajaxRequest := r.FormValue("action")
	if ajaxRequest != "" {

		switch ajaxRequest {
		case "read":
			switch r.FormValue("t") {
			case "trip":
				{
					id, _ := strconv.Atoi(r.FormValue("id"))
					includeTracks := r.FormValue("includeTracks") != ""
					var nots *[]*notificationManager.Notification
					nots,err = notificationChecker.GetActiveNotificationsForUser(usr.Id(),dbCon,false)
					if err != nil {
						dbg.E(TAG,"Error getting active notifications for user %d : ",usr.Id(), err)
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get GetActiveNotificationsForUser :( ", err), "", err, true)
					}
					res, err := tripMan.JSONSelectTrip(int64(id), includeTracks, true,nots,T, r.FormValue("history")!="",dbCon)
					if err != nil {
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get tripMan.JSONSelectTrip :( ", err), "", err, true)
					}
					tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

					marshaled, err = json.Marshal(res)
					if err != nil {
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
					}
				} // t=trip

			case "tripIds":
				{
					var minTime int64
					var maxTime int64
					if r.FormValue("minTime") != "" {
						minTime, _ = strconv.ParseInt(r.FormValue("minTime"), 10, 64)
					} else {
						minTime = 0
					}

					if r.FormValue("maxTime") != "" {
						maxTime, _ = strconv.ParseInt(r.FormValue("maxTime"), 10, 64)
					} else {
						maxTime = math.MaxInt64
					}
					deviceIds, err := tools.GetDeviceIds(r)
					if err != nil {
						dbg.W(TAG, "invalid call jsonApi?ajax=trackIds ", err)
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not parse device Id", err), "", err, true)
					}
					var res models.JSONSelectAnswer
					if len(deviceIds) > 0 {
						res, err = tripMan.JSONSelectTripIdsInTimeframe(minTime, maxTime, deviceIds, dbCon)
						if err != nil {
							return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get tripMan.JSONSelectTripIds :( ", err), "", err, true)
						}
					} else {
						res = models.GetBadJSONSelectAnswer("No devices selected")
					}
					tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

					marshaled, err = json.Marshal(res)
					if err != nil {
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
					}
				} // t=tripIds

			case "trips":
				{
					var minTime int64
					var maxTime int64
					if r.FormValue("minTime") != "" {
						minTime, _ = strconv.ParseInt(r.FormValue("minTime"), 10, 64)
					} else {
						minTime = 0
					}

					if r.FormValue("maxTime") != "" {
						maxTime, _ = strconv.ParseInt(r.FormValue("maxTime"), 10, 64)
					} else {
						maxTime = math.MaxInt64
					}
					deviceIds, err := tools.GetDeviceIds(r)
					if err != nil {
						dbg.W(TAG, "invalid call jsonApi?ajax=trackIds ", err)
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not parse device Id", err), "", err, true)
					}

					var nots *[]*notificationManager.Notification
					nots,err = notificationChecker.GetActiveNotificationsForUser(usr.Id(),dbCon,false)
					if err != nil {
						dbg.E(TAG,"Error getting active notifications for user %d : ",usr.Id(), err)
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get GetActiveNotificationsForUser :( ", err), "", err, true)
					}

					dbg.I(TAG, "Parameters given for ajax=trips : minTime: %d, maxTime: %d, device: %d", minTime, maxTime, deviceIds)
					includeTracks := r.FormValue("includeTracks") != ""
					res, err := tripMan.JSONSelectTripsInTimeframe(minTime, maxTime, deviceIds, includeTracks, true,usr.Id(),nots,T,r.FormValue("history")!="", dbCon)
					if err != nil {
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get tripMan.JSONSelectTripIds :( ", err), "", err, true)
					}
					tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

					marshaled, err = json.Marshal(res)
					if err != nil {
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
					}
				} // t=trips

			} // case action=read: switch formValue("t")

		case "update":
			{
				var err error
				var res tripManModels.JSONUpdateTripAnswer
				anything := false
				var nots *[]*notificationManager.Notification
				nots,err = notificationChecker.GetActiveNotificationsForUser(usr.Id(),dbCon,false)
				if err != nil {
					dbg.E(TAG,"Error getting active notifications for user %d : ",usr.Id(), err)
					return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get GetActiveNotificationsForUser :( ", err), "", err, true)
				}
				if r.FormValue("trip") != "" {
					anything = true
					res, err = tripMan.JSONUpdateTrip(r.FormValue("trip"), true, usr.Level()==1337,nots,T, dbCon)
					if err != nil {
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get tripMan.JSONSelectTrip :( ", err), "", err, true)
					}
				} else {
					if r.FormValue("trips") != "" {
						anything = true
						res, err = tripMan.JSONUpdateTrips(r.FormValue("trips"), true,usr.Level()==1337,nots,T, dbCon)
						if res.UpdatedNotifications {
							notificationChecker.UpdateOverDue(usr.Id(),dbCon)
						}
					}
				}
				if !anything {
					res = tripManModels.JSONUpdateTripAnswer{JSONUpdateAnswer: models.JSONUpdateAnswer{JSONAnswer: models.GetBadJSONAnswer("no trips given.")}}
				}
				if res.UpdatedNotifications {
					err = notificationChecker.UpdateOverDue(usr.Id(),dbCon)
					if err != nil {
						dbg.E(TAG,"Error updating overdue while trip-update for user  : ",usr.Id(), err)
						webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not UpdateOverDue :( ", err), "", err, true)
					}
				}
				tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

				marshaled, err = json.Marshal(res)
				if err != nil {
					return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
				}
			} // action=read

		case "create":
			{
				var nots *[]*notificationManager.Notification
				nots,err = notificationChecker.GetActiveNotificationsForUser(usr.Id(),dbCon,false)
				if err != nil {
					dbg.E(TAG,"Error getting active notifications for user %d : ",usr.Id(), err)
					return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get GetActiveNotificationsForUser :( ", err), "", err, true)
				}
				res, err := tripMan.JSONCreateOrReviveTripByTrackIds(r.FormValue("trackIds"), true,nots,T, dbCon)
				if err != nil {
					return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get tripMan.JSONCreateOrReviveTripByTrackIds :( ", err), "", err, true)
				}
				if res.UpdatedNotifications {
					err = notificationChecker.UpdateOverDue(usr.Id(),dbCon)
					if err != nil {
						dbg.E(TAG,"Error updating overdue while trip-create for user  : ",usr.Id(), err)
						return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not UpdateOverDue :( ", err), "", err, true)
					}
				}

				tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

				marshaled, err = json.Marshal(res)
				if err != nil {
					return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
				}
			}
		case "getDevices":
			{
				res, err := deviceManager.GetDevices(dbCon)
				marshaled, err = json.Marshal(res)
				if err != nil {
					return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
				}
			}

		} // switch ajaxRequest

		// handle faulty JSONrequests
		mString := string(marshaled)
		if mString == "" {
			marshaled, err = json.Marshal(models.GetBadJSONAnswer("Unknown action"))
			if err != nil {
				return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal models.GetBadJSONAnswer(\"Unknown action\") \n error: %v", err), "", err, true)
			}
			mString = string(marshaled)
		}
		vd.Data = map[string]interface{}{"Message": template.HTML(mString)}
		viewPath = "views/showDataMessage.htm"
		if dbg.Develop {
			dbg.D(TAG, "ajaxstring is %s", ajaxRequest, r)
		}
		return
	} // if ajaxrequest != ""

	vd.Data["Message"] = "Please provide a method"
	viewPath = "views/showDataMessage.htm"
	return
}
