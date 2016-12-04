package userManager

import (
	"encoding/json"
	"errors"
	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl-lib/translate"
	"net/http"
	"strings"
)

// Error_AccessDenied will occur when the permissions are not sufficient.
var Error_AccessDenied = errors.New("Access denied")

// JSONUsersAnswer is the default answer to a SELECT-request for users.
type JSONUsersAnswer struct {
	models.JSONAnswer
	Users []*AdminEditableUserData
}

// JSONKeysAnswer is the default answer to an SELECT-request for the current deviceKeys.
type JSONKeysAnswer struct {
	models.JSONAnswer
	Keys []*Key
}

// JSONKeysAnswer is the default answer to an delete-request for an deviceKey.
type JSONKeyDeleteAnswer struct {
	models.JSONDeleteAnswer
	Guid string
}

// JSONGetUsers returns all users, if the requesting user is an admin.
func JSONGetUsers(r *http.Request) (res JSONUsersAnswer, err error) {
	user, _, err := GetUserWithSession(r)
	if err != nil {
		dbg.E(TAG, "Error getting user with session", err)
		res = GetBadJsonUsersManAnswer("internal server error")
		return
	}
	if user == nil || user.Level() != 1337 {
		dbg.WTF(TAG, "Somebody tried to read all users without permission", user, dbg.GetRequest(r))
		err = Error_AccessDenied
		res = GetBadJsonUsersManAnswer(Error_AccessDenied.Error())
		return
	}
	res = JSONUsersAnswer{}
	res.Users, err = GetAllUsers()
	if err != nil {
		dbg.E(TAG, "Error getting GetAllUsers : ", err)
		err = nil
		res = GetBadJsonUsersManAnswer("Unknown error while getting users")
		return
	}
	res.Success = true
	return
}

// JSONGetKeys returns all current device-keys.
func JSONGetKeys() (res JSONKeysAnswer, err error) {

	res = JSONKeysAnswer{}
	res.Keys, err = GetAllKeys()
	if err != nil {
		dbg.E(TAG, "Error getting GetAllKeys : ", err)
		err = nil
		res = JSONKeysAnswer{JSONAnswer:models.GetBadJSONAnswer("Unknown error while getting keys")}
		return
	}
	res.Success = true
	return
}

// JSONDeleteKey deletes an device-key.
func JSONDeleteKey(keyJson string) (res JSONKeyDeleteAnswer, err error) {
	k := &Key{}
	if keyJson == "" {
		res = JSONKeyDeleteAnswer{JSONDeleteAnswer: models.GetBadJSONDeleteAnswer(NoDataGiven,-1)}
		return
	}
	err = json.Unmarshal([]byte(keyJson), k)
	if err != nil {
		dbg.W(TAG, "Could not read JSON %v in JSONDeleteKey : ", keyJson, err)
		res = JSONKeyDeleteAnswer{JSONDeleteAnswer: models.GetBadJSONDeleteAnswer("Invalid format",-1)}
		err = nil
		return
	}
	err = DeleteKey(k)
	if err != nil {
		dbg.W(TAG, "Error deleting key: ", keyJson, err)
		res = JSONKeyDeleteAnswer{JSONDeleteAnswer:models.GetBadJSONDeleteAnswer("internal server error",-1)}
		res.Guid = k.GUID
		err = nil
		return
	}
	res.Id = -1
	res.Guid = k.GUID
	res.RowCount = 1
	res.Success = true

	return
}

// JSONGenerateKeys generates the given amount of device-keys.
func JSONGenerateKeys(cnt int64) (res JSONKeysAnswer, err error) {
	res.Keys, err = GenerateKeys(cnt)
	if err != nil {
		dbg.E(TAG, "Error creating %d keys: ",cnt, err)
		res = JSONKeysAnswer{JSONAnswer:models.GetBadJSONAnswer("internal server error")}
		err = nil
		return
	}
	res.Success = true
	return
}

// GetBadJsonUsersManAnswer returns a bad JSONUsersAnswer in case an error occured.
func GetBadJsonUsersManAnswer(message string) JSONUsersAnswer {
	return JSONUsersAnswer{
		JSONAnswer: models.GetBadJSONAnswer(message),
	}
}

