package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
)

// Tag for logmessages identifying the file (usually starts with the first letter of the current file)
const TAG = dbg.Tag("goodl/defaultHybridController.go")

type DefaultHybridController struct {
}

// we can define our own json data type - if we want to be totally standard-conform we would do this in an extra file in models-folder
type OurOwnJSONType struct {
	SomeString      string   // Remember to start with a capital letter so json can find it
	SomeStringSlice []string `json:"renamedSlice,omitempty"` // optional additional parameters (renamed, dont send if empty)
	SomeMap         map[string]int
}

// func GetViewData returns a view either displaying a html-view or ajax-data
func (DefaultHybridController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() { // Error handling, if this getviewdata panics (should not happen)
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)))
		}
	}()
	// Message we know that we started
	dbg.D(TAG, "DefaultHybridController called")

	T := ctx.Value("T").(*Translater)
	viewPath = "views/defaultHybrid.html"

	/*mdl := models.LoginModel{
		webfw.Model{},
		"",
	}*/ // If we want to use our own model use the previous
	vd = webfw.ViewData{
		//Model: mdl,
		T: T,
	}

	// get current logged in user
	usr, _, _ := userManager.GetUserWithSession(r)
	// Just for security reasons (should not happen but does not hurt) :
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to map without logging in?"), "", nil)
	}

	if r.FormValue("ajax") != "" {
		viewPath = "views/showDataMessage.htm"
		// initialise data array
		vd.Data = make(map[string]interface{})
		// Init our json-data
		jsonData := OurOwnJSONType{SomeString: "bla"}
		// add additional data
		jsonData.SomeMap = make(map[string]int)
		jsonData.SomeMap["SomeKey"] = 1337
		jsonData.SomeStringSlice = make([]string, 0)
		jsonData.SomeStringSlice = append(jsonData.SomeStringSlice, "Hello World")

		var marshaled []byte
		// Stringify this shit
		marshaled, err = json.Marshal(jsonData)
		if err != nil {

			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Failed to datapolish.GetMapDataFromDB"), "", nil)
		}
		// Use this for empty page (JSON response)
		vd.Data["Message"] = template.JS(marshaled)
		// use this for embedded in html page
		vd.Data["Message"] = template.HTML(marshaled)
		return
	} // no else needed as we return if we are in above if

	// initialise data array
	vd.Data = make(map[string]interface{})
	vd.Data["Message"] = "Hi"
	vShared = "layout.html"
	return
}
