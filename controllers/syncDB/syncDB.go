package controllers

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/datapolish"
	"github.com/OpenDriversLog/goodl-lib/dbMan"
	"github.com/OpenDriversLog/goodl/utils/dataConverter"
	"github.com/OpenDriversLog/goodl/utils/userManager"

	"github.com/OpenDriversLog/goodl-lib/tools"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"github.com/OpenDriversLog/goodl/utils/notificationChecker"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/notificationManager"
)

// SyncDBController is responsible for uploading trips to a users database.
type SyncDBController struct {
}

const TAG = dbg.Tag("goodl/ctrl/SyncDB.go")

// GetViewData takes trackdata in different formats, converts it to CSV and processes it into the users database.
// (including GeoCoding, KeyPoint determination & classification)
func (SyncDBController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	//r.ParseMultipartForm(200 * 1024 * 1024) // Maximum memory usage of 200 mb :0
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	T := ctx.Value("T").(*Translater)
	//dbg.D(TAG, "I'm at SyncDBController with request: ", r)

	usr, _, _ := userManager.GetUserWithSession(r)
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to syncDb without logging in?"), "", nil, true)
	}
	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}
	vd.Data["Message"] = "Unknown"
	viewPath = "views/showDataMessage.htm"
	dbCon, err := userManager.GetLocationDb(usr.Id())
	defer dbCon.Close()
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "uid: %d, (lvl:%d) no Db in SyncDBController : %v", usr.Id(), usr.Level(), err), "", nil, true)
	}

	return UploadForUser(r.FormValue("upload_key"),r,viewPath,&vd,vShared,T,usr,dbCon)

}

