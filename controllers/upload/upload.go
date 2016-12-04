package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl/utils/userManager"

	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/crypto/bcrypt"
	"github.com/OpenDriversLog/goodl/controllers/syncDB"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/deviceManager"
)

// UploadController is responsible for managing the automatic upload process.
type UploadController struct {
}

const TAG = dbg.Tag("goodl/ctrl/Upload.go")

// GetViewData takes a GUID and Password to be matched to the device keys and uploads
// the given data into the matching users database.
func (UploadController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	//r.ParseMultipartForm(200 * 1024 * 1024) // Maximum memory usage of 200 mb :0
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	dbg.D(TAG,"UploadController called.")
	T := ctx.Value("T").(*Translater)

	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}
	viewPath = "views/showDataMessage.htm"
	if r.FormValue("GUID") == "" {
		vd.Data["Message"] = "No GUID given."
		return
	} else if r.FormValue("Password") == "" {
		vd.Data["Message"] = "No Password given."
		return
	}
	guid := r.FormValue("GUID")
	var uId int64

	key, err := userManager.GetKeyByGuid(guid)
	if err != nil {
		vd.Data["Message"] = "Device not set up."
		err = nil
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(key.Password), []byte(r.FormValue("Password")))
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.I(TAG, "Key: %v, Password did not match : %v", key.GUID, err), "wrong password", nil, true)

	}
	uId = int64(key.UserId)
	if uId==0 {
		return webfw.GetErrorViewData(TAG, 500, dbg.I(TAG, "key: %v, no UserId assigned, key.GUID", err), "No userId assigned to device.", nil, true)
	}
	vd.Data["Message"] = "Unknown issue in UploadController"

	dbCon, err := userManager.GetLocationDb(uId)
	defer dbCon.Close()
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "uid: %d, no Db in UploadController : %v", uId, err), "", nil, true)
	}
	usr, err := userManager.GetUserById(uId)
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "uid: %d for Key %v, User not found!", uId,key.GUID, err), "", nil, true)

	}

	var device *deviceManager.Device
	device,err = deviceManager.GetDeviceByGUID(dbCon,key.GUID)
	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG,"Error getting device with guid %v : ",key.GUID,err), "No device assigned", nil, true)

	}
	return controllers.UploadForUser(string(device.Description),r,viewPath,&vd,vShared,T,usr,dbCon)
}
