package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/datapolish"
	"github.com/OpenDriversLog/goodl-lib/jsonapi"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/deviceManager"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/tripMan"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"github.com/OpenDriversLog/goodl/utils/notificationChecker"
)

// Tag for logmessages identifying the file (usually starts with the first letter of the current file)
const TAG = dbg.Tag("goodl/c/jsonApi.go")

// JsonApiController provides several JSON-methods for managing devices, notifications, tracks & trips
type JsonApiController struct {
}

// GetViewData provides several JSON-methods for managing devices, notifications, tracks & trips (I don't like this thing at all!)
// possible options for Formvalue "ajax":
// "track" - returns the Track with the given "id"
// "trackIds" - returns the trackIds in the timeFrame from "minTime" to "maxTime" for all devices.
// "points" - returns gets all TrackPoints for the track with the given "id"
// "lastKP" - returns the last KeyPoint before the "before"-timestamp
// "trip" - returns the trip with the given "id"
// "tripIds" - returns the tripIds in the given "minTime" to "maxTime" for all devices.
// "device" - NOT YET IMPLEMENTED - returns all tracks the device with the given "id"
// "devices" - returns all devices.
// "reprocess" - if the requesting user has admin-permissions, reproccesses the tracks in the given timespan from "minTime" to "maxTime" for the given "device"-id, recreating all trips
// "minMaxTime" - get minimum & maximum timestamp for all devices

func (JsonApiController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() { // Error handling, if this getviewdata panics (should not happen)
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()

	// Message we know that we started
	dbg.D(TAG, "JsonApiController called")

	T := ctx.Value("T").(*Translater)

	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}

	// get current logged in user
	usr, _, _ := userManager.GetUserWithSession(r)
	// Just for security reasons (should not happen but does not hurt) :
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to jsonApi without logging in?"), "", nil, true)
	}

	database, err := userManager.GetLocationDb(usr.Id())
	defer database.Close()

	ajaxRequest := r.FormValue("ajax")
	if ajaxRequest != "" {
		viewPath = "views/showDataMessage.htm"
		if dbg.Develop {
			dbg.D(TAG, "ajaxstring is %s", ajaxRequest, r)
		}
		handleAjax(ajaxRequest, &vd, r,usr.Id(),T, database)
		return
	}
	return webfw.GetErrorViewData(TAG, 404, dbg.W(TAG, "Wrong jsonApi call"), "", errors.New("Unknown request"), true)
}