// JSONCreateUser creates a new user
func JSONCreateUser(r *http.Request, userJson string, T *translate.Translater) (res models.JSONInsertAnswer, err error) {

	if r.FormValue("reqId") == "" {
		err = Error_AccessDenied
		res = models.GetBadJSONInsertAnswer(Error_AccessDenied.Error())
		return
	}
	u := &UserEditableUserData{}
	if userJson == "" {
		res = models.GetBadJSONInsertAnswer(NoDataGiven)
		return
	}
	err = json.Unmarshal([]byte(userJson), u)
	if err != nil {
		dbg.W(TAG, "Could not read JSON %v in JSONCreateUser : ", userJson, err)
		res = models.GetBadJSONInsertAnswer("Invalid format")
		err = nil
		return
	}
	if r.FormValue("reqId") == "" { // No encryption inited
		dbg.D(TAG, "No reqId provided")
		dbg.WTF(TAG, "Somebody attempted to create user without reqId ", err, dbg.GetRequest(r))
		err = errors.New("NotEncrypted")
		return
	}
	var dcuser *OdlUser
	dcuser, err = getOpenUserFromRequest(r)
	dckey := dcuser.Nonce().PrivKey
	if err != nil {
		dbg.E(TAG, "Error getting open user from request", err)
		return
	}
	u.Password, err = getDecryptedPw(u.Password, dckey)
	if err != nil {
		dbg.E(TAG, "Error decrypting password", err)
		return
	}
	u.RepeatedPassword, err = getDecryptedPw(u.RepeatedPassword, dckey)

	if err != nil {
		dbg.E(TAG, "Error decrypting repeated password", err)
		return
	}

	var key int64
	var _errors map[string]string
	key, err, _errors = CreateUser(r, u, dcuser, r.FormValue("inviteKey"), T)
	if err != nil {
		if err.Error() == "inviteOnly" {
			err = nil
			res = models.GetBadJSONInsertAnswer("inviteOnly")
			res.Errors = _errors
			return
		}
		if err.Error() == "verificationFailed" || err.Error() == "Verification failed" {
			err = nil
			res = models.GetBadJSONInsertAnswer("verificationFailed")
			res.Errors = _errors
			return
		}
		if err.Error() == "timedout" {
			err = nil
			res = models.GetBadJSONInsertAnswer("requestTimeout")
			res.Errors = _errors
			return
		}
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			res = models.GetBadJSONInsertAnswer("user_exists")
			err = nil
			return
		}
		if err.Error() == "error_invalidEmail" {
			res = models.GetBadJSONInsertAnswer("error_invalidEmail")
			err = nil
			return
		}
		if err.Error() == "error_passwordsNotMatch" {
			res = models.GetBadJSONInsertAnswer("error_passwordsNotMatch")
			err = nil
			return
		}
		if err.Error() == "error_insecurePassword" {
			res = models.GetBadJSONInsertAnswer("error_insecurePassword")
			err = nil
			return
		}
		dbg.E(TAG, "Error in JSONCreateUser CreateUser: ", err, _errors)
		err = nil
		res.Errors = _errors
		res = models.GetBadJSONInsertAnswer("Internal server error")
		return
	}
	res.LastKey = key
	res.Success = true
	return

}

// JSONDeleteUser deletes the given user, if the requesting user has admin-irhgts
func JSONDeleteUser(r *http.Request, userJson string) (res models.JSONDeleteAnswer, err error) {
	user, _, err := GetUserWithSession(r)
	if err != nil {
		dbg.E(TAG, "Error getting user with session", err)
		res = models.GetBadJSONDeleteAnswer("internal server error", -1)
		return
	}
	if user == nil || user.Level() != 1337 {
		err = Error_AccessDenied
		res = models.GetBadJSONDeleteAnswer(Error_AccessDenied.Error(), -1)
		return
	}
	c := &AdminEditableUserData{}
	if userJson == "" {
		res = models.GetBadJSONDeleteAnswer(NoDataGiven, -1)
		return
	}
	err = json.Unmarshal([]byte(userJson), c)
	if err != nil {
		dbg.W(TAG, "Could not read JSON %v in DeleteUserJSON : ", userJson, err)
		res = models.GetBadJSONDeleteAnswer("Invalid format", -1)
		err = nil
		return
	}
	err = DeleteUser(int64(c.Id))
	if err != nil {
		dbg.E(TAG, "Error in DeleteUserJSON DeleteUser: ", err)
		errMsg := "Internal server error"
		err = nil
		res = models.GetBadJSONDeleteAnswer(errMsg, int64(c.Id))
		return
	}
	res.RowCount = -1
	res.Id = int64(c.Id)
	res.Success = true
	return

}

const NoDataGiven = "Please fill at least one entry."

