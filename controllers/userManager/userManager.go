package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"
	"golang.org/x/net/context"
)

const TAG = dbg.Tag("goodl/userManagerController.go")

func HandleUserManRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
	}()
	dbg.D(TAG, "UserManagerController called")
	T := ctx.Value("T").(*Translater)

	user, _, _ := userManager.GetUserWithSession(r)
	var marshaled []byte
	var err error
	if r.FormValue("action") == "getActive" {
		uData := userManager.GetUDataFromUser(user)
		res := models.JSONSelectAnswer{
			JSONAnswer:models.GetGoodJSONAnswer(),
			Result:interface{}(uData),
		}
		marshaled, err = json.Marshal(res)
		if err != nil {
			dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
	}
	if r.FormValue("action") == "read" {
		if user.Level() != 1337 {
			dbg.WTF(TAG, "Somebody attempted to read user data without having permissions", err, user)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "Insufficient permissions", 403, true)
			return
		}
		var res userManager.JSONUsersAnswer
		res, err = userManager.JSONGetUsers(r)
		if err != nil {
			dbg.E(TAG, "Could not get userManager.JSONGetUsers :( ", err)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
	}

	if r.FormValue("action") == "create" {
		if r.FormValue("reqId") == "" { // No encryption inited
			dbg.D(TAG, "No reqId provided")
			dbg.WTF(TAG, "Somebody attempted to access userManagerController without reqId ", err, user)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "Insufficient permissions", 403, true)
			return
		}
		var res models.JSONInsertAnswer
		res, err = userManager.JSONCreateUser(r, r.FormValue("user"), T)
		if err != nil {
			dbg.E(TAG, "Could not get userManager.JSONCreateUser :( ", res)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
	}
	if r.FormValue("action") == "update" {

		var res models.JSONUpdateAnswer
		res, err = userManager.JSONUpdateUser(r, r.FormValue("user"), T)
		if err != nil {
			if err == userManager.Error_AccessDenied {
				dbg.WTF(TAG, "Somebody attempted to access userManagerController without reqId ", err, user)
				webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "Insufficient permissions", 403, true)
				return
			}
			dbg.E(TAG, "Could not get userManager.JSONUpdateAnswer :( ", err)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}

		if user.Id() == res.Id {
			u, err := userManager.GetUserById(user.Id())
			if err != nil {
				dbg.E(TAG, "Error getting updated user : ", err)
				webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
				return
			}
			userManager.UpdateUserDataForSession(r, u.Email(), w)
		}
	}
	if r.FormValue("action") == "delete" {
		if user.Level() != 1337 {
			dbg.WTF(TAG, "Somebody attempted to read user data without having permissions", err, user)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "Insufficient permissions", 403, true)
			return
		}
		var res models.JSONDeleteAnswer
		res, err = userManager.JSONDeleteUser(r, r.FormValue("user"))
		if err != nil {
			dbg.E(TAG, "Could not get userManager.JSONDeleteUser :(", err)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
	}
	mString := string(marshaled)
	if mString == "" {
		marshaled, err = json.Marshal(models.GetBadJSONAnswer("Unknown action"))
		if err != nil {
			dbg.E(TAG, "unable to marshal models.GetBadJSONAnswer(\"Unknown action\") \n error: %v", err)
			webfw.DirectShowError_NoVD(ctx, w, r, errors.New(fmt.Sprintf("%s", err)), "internal server error", 500, true)
			return
		}
		mString = string(marshaled)
	}

	w.Write(marshaled)
	return
}
