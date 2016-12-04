package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
)

// ResetPasswordController is responsible for handling the password-reset-process.
type ResetPasswordController struct {
}

const TAG = dbg.Tag("goodl/resetPassword.go")

// GetViewData sends an password-reset-mail if only "reset_email" is given.
// If "reset_email", "resetKey" and "reset_password" are given and correct, the password gets changed.
func (ResetPasswordController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	dbg.D(TAG, "I'm at ResetPasswordController")
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	viewPath = "views/showDataMessage.htm"
	T := ctx.Value("T").(*Translater)
	resetKey := r.FormValue("resetKey")
	mdl := webfw.Model{}
	vd = webfw.ViewData{
		Model: &mdl,
		T:     T,
		Data:  make(map[string]interface{}),
	}
	email := r.FormValue("reset_email")
	// Make it slow because of bruteforce stuff
	time.Sleep(1 * time.Second)
	if resetKey == "" {
		if email != "" {
			var key string
			key, err = userManager.GetPasswordResetKey(email)
			if err != nil || key == "" {

				if err == sql.ErrNoRows {
					dbg.W(TAG, "User not found for reset mail ", err)
					vd.Data["Message"] = "UserNotFound"
				} else {
					dbg.E(TAG, "Damnit - error for GetPasswordResetKey - Key : ,  Error : ", key, err)
					vd.Data["Message"] = "Unknown error"
				}
				err = nil
				return
			}

			err = userManager.SendResetMail(email, key, T)
			if err != nil || key == "" {
				dbg.E(TAG, "Damnit - error for SendResetMail", err)
				vd.Data["Message"] = "Unknown error"
				return
			}
			vd.Data["Message"] = "success"
		}
	} else {

		keyValid, _ := userManager.CheckResetKeyValidity(resetKey, email)
		if !keyValid {

			vd.Data["Message"] = "resetKeyInvalid"
			err = nil
			return
		}
		if pw := r.FormValue("reset_password"); pw != "" {

			pw, err = userManager.GetDecryptedPw(pw, r)
			if err != nil {
				dbg.E(TAG, "Error getting decrypted pw", dbg.GetRequest(r))
				vd.Data["Message"] = "Internal server error"
				return
			}
			succ, msg := userManager.VerifyPassword(pw, pw)
			if !succ {
				vd.Data["Message"] = template.HTML(msg)
				err = nil
				return
			}

			succ = userManager.ChangePassword(email, pw, resetKey)

			if !succ {
				vd.Data["Message"] = "resetKeyInvalid"
				return
			} else {
				vd.Data["Message"] = "success"
				return
			}
		} else {
			vd.Data["Message"] = "error_insecurePassword"
			return
		}

	}

	return
}