// UpöpadForUser processes the given upload-tracks, converts it to CSV, processes it
// (including GeoCoding, KeyPoint determination & classification)
// TODO : Let this use device IDs instead of string keys....
func UploadForUser(deviceKey string,r *http.Request,_viewPath string,_viewData *webfw.ViewData,_vShared string,T *Translater,usr *userManager.OdlUser,dbCon *sql.DB) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	viewPath = _viewPath
	vd = *_viewData
	vShared = _vShared
	uploadType := r.FormValue("upload_dataType")
	if uploadType=="" {
		uploadType = "NMEA/GPRMC"
	}
	if uploadType != "" { // somebody is trying to upload a different format
		dbg.V(TAG, "...called with uploadtype ", uploadType)
		viewPath = "views/showDataMessage.htm"
		var res string
		key := deviceKey
		if key == "" {
			vd.Data["Message"] = T.T("No key provided for uploadData")
			return
		}
		// row := dbCon.QueryRow("SELECT MAX(timeMillis) FROM TrackRecords where DeviceKey=(SELECT id from DEVICES WHERE desc=?)",
		// 	key)
		// var lastTs int64

		devId := GetDeviceId(key, dbCon) // find key for value
		if devId == -1 {                 // before import the device did not exist
			vd.Data["Message"] = "Please_select_device"
			return
		}
		dataex := r.FormValue("upload_externalData")
		datas := make(map[string]string) // TODO: Maybe make this less ram consuming (working with file scanners instead of strings)

		isDataEx := true

		// no form upload - see if we got a file upload! http://sanatgersappa.blogspot.de/2013/03/handling-multiple-file-uploads-in-go.html
		dbg.D(TAG, "uid: %d (lvl: %d) File upload started", usr.Id(), usr.Level())
		isDataEx = false
		m := r.MultipartForm
		files := m.File["upload_files"]
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "uid: %d (lvl: %d) Unable to read file %s : ", usr.Id(), usr.Level(), files[i].Filename, err), "", nil, true)
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(file)
			datas[files[i].Filename] = buf.String()
			//dbg.WTF(TAG,"DATA for %s: %s",files[i].Filename,datas[files[i].Filename])
		}
		if dataex != "" {
			datas["form"] = dataex
		}

		respTxt := ""
		failed := false

		var nots *[]*notificationManager.Notification
		nots,err = notificationChecker.GetActiveNotificationsForUser(usr.Id(),dbCon,false)
		if err != nil {
			dbg.E(TAG,"Error getting active notifications for user %d : ",usr.Id(), err)
			return
		}

		for dataKey, data := range datas {
			if !isDataEx && respTxt != "" {
				respTxt += "<br />" + dataKey + " : "
			}
			_, lastTs, err3 := datapolish.GetDeviceTimeRange(devId, dbCon)
			if err3 != nil {
				dbg.I(TAG, "No entries existing for app key %s for user %s", key, usr.Id())
			}
			res, err = dataConverter.ConvertAnythingToCSV(data, uploadType, lastTs+1)
			if err != nil {
				failed = true
				if err == dataConverter.Err_UnknownFormat {
					dbg.E(TAG, "uid: %d (lvl: %d) unknown format %s : ", usr.Id(), usr.Level(), uploadType)
					vd.Data["Message"] = template.HTML(respTxt + T.T("Unknown format"))
				} else if err == dataConverter.Err_NoData {
					dbg.E(TAG, "uid: %d (lvl: %d) no data %s : ", usr.Id(), usr.Level(), uploadType)
					vd.Data["Message"] = template.HTML(respTxt + T.T("No data"))
				} else {
					dbg.E(TAG, "uid: %d (lvl: %d) unknown error %s : ", usr.Id(), usr.Level(), err)
					vd.Data["Message"] = template.HTML(respTxt + T.T("Unknown error"))
				}
				err = nil
				return
			}

			//dbg.W("File content for file %s: %s", dataKey, res)

			var cnt int
			var minTime int64
			var maxTime int64
			cnt, minTime, maxTime, err = dbMan.InsertCSVToDb(res, key, usr.Id(), dbCon)

			if err != nil {
				failed = true
				dbg.E(TAG, "uid: %d (lvl: %d) Error on dbMan.InsertCSVToDb : ", usr.Id(), usr.Level(), err)
				vd.Data["Message"] = template.HTML(respTxt + T.T("Unknown error"))
				return
			}
			dbg.I(TAG, "uid: %d (lvl: %d) Inserted %d new entries from external from %d to %d for key %s", usr.Id(), usr.Level(), cnt, minTime, maxTime, key)

			if cnt == 0 {
				failed = true
				vd.Data["Message"] = template.HTML(respTxt + T.T("No new data provided"))
				return
			}
			err = datapolish.ProcessGPSData(minTime, maxTime, devId, false,usr.Id(),nots,T, dbCon)
			if err != nil {
				failed = true
				dbg.E(TAG, "uid: %d (lvl: %d) Error on datapolish.ProcessGPSData : ", usr.Id(), usr.Level(), err)
				if err == datapolish.ErrGpsDataAlreadyImported {
					vd.Data["Message"] = template.HTML(respTxt + T.T("Data already imported"))
				} else {
					vd.Data["Message"] = template.HTML(respTxt + T.T("Unknown error"))
				}
				err = nil
				return
			}
			dbg.I(TAG, "uid: %d (lvl: %d) Finished ProcessGPSData for %s", usr.Id(), usr.Level(), dataKey)
			respTxt += T.T("Success")
		}

		err = notificationChecker.UpdateOverDue(usr.Id(),dbCon)
		if err != nil {
			dbg.E(TAG,"Error updating overdue while ProcessGPSData for user  : ",usr.Id(), err)
			webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not UpdateOverDue :( ", err), "", err, true)
		}

		dbg.I(TAG, "uid: %d (lvl: %d) Finished ProcessGPSData", usr.Id(), usr.Level())
		if !failed {
			vd.Data["Message"] = "success"
		} else {
			vd.Data["Message"] = template.HTML(respTxt)
		}
		// Processdata mit deviceid + startTimeStamp
		return
	} // if r.FormValue("upload_dataType") != ""

	getLastTsForKey := r.FormValue("getLastTSForKey")
	if getLastTsForKey != "" { // Somebody is requesting our last timestamp for a device
		// TODO: FS @CS why would goodl query trackrecords DB directly??? it is supposed to have no knowledge about the db or its structure at all
		row := dbCon.QueryRow("SELECT MAX(timeMillis) FROM TrackRecords WHERE deviceId=(SELECT _deviceId FROM devices WHERE desc=?)",
			getLastTsForKey)
		var timeMillis int64

		if err := row.Scan(&timeMillis); err != nil {
			dbg.I(TAG, "No entries existing for app key %s for user %s", getLastTsForKey, usr.Id())
		}

		vd.Data["Message"] = timeMillis
		return
	}

	getDevices := r.FormValue("getDevices")
	if getDevices != "" { // Somebody wants the list of devices

		deviceMap, err2 := datapolish.GetDeviceStrings(dbCon)

		err = err2
		if err != nil && err != sql.ErrNoRows {

			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "getDevices query failed : ", err), "", nil, true)
		}
		// defer rows.Close()
		var resp = "id,descr"
		// for rows.Next() {
		// 	var id sql.NullInt64
		// 	var descr sql.NullString
		// 	err = rows.Scan(&id, &descr)
		// 	if err != nil {

		// 		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not scan getDevices-row."), "", nil)
		// 	}
		for key, value := range deviceMap {
			resp += fmt.Sprintf("\r\n%v,%v", key, value)
		}

		vd.Data["Message"] = resp
		return
	}

	getDataSince := r.FormValue("getDataSince")
	getDataBefore := r.FormValue("getDataBefore")
	if getDataSince != "" || getDataBefore != "" { // somebody wants to get csv-data for a device in a given timeframe
		deviceKey := r.FormValue("deviceKey")

		if deviceKey == "" {
			vd.Data["Message"] = "No deviceKey given"
			return
		}
		if getDataSince == "" {
			getDataSince = "0"
		}
		if getDataBefore == "" {
			getDataBefore = "999999999999999999"
		}

		var sinceData int64
		sinceData, err = strconv.ParseInt(getDataSince, 10, 64)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 422, dbg.I(TAG, "Bad format for GetDataSince"), T.T("syncDB_error_bad_formatGDS"), nil, true)
		}
		var beforeData int64
		beforeData, err = strconv.ParseInt(getDataBefore, 10, 64)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 422, dbg.I(TAG, "Bad format for GetDataBefore"), T.T("syncDB_error_bad_formatGDB"), nil, true)

		}
		var dataSince string

		var dbRoot = webfw.Config().SharedDir + "/upload/" + fmt.Sprintf("%d", usr.Id())
		// Its a bit of dirty that we need this path below for datapolish instead of using dbMan.GetLocationDb, but ok
		dbPath, _ := tools.GetCleanFilePath("trackrecords.db", dbRoot)

		dataSince, err = dbMan.GetDbCSV(dbPath, sinceData, beforeData, deviceKey,usr.Id())
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "CSV not created in SyncDBController (2) : %+v", err), T.T("syncDB_error_bad_formatCSV"), nil, true)

		}
		vd.Data["Message"] = dataSince
		return
	}
	uploadData := r.FormValue("uploadData")
	if uploadData != "" { // somebody wants to upload a csv

		key := r.FormValue("key")
		if key == "" {
			return webfw.GetErrorViewData(TAG, 422, dbg.I(TAG, "No key provided for uploadData"), T.T("syncDB_error_no_key"), nil, true)

		}

		var cnt int
		var minTime int64
		var maxTime int64
		cnt, minTime, maxTime, err = dbMan.InsertCSVToDb(uploadData, key, usr.Id(), dbCon)
		if minTime < 1 {
			dbg.WTF(TAG, "uid: %d (lvl: %d) How can dbMan.InsertCSVToDb return minTime of %d???", usr.Id(), usr.Level(), minTime)
			err = errors.New("Bad data")
		}
		if err != nil {
			switch err {
			case dbMan.EMissingHeading:
				{
					return webfw.GetErrorViewData(TAG, 422, dbg.I(TAG, "Missing heading for uploadData"), T.T("syncDB_error_missing_header"), nil, true)
				}
			case dbMan.EForbiddenHeading:
				{
					return webfw.GetErrorViewData(TAG, 422, dbg.I(TAG, "Forbidden heading for uploadData"), T.T("syncDB_error_forbidden_header"), nil, true)
				}
			default:
				{
					return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Error on insert for uploadData"), T.T("syncDB_error_insert_data"), nil, true)
				}
			}
			return
		}

		dbg.I(TAG, "Inserted %d new entries from %d to %d for key %v", cnt, minTime, maxTime, key)
		devId := GetDeviceId(key, dbCon)
		var nots *[]*notificationManager.Notification
		nots,err = notificationChecker.GetActiveNotificationsForUser(usr.Id(),dbCon,false)
		if err != nil {
			dbg.E(TAG,"Error getting active notifications for user %d : ",usr.Id(), err)
			return
		}
		err = datapolish.ProcessGPSData(minTime, maxTime, devId, false,usr.Id(),nots,T, dbCon)
		if err != nil {
			dbg.E(TAG, "Error on datapolish.ProcessGPSData : ", err)
			if err == datapolish.ErrGpsDataAlreadyImported {
				vd.Data["Message"] =  T.T("Data already imported")
			} else {
				vd.Data["Message"] = T.T("Unknown error")
			}
			return
		}
		err = notificationChecker.UpdateOverDue(usr.Id(),dbCon)
		if err != nil {
			dbg.E(TAG,"Error updating overdue while ProcessGPSData for user  : ",usr.Id(), err)
			webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not UpdateOverDue :( ", err), "", err, true)
		}

		dbg.I(TAG, "Finished ProcessGPSData")
		vd.Data["Message"] = "success"
		return
	}
	return
}
func GetDeviceId(key string, dbCon *sql.DB) (devId int) {
	devmap, _ := datapolish.GetDeviceStrings(dbCon)
	devId = -1
	for devmapKey, val := range devmap {
		if val == key {
			devId = devmapKey
		}
	}
	return
}
