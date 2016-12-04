package controllers

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"net/http"

	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/datapolish"
	"strings"
)

const TAG = dbg.Tag("goodl/odl.go")

// OdlController displays the main-odl-page.
type OdlController struct {
}

// GetViewData returns a view displaying the ODL-mainpage.
func (OdlController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), false)
		}
	}()

	dbg.D(TAG, "I am at OdlController")
	T := ctx.Value("T").(*Translater)

	mdl := webfw.Model{}
	vd = webfw.ViewData{
		T:     T,
		Model: mdl,
		Data:make(map[string]interface{}),
	}

	if datapolish.LocConfig == nil {
		datapolish.LocConfig = datapolish.GetDefaultLocationConfig()
	}
	vd.Data["BASEURL"] = r.URL.Path[0:strings.Index(r.URL.Path,"/odl/")] + "/odl"
	viewPath = "views/odl.html"
	return
}