// JSONUpdateUser updates the given user.
func JSONUpdateUser(r *http.Request, usrJson string, T *translate.Translater) (res models.JSONUpdateAnswer, err error) {
	user, _, err := GetUserWithSession(r)
	if err != nil {
		dbg.E(TAG, "Error getting user with session", err)
		res = models.GetBadJSONUpdateAnswer("internal server error", -1)
		return
	}
	if user == nil || !user.IsLoggedIn() {
		err = Error_AccessDenied
		dbg.WTF(TAG, "Somebody was not logged in and tried to update user", user, dbg.GetRequest(r))
		res = models.GetBadJSONUpdateAnswer(Error_AccessDenied.Error(), -1)
		return
	}
	res.Success = false
	if user == nil {
		err = Error_AccessDenied
		res.Error = true
		res.ErrorMessage = "Access denied"
		return
	}

	if usrJson == "" {
		res = models.GetBadJSONUpdateAnswer(NoDataGiven, -1)
		return
	}
	isAdmin := false
	u := &AdminEditableUserData{}
	if user.Level() == 1337 { // Admin can edit every user & every user field
		isAdmin = true
		d := &AdminEditableUserData{}
		err = json.Unmarshal([]byte(usrJson), d)
		if err != nil {
			dbg.W(TAG, "Could not read JSON %v in UpdateUsrJSON (Admin): ", usrJson, err)
			res = models.GetBadJSONUpdateAnswer("Invalid format", -1)
			err = nil
			return
		}
		u = d
		if u.Id > 0 && user.Id() != u.Id {
			user, err = GetUserById(u.Id)
			if err != nil {
				dbg.E(TAG, "Error while getting user to update : ", u, err)
				return
			}
		} else {
			u.Id = user.Id()
		}
	} else { // normal user can only edit himself.
		d := &UserEditableUserData{}
		err = json.Unmarshal([]byte(usrJson), d)
		if err != nil {
			dbg.W(TAG, "Could not read JSON %v in UpdateUsrJSON (User): ", usrJson, err)
			res = models.GetBadJSONUpdateAnswer("Invalid format", -1)
			err = nil
			return
		}
		u.Id = user.Id()
		u.Title = d.Title
		u.FirstName = d.FirstName
		u.LastName = d.LastName
		u.LoginName = d.LoginName
		u.Password = d.Password
		u.RepeatedPassword = d.RepeatedPassword
		u.TutorialDisabled = d.TutorialDisabled
		u.NotificationsEnabled = d.NotificationsEnabled

	}
	if u.Password != "" {
		if r.FormValue("reqId") == "" { // No encryption inited
			dbg.D(TAG, "No reqId provided")
			dbg.WTF(TAG, "Somebody attempted to change password without reqId ", err, user)
			err = errors.New("NotEncrypted")
			return
		}
		var dcuser *OdlUser
		dcuser, err = getOpenUserFromRequest(r)
		key := dcuser.Nonce().PrivKey
		if err != nil {
			dbg.E(TAG, "Error getting open user from request", err)
			return
		}
		u.Password, err = getDecryptedPw(u.Password, key)
		if err != nil {
			dbg.E(TAG, "Error decrypting password", err)
			return
		}
		u.RepeatedPassword, err = getDecryptedPw(u.RepeatedPassword, key)
		if err != nil {
			dbg.E(TAG, "Error decrypting repeated password", err)
			return
		}
	}

	var rowCount int64
	rowCount, err = UpdateUser(user, u, r, T, isAdmin, true, true)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			res = models.GetBadJSONUpdateAnswer("user_exists", int64(u.Id))
			err = nil
			return
		}
		if err.Error() == "NoChangesFound" {
			res = models.GetBadJSONUpdateAnswer("NoChangesFound", int64(u.Id))
			err = nil
			return
		}
		if err.Error() == "error_invalidEmail" {
			res = models.GetBadJSONUpdateAnswer("error_invalidEmail", int64(u.Id))
			err = nil
			return
		}
		if err.Error() == "error_passwordsNotMatch" {
			res = models.GetBadJSONUpdateAnswer("error_passwordsNotMatch", int64(u.Id))
			err = nil
			return
		}
		if err.Error() == "error_insecurePassword" {
			res = models.GetBadJSONUpdateAnswer("error_insecurePassword", int64(u.Id))
			err = nil
			return
		}
		dbg.E(TAG, "Error in JSONUpdateUser: ", err)
		err = nil
		res = models.GetBadJSONUpdateAnswer("internal server error", int64(u.Id))
		return
	}
	res.RowCount = rowCount
	res.Id = int64(u.Id)
	res.Success = true
	return
}
