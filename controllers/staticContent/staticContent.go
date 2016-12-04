package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/tools"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
)

const TAG = dbg.Tag("goodl/StaticContent.go")

var existingTemplates = make(map[string]bool)

// StaticContentController is responsibe for serving static content.
type StaticContentController struct {
}

// GetViewData servers data from the "views/static"-path.
func (StaticContentController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), false)
		}
	}()
	dbg.D(TAG, "StaticContentController called")
	T := ctx.Value("T").(*Translater)

	staticPath := ctx.Value("staticView").(string)
	vd = webfw.ViewData{
		T:        T,
		Data:     make(map[string]interface{}),
		ViewName: "static/" + staticPath,
	}
	vd.Data["BASEURL"] = "../" + staticPath
	viewPath, _ = tools.GetCleanFilePath(staticPath, "views/static/")
	return
}
