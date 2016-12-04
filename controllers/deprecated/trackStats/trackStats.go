package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl/utils/userManager"

	"github.com/OpenDriversLog/goodl-lib/jsonapi/deviceManager"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
)

// Tag for logmessages identifying the file (usually starts with the first letter of the current file)
const tsTag = dbg.Tag("goodl/trackStats.go")

type TrackStatsController struct {
}

// we can define our own json data type - if we want to be totally standard-conform we would do this in an extra file in models-folder
type TrackStatsJSONType struct {
	SomeString      string   // Remember to start with a capital letter so json can find it
	SomeStringSlice []string `json:"renamedSlice,omitempty"` // optional additional parameters (renamed, dont send if empty)
	SomeMap         map[string]int
}

// func GetViewData returns a view either displaying a html-view or ajax-data
func (TrackStatsController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() { // Error handling, if this getviewdata panics (should not happen)
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(tsTag, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	// Message we know that we started
	dbg.D(tsTag, "TrackStatsController called")

	T := ctx.Value("T").(*Translater)

	/*mdl := models.LoginModel{
		webfw.Model{},
		"",
	}*/ // If we want to use our own model use the previous
	vd = webfw.ViewData{
		//Model: mdl,
		T:    T,
		Data: make(map[string]interface{}),
	}

	// get current logged in user
	usr, _, _ := userManager.GetUserWithSession(r)
	// Just for security reasons (should not happen but does not hurt) :
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(tsTag, 500, dbg.WTF(tsTag, "This is impossible. How did I get to trackStats without logging in?"), "", nil, true)
	}

	database, err := userManager.GetLocationDb(usr.Id())
	defer database.Close()

	ajaxRequest := r.FormValue("ajax")
	if ajaxRequest != "" {
		viewPath = "views/showDataMessage.htm"
		if dbg.Develop {
			dbg.D(tsTag, "ajaxstring is %s", ajaxRequest, r)
		}
		handleAjax(ajaxRequest, &vd, r, database)
		return
	} else { // no else needed as we return if we are in above if
		vd.Data["Message"] = "Hi, we get some stats here soonish"
	}

	t2 := time.Now().UnixNano()
	res, err := deviceManager.GetDevices(database)
	t1 := time.Now().UnixNano()

	dbg.D(tsTag, "ms took for GetDevices : %s", (t1-t2)/1000/1000)
	if err != nil {
		dbg.E(tsTag, "failed to get Devices json", err)
	}

	t1 = time.Now().UnixNano()
	lastMig, cDevices, cTRs, cTs, cKPs, err := userManager.GetLocationDbNumbers(usr.Id())
	t2 = time.Now().UnixNano()
	dbg.D(tsTag, "ms took for GetLocationDbNumbers : %s", (t2-t1)/1000/1000)

	vd.Data["DbVersion"] = lastMig
	vd.Data["CDevices"] = cDevices
	vd.Data["CTrackRecords"] = cTRs
	vd.Data["CTracks"] = cTs
	vd.Data["CKeyPoints"] = cKPs

	devmapJSON, err := json.Marshal(res)

	vd.Data["DeviceMap"] = template.JS(devmapJSON)
	vd.Data["DeviceMapHTML"] = template.HTML(devmapJSON)

	viewPath = "views/trackStats.html"
	vShared = "layout.html"
	return
}

// handles ajax requests for trackStats.go
func handleAjax(ajaxRequest string, vd *webfw.ViewData, r *http.Request, db *sql.DB) error {
	var marshaled []byte
	var err error

	if dbg.Develop {
		dbg.D(tsTag, "ajaxstring is %s", ajaxRequest, r)
	}
	if err != nil {
		return err
	}
	// Use this for empty page (JSON response)
	vd.Data["Message"] = template.HTML(marshaled)
	return err

}
