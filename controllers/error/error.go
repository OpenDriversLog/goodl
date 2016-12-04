package controllers

import (
	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/webfw"
)

const TAG = dbg.Tag("goodl/error.go")

var Vulcanized = false

type ErrorController struct {
}

// GetViewData returns an page showing the errorMessage given in ViewData
func (ErrorController) GetViewData(vd webfw.ViewData, errIn error) (vdNew webfw.ViewData, viewPath string, vShared string, err error) {
	vdNew = vd
	if vdNew.T == nil {
		dbg.E(TAG, "GetViewData called with nil vd.T - this should not happen. Still displaying default error page (german)")
		vdNew.T = webfw.DefaultTranslater()
	}
	dbg.W(TAG, "ErrorController called with message : \n %s \n and error : \n %v ", vdNew.ErrorMessage, errIn)
	if vdNew.T != nil && vdNew.ErrorMessage != "" { // Translate if errorMessage is a string (and not e.g. Template.HTML)
		if eString, ok := vdNew.ErrorMessage.(string); ok {
			vdNew.ErrorMessage = vdNew.T.T(eString)
		}
	}

	//TODO: Remove dirty workaround (because of special calling of ViewDataPolishFunc in webfw for Errors)
	if Vulcanized {
		viewPath = "views/vulcanized/odl.html"
	} else {
		viewPath = "views/odl.html"
	}

	vdNew.ViewName = ""

	return
}
