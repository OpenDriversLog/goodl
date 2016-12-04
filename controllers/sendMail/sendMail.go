package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
)

const TAG = dbg.Tag("goodl/sendMail.go")

// SendMailController is responsible for sending E-Mails to the support-team as well as creating issues for support-requests.
// TODO: Remove hardcoded "info@opendriverslog.de" and make it configurable!
// !!!TO BE DISCUSSED!!!
type SendMailController struct {
}

var sendMailTemplate *template.Template

// GetViewData sends the mail with the given issue, and creates an issue in gitlab, if possible.
func (SendMailController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	dbg.D(TAG, "SendMailController called")
	T := ctx.Value("T").(*Translater)
	usr, _, _ := userManager.GetUserWithSession(r)

	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to map SendMailController logging in?"), "", nil, true)
	}

	anrede := userManager.GetAnrede(usr, T)
	dbg.D(TAG, "Got Anrede")
	mailModel := models.SendUserMailModel{
		webfw.Model{},
		&models.SendUserMailEnhance{
			UserId: usr.Id(),
			Anrede: anrede,
			Text:   template.HTML(r.FormValue("text")),
			Mail:   usr.Email(),
			T:      T,
		},
	}
	dbg.D(TAG, "Inited Model")
	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}
	dbg.D(TAG, "Inited ViewData")
	if sendMailTemplate == nil {
		var txt []byte
		txt, err = ioutil.ReadFile(webfw.Config().RootDir + "/views/mailFromLoggedInUser.htm")
		if err != nil {
			dbg.E(TAG, "Error reading sendMailTemplate template : ", err)
			return
		}
		sendMailTemplate = template.Must(template.New("sendMailTemplate").Delims("{[{", "}]}").Parse(string(txt)))
	}
	dbg.D(TAG, "Inited template ", sendMailTemplate)
	buffer := new(bytes.Buffer)
	err = sendMailTemplate.Execute(buffer, &mailModel)
	dbg.D(TAG, "Executed template")
	if err != nil {
		dbg.E(TAG, "Error filling sendMail template : ", err)
		return
	}
	viewPath = "views/showDataMessage.htm"

	if r.FormValue("text") == "" {
		vd.Data["Message"] = template.HTML(T.T("EmptyMessage"))
		return
	}

	err = tools.SendODLMail([]string{"info@opendriverslog.de"}, "We got contacted by "+usr.Email(), string(buffer.Bytes()), true)

	if err == nil {
		vd.Data["Message"] = "Success"
	} else {
		dbg.E(TAG, "Error sending Mail in SendMailController : ", err)
		vd.Data["Message"] = T.T("Unknown error")
	}
	return
}