// handleAjax handles ajax requests for jsonApi.go - for possible options see "GetViewData"
func handleAjax(ajaxRequest string, vd *webfw.ViewData, r *http.Request,uId int64,T *Translater, db *sql.DB) error {
	var marshaled []byte
	var err error
	if dbg.Develop {
		dbg.D(TAG, "ajaxstring is %s", ajaxRequest, r)
	}
	var id int

	switch ajaxRequest {
	case "track":
		if r.FormValue("id") != "" {
			id, _ = strconv.Atoi(r.FormValue("id"))
		} else {
			id = 1
		}
		marshaled, err = jsonapi.GetTrackById(db, int64(id))

	case "trackIds":
		var minTime int64
		var maxTime int64
		deviceIds, err := tools.GetDeviceIds(r)
		if err != nil {
			dbg.W(TAG, "invalid call jsonApi?ajax=trackIds ", err)
			return errors.New("Could not parse device Id")
		}
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
		marshaled, err = jsonapi.GetTrackIdsInTimeRange(minTime, maxTime, deviceIds, db)
		var unm map[string]interface{}
		err = json.Unmarshal(marshaled, &unm)
		dbg.D(TAG, "got trackids for ", minTime, maxTime, id, unm)

	case "points":
		if r.FormValue("id") != "" {
			id, _ = strconv.Atoi(r.FormValue("id"))
		} else {
			id = 2
		}
		marshaled, err = tripMan.JSONGetTrackPointsForTrack(db, int64(id))

	case "lastKP":
		var startTime int64
		if r.FormValue("device") != "" {
			id, _ = strconv.Atoi(r.FormValue("device"))
		} else {
			dbg.W(TAG, "invalid call jsonApi?ajax=lastKP ")
			return errors.New("no deviceId given")
		}
		if r.FormValue("before") != "" {
			startTime, _ = strconv.ParseInt(r.FormValue("before"), 10, 64)
		} else {
			startTime = 0
		}

		marshaled, err = jsonapi.GetLastKeyPointForDeviceBefore(int64(id), startTime, db)
		var unm map[string]interface{}
		err = json.Unmarshal(marshaled, &unm)
		dbg.D(TAG, "got last KP for ", id, startTime, unm)

	case "trip":
		if r.FormValue("id") != "" {
			id, _ = strconv.Atoi(r.FormValue("id"))
		} else {
			id = 1
		}
		nots,err := notificationChecker.GetActiveNotificationsForUser(uId,db,false)
		if err != nil {
			dbg.E(TAG,"Error getting active notifications for user %d : ",uId, err)
			return err
		}
		includeTracks := r.FormValue("includeTracks") != ""
		res, _ := tripMan.JSONSelectTrip(int64(id), includeTracks, true,nots,T,r.FormValue("history")!="", db)
		marshaled, err = json.Marshal(res)

	case "tripIds":
		var minTime int64
		var maxTime int64
		deviceIds, err := tools.GetDeviceIds(r)
		if err != nil {
			dbg.W(TAG, "invalid call jsonApi?ajax=trackIds ", err)
			return errors.New("Could not parse device Id")
		}
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
		res, _ := tripMan.JSONSelectTripIdsInTimeframe(minTime, maxTime, deviceIds, db)
		dbg.D(TAG, "got trackids for ", minTime, maxTime, id, res)

		marshaled, err = json.Marshal(res)

	case "device":
		if r.FormValue("id") != "" {
			id, _ = strconv.Atoi(r.FormValue("id"))
		} else {
			id = 2
		}
		marshaled, err = jsonapi.GetTracksByDevice(db, int64(id))

	case "devices":
		if r.FormValue("id") != "" {
			id, _ = strconv.Atoi(r.FormValue("id"))
		} else {
			id = 2
		}
		res, err := deviceManager.GetDevices(db)
		marshaled, err = json.Marshal(res)

		if err != nil {
			dbg.E(TAG, "invalid call jsonApi?ajax=devices ", err)
		}

	case "reprocess":
		jsonData := models.JSONAnswer{}

		// Stringify json data for streaming
		marshaled, err = json.Marshal(jsonData)

		var minTime int64
		var maxTime int64
		var devices map[int]string
		if r.FormValue("device") != "" {
			id, err = strconv.Atoi(r.FormValue("device"))
			if err != nil {
				dbg.W(TAG, "Wrong format for device ", r.FormValue("device"))
			}
			devices = map[int]string{id: "Unknown"}
			jsonData.ErrorMessage = "bad deviceId"
			jsonData.Error = true
			break
		} else {
			devices, err = datapolish.GetDeviceStrings(db)
			if err != nil {
				dbg.E(TAG, "Error getting devices : ", err)
				jsonData.ErrorMessage = "could not get devices"
				jsonData.Error = true
				break
			}

		}
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

		nots,err := notificationChecker.GetActiveNotificationsForUser(uId,db,false)
		if err != nil {
			dbg.E(TAG,"Error getting active notifications for user %d : ",uId, err)
			return err
		}
		for deviceId, _ := range devices {
			err = datapolish.ReprocessDataForDeviceInTimeRange(minTime, maxTime, deviceId,uId,nots,T, db)
			dbg.WTF(TAG, "got reprocessed data from %d to %d for device %d", minTime, maxTime, deviceId)
			if err != nil {
				dbg.E(TAG, "Error reprocessing data : ", err)
				jsonData.ErrorMessage = "Internal server error"
				jsonData.Error = true
				break
			}
		}
		jsonData.Success = true

	case "minMaxTime":
		marshaled, err = jsonapi.GetMinMaxTime(db)

		if err != nil {
			var unm map[string]interface{}
			err = json.Unmarshal(marshaled, &unm)
			dbg.D(TAG, "got minmaxtime for all devices ", unm, err)
		}

	default:
		jsonData := models.JSONAnswer{ErrorMessage: "Unknown method"}

		// Stringify json data for streaming
		marshaled, err = json.Marshal(jsonData)
	}

	if err != nil {
		return err
	}
	// Use this for empty page (JSON response)
	vd.Data["Message"] = template.HTML(marshaled)
	return err
}
