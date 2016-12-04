// Package userManager is responsible for CRUD users, logging in users, inviting users, activating users, resetting passwords
// as well as getting and creating/updating user-databases and managing global device-keys.
// The here implemented basic encryption DOES NOT REPLACE THE REQUIREMENT OF USING SSL. Please USE SSL :>
package userManager

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/Compufreak345/alice"
	"github.com/gorilla/sessions"
	libDbMan "github.com/OpenDriversLog/goodl-lib/dbMan"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Compufreak345/dbg"

	"github.com/OpenDriversLog/webfw"
	"github.com/OpenDriversLog/webfw/login"
	. "github.com/OpenDriversLog/goodl-lib/translate"

	libtools "github.com/OpenDriversLog/goodl-lib/tools"
	"github.com/OpenDriversLog/goodl/models"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/nu7hatch/gouuid"
)

const LVL_USER = 1
const LVL_PAID = 2
const LVL_PREMIUM = 3
const LVL_ADMIN = 1337

var WrongActivationKey = errors.New("Wrong activation key")
var AccountNotActivated = errors.New("Account not activated")

// OdlUser represents the internal Odl-User.
type OdlUser struct {
	login.User
	IActivationKey    string
	ILevel            int
	ITimeEnter        int64
	ITutorialDisabled int
	INotificationsEnabled int
	INextNotificationTime int64
}

// IsLoggedIn returns true if the user is logged in.
func (u *OdlUser) IsLoggedIn() bool {
	return u.IisLoggedIn
}

// SetLevel sets the users access-level - 1337=Admin, default = 0
func (u *OdlUser) SetLevel(level int) {
	u.ILevel = level
}

// Level returns the users access-level - 1337=Admin, default = 0
func (u *OdlUser) Level() int {
	return u.ILevel
}

// TutorialDisabled returns if the user disabled the tutorial.
func (u *OdlUser) TutorialDisabled() int {
	return u.ITutorialDisabled
}

// SetTutorialDisabled sets if the user disabled the tutorial.
func (u *OdlUser) SetTutorialDisabled(tutorialDisabled int) {
	u.ITutorialDisabled = tutorialDisabled
}

// NotificationsEnabled returns if the user has notifications enabled.
func (u *OdlUser) NotificationsEnabled() int {
	return u.INotificationsEnabled
}

// SetNotificationsEnabled sets if the user has notifications enabled.
func (u *OdlUser) SetNotificationsEnabled(notificationsEnabled int) {
	u.INotificationsEnabled = notificationsEnabled
}

// NextNotificationTime returns the next time a notification for the user needs to be sent.
func (u *OdlUser) NextNotificationTime() int64 {
	return u.INextNotificationTime
}

// SetNextNotificationTime sets the next time a notification for the user needs to be sent.
func (u *OdlUser) SetNextNotificationTime(notificationTime int64) {
	u.INextNotificationTime = notificationTime
}

// SetTimeEnter sets the time when the user registered.
func (u *OdlUser) SetTimeEnter(timeEnter int64) {
	u.ITimeEnter = timeEnter
}

// TimeEnter returns the time when the user registered.
func (u *OdlUser) TimeEnter() int64 {
	return u.ITimeEnter
}

// SetActivationKey sets the activationKey for the user (if it is empty the account is activated)
func (u *OdlUser) SetActivationKey(activationKey string) {
	u.IActivationKey = activationKey
}

// ActivationKey gets the activationKey for the user (if it is empty the account is activated)
func (u *OdlUser) ActivationKey() string {
	return u.IActivationKey
}

// LoginAnswer is sent to notify if the login attempt was succesful, including a CookieVal to be used in an app.
type LoginAnswer struct {
	Success          bool
	Timedout         bool
	Error            bool
	ErrorMessage     string
	SessionCookieVal string
	Errors           map[string]string
	User map[string]interface{}
}

// openLoginUsers contains the currently logged in users
var openLoginUsers map[string]*OdlUser
var loginMutex *sync.Mutex
var userDbExists bool

// InviteRequired allows you to set if invites are required to register.
const InviteRequired = true
const TAG = "goodl/userManager.go"

// init initialises userManager
func init() {
	dbg.D(TAG, "Initting userManager.go")
	loginMutex = &sync.Mutex{}
	loginMutex.Lock()
	openLoginUsers = make(map[string]*OdlUser)
	loginMutex.Unlock()
	gob.Register(&OdlUser{})

}

// LoginWallHandler is a chained alice.CtxHandler  to block not logged in users from pages with login requirement, redirecting
// to the login-page (or sending "NotLoggedIn" if it is an ajax-request).
func LoginWallHandler(ctx context.Context, next alice.CtxHandler) alice.CtxHandler {

	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				dbg.E(TAG, "panic in LoginWallHandler: %v for request : %v", err, dbg.GetRequest(r))
				webfw.DirectShowError(webfw.ViewData{ErrorType: 500}, errors.New(fmt.Sprintf("%s", err)), w)
			}
		}()
		// Get a session

		if dbg.Develop {
			dbg.D(TAG, "I'm at LoginWallHandler for %v", r.RequestURI)
		} else {
			dbg.D(TAG, "I'm at LoginWallHandler for %v", r.URL.Path)
		}
		usr, session, err := GetUserWithSession(r)

		if usr == nil || err != nil || !usr.IsLoggedIn() || session == nil {
			if r.FormValue("ajax") != "" {
				io.WriteString(w, "NotLoggedIn")
				return
			}

			subdir := webfw.Config().SubDir
			dbg.D(TAG, "Redirect with subdir: %v", subdir)
			http.Redirect(w, r, subdir+"/"+ctx.Value("T").(*Translater).UrlLang+"/odl/login/", 307)
			return
		}
		uTime := session.Values["LastUserUpdate"]

		if uTime == nil || time.Now().Unix()-uTime.(int64) > 60 { // Every minute refresh user from database
			UpdateUserDataForSession(r, usr.Email(), w)

			if usr == nil || err != nil || !usr.IsLoggedIn() || session == nil {
				usr, session, err = GetUserWithSession(r)
				if r.FormValue("ajax") != "" {
					io.WriteString(w, "NotLoggedIn")
					return
				}

				subdir := webfw.Config().SubDir
				dbg.D(TAG, "Redirect after UpdateUserData with subdir: %v", subdir)
				http.Redirect(w, r, subdir+"/"+ctx.Value("T").(*Translater).UrlLang+"/login", 307)
				return
			}
		}
		next.ServeHTTP(ctx, w, r)
	}

	return alice.CtxHandlerFunc(fn)
}

// GetUserWithSession Gets the requesting user together with his session
func GetUserWithSession(r *http.Request) (usr *OdlUser, session *sessions.Session, err error) {
	dbg.V(TAG, "I am at GetUserWithSession")
	defer func() {
		if errr := recover(); errr != nil {
			dbg.E(TAG, "panic in GetUserWithSession: %v for request : %v", errr, dbg.GetRequest(r))
			err = errors.New(fmt.Sprintf("%s", err))
		}
	}()
	session, err = webfw.SessionStore.Get(r, "usr_new")

	if err != nil {
		dbg.D(TAG, "Session get error - try with new")
		session, err = webfw.SessionStore.New(r, "usr_new")

		if err != nil {
			dbg.I(TAG, "Could not get session : ", err)
			return
		}
	}

	susr := session.Values["User"]

	if susr != "" && susr != nil {
		u := susr.(*OdlUser)
		usr = u
	}

	dbg.V(TAG, "Finished GetUserWithSession")
	return
}

// UpdateUserDataForSession gets new user data from the database.
// It can take up to 60 seconds for other sessions to update the data
func UpdateUserDataForSession(r *http.Request, email string, w http.ResponseWriter) {

	user, session, err := GetUserWithSession(r)

	if session == nil || user == nil || !user.IsLoggedIn() || err != nil {
		dbg.E(TAG, "Error in UpdateUserDataForSession - user : \r\n\t session : \r\n\t err : ", user, session, err)
		return
	}
	user, err = GetUserFromDb(email)
	if user != nil && err == nil {
		user.IisLoggedIn = true
	} else {
		dbg.E(TAG, "Error in UpdateUserDataForSession (2) - user : \r\n\t session : \r\n\t err : ", user, session, err)
		return
	}
	session.Values["User"] = user
	session.Values["LastUserUpdate"] = time.Now().Unix()
	session.Save(r, w)

}

// InitEncryptionHandler initialises the encryption
func InitEncryptionHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			dbg.E(TAG, "panic in InitEncryptionHandler: %v for request : %v", err, dbg.GetRequest(r))
			webfw.DirectShowError(webfw.ViewData{ErrorType: 500}, errors.New(fmt.Sprintf("%s", err)), w)
		}
	}()
	u := login.NewUser()
	ou := OdlUser{
		*u,
		"",
		LVL_USER,
		0,
		0,
		1,
		0,
	}

	nonce := ou.Nonce()
	rndd,_ := uuid.NewV4()
	rnd := rndd.String()
	loginMutex.Lock()
	// TODO: Do not use a simple count here, instead use Random string - or somehow merge it with CSRF token approach later on
	openLoginUsers[rnd] = &ou
	loginMutex.Unlock()
	createPubKeyString := true
	pub_pem := ""
	if createPubKeyString {
		// https://github.com/golang-samples/cipher/blob/master/crypto/rsa_keypair.go
		// Get der format. priv_der []byte
		pub := nonce.PrivKey.PublicKey
		pub_der, err := x509.MarshalPKIXPublicKey(&pub)

		if err != nil {
			dbg.E(TAG, "Failed to get der format for PublicKey.", err)
			return
		}

		pub_blk := pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pub_der,
		}
		pub_pem = fmt.Sprintf("%s", pem.EncodeToMemory(&pub_blk))
	}
	data := login.ReqNonce{
		Nonce: pub_pem,
		ReqId: rnd,
		Salt:  nonce.Salt,
	}
	setNonceNAndE := false
	if setNonceNAndE { // N and E belong to public key
		data.NonceN = nonce.PrivKey.N.String()
		data.NonceE = nonce.PrivKey.E
	}
	marshaled, err := json.Marshal(data)

	if err != nil {
		dbg.E(TAG, "Error while marshalling JSON in userManager.go login 1", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	w.Write(marshaled)

	// if 2 requests get handled at the exact same time, prevent them from being mixed

	go func(key string) {
		defer func() {
			if err := recover(); err != nil {
				dbg.E(TAG, "Error setting openLoginUser nil:", err)
			}
		}()
		// One key is only valid for 60 seconds
		time.Sleep(60 * time.Second)
		loginMutex.Lock()
		openLoginUsers[key] = nil
		loginMutex.Unlock()
	}(rnd)
}

// HandleLoginHandler manages the login-process. Only works with InitLoginHandler being called during the previous
// 60 seconds to provide additional encryption.
func HandleLoginHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			dbg.E(TAG, "panic in HandleLoginHandler: %v for request : %v", err, dbg.GetRequest(r))
			webfw.DirectShowError(webfw.ViewData{ErrorType: 500}, errors.New(fmt.Sprintf("%s", err)), w)
		}
	}()

	dbg.D(TAG, "I'm at HandleLoginHandler")
	T := ctx.Value("T").(*Translater)

	if r.FormValue("reqId") == "" { // Encryption was not initialised.
		dbg.WTF(TAG, "No reqId provided for login - this should not have happened", dbg.GetRequest(r))
		w.Write([]byte("Please init encryption!"))
		return
	}
	// reqId was sent - check if login is correct and create user session
	ou, err := getOpenUserFromRequest(r)
	if err != nil {
		dbg.W(TAG, "Could not get open user from request in HandleLoginHandler - if this occures too often we might are getting attacked : ", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if(r.FormValue("cookie")!="") {
		var c string
		pk := ou.Nonce().PrivKey
		for _,v := range strings.Split(r.FormValue("cookie"),"_____") {
			var k string
			k,err = getDecryptedPw(v,pk)
			if err != nil {
				dbg.E(TAG, "Error getting decrypted cookie : ", err, dbg.GetRequest(r))
				return
			}
			c += k
		}

		if err != nil {
			dbg.E(TAG, "Error getting decrypted cookie : ", err, dbg.GetRequest(r))
			return
		}
		cookie := &http.Cookie{
			Name:"usr_new",
			Value:c,
			Secure:true,
			HttpOnly:true,
		}
		http.SetCookie(w,cookie)
		return
	}

	var passwd string
	passwd, err = getDecryptedPw(r.FormValue("password"), ou.Nonce().PrivKey)
	if err != nil {
		dbg.E(TAG, "Error getting decrypted PW : ", err, dbg.GetRequest(r))
		return
	}
	answer := &LoginAnswer{Errors: make(map[string]string)}
	if ou == nil {
		dbg.W(TAG, "Somebody tried to request dead user in userManager.go login. If this occures too often we might are getting bruteforced.")
		answer.Timedout = true
	} else { // Request not timed out (key is still valid) - check login data
		usr, session, err := checkLoginAndInitSession(passwd, r.FormValue("email"), answer, r, T)

		if err != nil {
			dbg.I(TAG, "Error checking login and initing session : ", err)

		}

		if answer.Success {
			if dbg.Develop {
				dbg.D(TAG, "Logged in usr %v with session %v", usr, session)
			}
			var err error
			err = session.Save(r, w)

			if r.FormValue("FromApp") != "" {
				answer.SessionCookieVal, _ = webfw.SessionStore.GetEncoded(session) // SessionID can't be sent unencrypted and btw is of no use for the client!
			}
			if err != nil {
				dbg.E(TAG, "Error saving session : ", err)
			}
		} else {
			dbg.I(TAG, "Failed login for username ", r.FormValue("email"))
		}
	}

	marshaled, err := json.Marshal(answer)

	if err != nil {

		dbg.E(TAG, "Error while marshalling JSON in userManager.go login 2", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	w.Write(marshaled)

}

// getOpenUserFromRequest gets the currently openLoginUser from the given request.
func getOpenUserFromRequest(r *http.Request) (ou *OdlUser, err error) {
	loginMutex.Lock()
	reqId := r.FormValue("reqId")

	dbg.I(TAG, "Getting open user for reqId %v", reqId)
	ou = openLoginUsers[reqId]
	// destroy the reference to the user to be sure it can not be
	// compromised
	openLoginUsers[reqId] = nil
	loginMutex.Unlock()

	return
}

// TODO: Store sessionIDs per user in extra (redis or SQLite?)-database and implement delete of sessions with user interface (e.g. for revoking an android device)

// checkLoginAndInitSession Checks login and fills userdata.
func checkLoginAndInitSession(passwd string, mailAddr string, answer *LoginAnswer, r *http.Request, T *Translater) (usr *OdlUser, session *sessions.Session, err error) {
	// better make a completely new user to remove any possibility to hijack him
	defer func() {
		if errr := recover(); errr != nil {
			dbg.E(TAG, "panic in checkLoginAndInitSession: %v for request : %v", errr, dbg.GetRequest(r))
			err = errors.New(fmt.Sprintf("%s", err))
		}
	}()
	usr, err = GetUserFromDb(mailAddr)

	if usr == nil || usr.Email() == "" || err != nil {
		if err == nil {
			err = errors.New("No user found")
		}
		answer.Error = true
		answer.Success = false
		answer.ErrorMessage = T.T("login_wrongCredentials")
		return
	}
	// TODO: Implement timeout for too many login attempts
	/*if GetUserLockoutMinutes(usr) != 0 {
		time.Sleep(7 * time.Second)
		answer.Success = false
		answer.Error = true
		answer.ErrorMessage = T.T("accountLocked"), minutesLocked
	}*/
	err = bcrypt.CompareHashAndPassword([]byte(usr.PwHash()), []byte(passwd))
	// Put this here so an attacker has to wait for the password to be encrypted but he does not know if the password was correct if account is not activated :)
	if usr.ActivationKey() != "" {
		err = AccountNotActivated
	}
	if err == nil {
		usr.IisLoggedIn = true
		answer.Success = true

		_, session, err = GetUserWithSession(r)
		if err != nil {
			//increaseLoginCountForUser(usr)
			dbg.E(TAG, "Login unsuccesful with error : ", err)

			return
		}
		session.Values["User"] = usr
		session.Values["LastUserUpdate"] = time.Now().Unix()
		uData := GetUDataFromUser(usr)

		answer.User = uData
		dbg.D(TAG, "Login successful")

		return

	} else {

		dbg.D(TAG, "Login unsuccessful %s", err)
		answer.Error = true
		answer.Success = false

		if err == AccountNotActivated {
			answer.ErrorMessage = T.T("accountNotActivated")
			return

		}
		//increaseLoginCountForUser(usr)
		err = errors.New("Wrong password")
		answer.ErrorMessage = T.T("login_wrongCredentials")
		return
	}

}

// GetUDataFromUser converts an OdlUser to an UserData-Map.
func GetUDataFromUser(usr * OdlUser) (map[string]interface{}) {
	uData := make(map[string]interface{})
	uData["LoginName"] = usr.LoginName()
	uData["FirstName"] = usr.FirstName()
	uData["LastName"] = usr.LastName()
	uData["Title"] = usr.Title()
	uData["Id"] = usr.Id()
	uData["Level"] = usr.Level()
	uData["TutorialDisabled"] = usr.TutorialDisabled()
	uData["NextNotificationTime"] = usr.NextNotificationTime()
	uData["NotificationsEnabled"] = usr.NotificationsEnabled()
	return uData
}

// TODO : Implement this to prevent bruteforcing.
func GetUserLoginCount(usr *OdlUser, db *sql.DB) (tryCount int, lastTryTime int64) {
	return
}

// TODO : Implement this to prevent bruteforcing.
func increaseLoginCountForUser(usr *OdlUser, db *sql.DB, tryCount int, lastTryTime int64) {
	//usr.SetFailedLoginCount(usr.FailedLoginCount() + 1)

	db.Exec("UPDATE LOGINCOUNTS set failedLoginCount=? where userid=?", usr.Id)
}

// LogoutHandler logs out the requesting user.
func LogoutHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			dbg.E(TAG, "panic in LogoutHandler: %v for request : %v", err, dbg.GetRequest(r))
			webfw.DirectShowError(webfw.ViewData{ErrorType: 500}, errors.New(fmt.Sprintf("%s", err)), w)
		}
	}()

	dbg.D(TAG, "I'm at LogoutHandler")
	usr, session, _ := GetUserWithSession(r)
	if usr != nil {
		usr.IisLoggedIn = false
	}

	session.Values["User"] = ""
	session.Options.MaxAge = -1
	session.Save(r, w)

	subdir := webfw.Config().SubDir
	http.Redirect(w, r, subdir+"/"+ctx.Value("T").(*Translater).UrlLang+"/odl/login/", 307)
}

const userColumns = "id,firstName,lastName,pwhash,title,mail,activationkey,level,lowercasemail,strftime('%s', timeEnter),tutorialDisabled,notificationsEnabled,nextNotificationTime"

// GetUserFromDb gets the user with the given Mailaddress from the database.
func GetUserFromDb(mail string) (usr *OdlUser, err error) {

	db := getDbConnection()
	defer db.Close()
	row := db.QueryRow("SELECT "+userColumns+" FROM USERS where lowercasemail=?",
		strings.ToLower(mail))
	if usr, err = getOdlUserFromRow(row); err != nil {
		dbg.I(TAG, "User with mail "+mail+" not found ", err)
		return
	}
	return
}

// GetKeyByGuid gets the device-Key by its GUID.
func GetKeyByGuid(guid string) (key *Key, err error) {

	db := getDbConnection()
	defer db.Close()
	row := db.QueryRow("SELECT userId,guid,password,created FROM Keys WHERE guid=?",guid)
	if key, err = getOdlKeyFromRow(row); err != nil {
		dbg.I(TAG, "Device with guid "+guid+" not found ", err)
		return
	}
	return
}

// GetAllKeys gets all device-keys.
func GetAllKeys() (keys[]*Key , err error) {
	db := getDbConnection()
	defer db.Close()

	rows, err := db.Query("SELECT userId,guid,password,created FROM Keys")
	if err != nil {
		return
	}
	keys = make([]*Key, 0)
	for rows.Next() {
		var key *Key
		key, err = getOdlKeyFromRow(rows)
		if err != nil {
			return
		}
		keys = append(keys, key)

	}
	return
}

// UpdateKey updates the given device-key by its GUID.
func UpdateKey(key *Key) (err error) {

	db := getDbConnection()
	defer db.Close()
	prevKey,err := GetKeyByGuid(key.GUID)
	if prevKey.Password != key.Password && key.Password !=""{
		// Password changed - hash it.
		key.Password,err = GetPwHash(key.Password)
		if err != nil {
			dbg.E(TAG,"Error hashing password : ", err)
			return
		}
	}
	if prevKey.UserId>0 && key.UserId!=prevKey.UserId {
		// Can't change userId for a key that is still in use!
		return errors.New("Device already in use!")
	}

	changedFields := make([]interface{},0)
	changedFields = append(changedFields,key.GUID)
	changeStr := "guid=?"
	if int(key.UserId)==-1 {
		changedFields = append(changedFields, nil)
		changeStr += ",userId=?"
	} else if key.UserId!=0 {
		changedFields = append(changedFields, key.UserId)
		changeStr += ",userId=?"
	}

	if key.Password!="" {
		changedFields = append(changedFields, key.Password)
		changeStr += ",password=?"
	}
	changedFields = append(changedFields, key.GUID)

	_,err = db.Exec("UPDATE Keys SET "+changeStr+" WHERE guid=?",changedFields...)

	if err != nil {
		dbg.E(TAG,"Error updating KEYS UPDATE Keys SET "+changeStr+" WHERE guid=? with pms : %v : ",append(changedFields, err)...)
	}


	return
}

// createKey creates a new device-key.
func createKey(key *Key) (err error) {
	db := getDbConnection()
	defer db.Close()
	key.Created = time.Now().Unix()*1000
	pw := ""
	pw,err = GetPwHash(key.Password)
	if err != nil {
		dbg.E(TAG,"Error hashing password : ", err)
		return
	}
	_,err = db.Exec("INSERT INTO Keys (userId,guid,password,created) VALUES (?,?,?,?)",key.UserId,key.GUID,pw,key.Created)
	if err != nil {
		dbg.E(TAG,"Error creating key : ", err)
		return
	}
	return
}

// DeleteKey deletes the given device-key by its GUID.
func DeleteKey(key *Key) (err error) {
	db := getDbConnection()
	defer db.Close()
	_,err = db.Exec("DELETE FROM Keys WHERE guid=?",key.GUID)
	if err != nil {
		dbg.E(TAG,"Error deleting key : ", err)
		return
	}
	return
}

// GenerateKeys genereates the given amount of device-keys.
func GenerateKeys(cnt int64) (keys []*Key, err error) {
	keys = make([]*Key,0)
	for i:=int64(0);i<cnt;i++ {
		var key Key
		key, err = GenerateKey()
		if err != nil {
			dbg.E(TAG,"Error generating key :" , err)
			keys = make([]*Key,0)
			return
		}
		keys = append(keys,&key)
	}
	return
}

// GenerateKey generates a new device-key.
func GenerateKey() (key Key, err error) {
	u, err := uuid.NewV4()
	if err != nil {
		dbg.E(TAG,"Error generating key :" , err)
		return
	}
	key.GUID = u.String()
	key.Password = tools.RandSeq(256)
	err = createKey(&key)
	if err != nil {
		dbg.E(TAG,"Error creating key : %+v ",key, err)
		key = Key{}
		return
	}
	return
}

// getOdlKeyFromRow scans a row into a Key-object.
func getOdlKeyFromRow(row Scannable) (key *Key, err error) {
	key = &Key{}
	err = row.Scan(&key.UserId,&key.GUID,&key.Password,&key.Created)
	return
}

// GetUserById gets an OdlUser by its ID.
func GetUserById(id int64) (usr *OdlUser, err error) {

	db := getDbConnection()
	defer db.Close()
	row := db.QueryRow("SELECT "+userColumns+" FROM USERS where id=?",
		id)
	if usr, err = getOdlUserFromRow(row); err != nil {
		dbg.I(TAG, "User with id "+strconv.FormatInt(id, 64)+" not found ", err)
		return
	}
	return
}

// Scannable defines an interface for scannable objects, mainly SQL-rows.
type Scannable interface {
	Scan(...interface{}) error
}

// getOdlUserFromRow scans a row and fills the ODLUser-object.
func getOdlUserFromRow(row Scannable) (usr *OdlUser, err error) {
	var fName, lName, pwH, title, email, activationkey, lowercasemail string

	var id, timeEnter,nextNotificationTime int64
	var _timeEnter sql.NullInt64
	var level, tutorialDisabled,notificationsEnabled int
	if err = row.Scan(&id, &fName, &lName, &pwH, &title, &email,
		&activationkey, &level, &lowercasemail, &_timeEnter, &tutorialDisabled,
		&notificationsEnabled,
		&nextNotificationTime); err != nil {
		return
	}
	if _timeEnter.Valid {
		timeEnter = _timeEnter.Int64
	}
	usr = &OdlUser{}
	usr.SetId(id)
	usr.SetFirstName(fName)
	usr.SetLastName(lName)
	usr.SetLevel(level)
	usr.SetPwHash(pwH)
	usr.SetTitle(title)
	usr.SetEmail(email)
	usr.SetLoginName(lowercasemail)
	usr.SetActivationKey(activationkey)
	usr.SetTimeEnter(timeEnter)
	usr.SetTutorialDisabled(tutorialDisabled)
	usr.SetNotificationsEnabled(notificationsEnabled)
	usr.SetNextNotificationTime(nextNotificationTime)

	return
}

// GetAllUsers returns all users as AdminEditableUserData
func GetAllUsers() (usrs []*AdminEditableUserData, err error) {
	db := getDbConnection()
	defer db.Close()
	rows, err := db.Query("SELECT " + userColumns + " FROM USERS")
	if err != nil {
		return
	}
	usrs = make([]*AdminEditableUserData, 0)

	for rows.Next() {
		var usr *OdlUser
		usr, err = getOdlUserFromRow(rows)
		u := &AdminEditableUserData{
			Id:               usr.Id(),
			Title:            usr.Title(),
			FirstName:        usr.FirstName(),
			LastName:         usr.LastName(),
			LoginName:        usr.LoginName(),
			Password:         "",
			RepeatedPassword: "",
			Level:            usr.Level(),
			ActivationKey:    usr.ActivationKey(),
		}
		usrs = append(usrs, u)
		if err != nil {
			return
		}
	}

	return
}

// DeleteUserFromDb deletes the given users from the database.
func DeleteUserFromDb(mail string) (err error) {
	//TODO: Implement something to also delete user-db (maybe delayed?)
	db := getDbConnection()
	defer db.Close()
	var _, db_err = db.Exec("DELETE FROM USERS where lowercasemail=?",
		strings.ToLower(mail))
	return db_err
}

// DeleteUser deletes the user by the given ID.
func DeleteUser(id int64) (err error) {
	usr, err := GetUserById(id)
	if err != nil {
		return
	}
	return DeleteUserFromDb(usr.Email())
}

// CreateUser creates a new user.
func CreateUser(r *http.Request, user *UserEditableUserData, ou *OdlUser, inviteKey string, T *Translater) (id int64, err error, _errors map[string]string) {
	defer func() {
		if errr := recover(); errr != nil {
			dbg.E(TAG, "panic in CreateUser: %v for request : %v", err, dbg.GetRequest(r))
			err = errors.New(fmt.Sprint(errr))
		}
	}()

	dbg.D(TAG, "I'm at CreateUser")

	invited := true
	if InviteRequired {
		invited = false
		if inviteKey != "" {
			invited, err = CheckInviteKey(inviteKey)
			if err != nil {
				err = errors.New("inviteOnly")
				return
			}
		}
	}
	if !invited {
		err = errors.New("inviteOnly")
		return
	}
	success := false
	success, _errors = verifyRegisterFields(user)
	if !success {
		err = errors.New("Verification failed")
		return
	}

	ou.SetLevel(LVL_USER)
	ou.SetActivationKey(tools.RandSeq(32))
	ou.SetTitle(user.Title)
	var pw string
	pw = user.Password
	if err != nil {
		dbg.E(TAG, "Error decrypting password", err)
		return
	}
	pw, err = GetPwHash(pw)
	if err != nil {
		dbg.E(TAG, "Error hashing password", err)
		return
	}
	ou.SetPwHash(pw)
	ou.SetLoginName(user.LoginName)
	ou.SetEmail(user.LoginName)
	ou.SetLastName(user.LastName)
	ou.SetFirstName(user.FirstName)
	ou.SetTimeEnter(int64(time.Now().Unix()))

	err = CreateNewUser(ou)
	id = ou.Id()
	if err != nil {
		return
	} else {
		err = SendActivationMail(T, ou)
		if err != nil {
			dbg.E(TAG, "Error sending activation mail : ", err)
			return
		} else { // Complete activation succesful - remove invitekey
			if InviteRequired {
				err = RemoveInviteKey(inviteKey)
				if err != nil {
					dbg.E(TAG, "Error removing invite key : ", err)
				}
			}
		}
	}

	return

}

var nilUser = errors.New("Nil user")

// GetPwHash hashs the given password, using bcrypt
func GetPwHash(passwd string) (npasswd string, err error) {
	start := time.Now()

	// hash passwd
	passwdByte, err := bcrypt.GenerateFromPassword([]byte(passwd), 13)

	if err != nil {
		dbg.E(TAG, " bcrypting failed :( : ", err)
		return
	}

	npasswd = string(passwdByte)

	elapsed := time.Since(start)
	dbg.I(TAG, "Pw encrypt took %s", elapsed)

	return
}

var registerTemplate *template.Template
var editUserTemplate *template.Template
var resetPasswordTemplate *template.Template

// SendActivationMail sends the mail for activating a user-account.
func SendActivationMail(T *Translater, ou *OdlUser) (err error) {
	defer func() {
		if errr := recover(); errr != nil {
			ou.SetPwHash("Hidden")
			dbg.E(TAG, "panic in SendActivationMail: %v for user : %v", errr, ou)
			err = errors.New(fmt.Sprintf("%s", errr))
		}
	}()
	if registerTemplate == nil {
		txt, err := ioutil.ReadFile(webfw.Config().RootDir + "/views/registerMail.htm")
		if err != nil {
			dbg.E(TAG, "Error reading registerMail template : ", err)
			return err
		}
		registerTemplate = template.Must(template.New("registerMailTemplate").Delims("{[{", "}]}").Parse(string(txt)))
	}

	lnk := webfw.Config().WebUrl + "/" + T.UrlLang + "/register?mail=" + ou.Email() + "&activateKey=" + ou.ActivationKey()
	anrede := GetAnrede(ou, T)

	mdl := models.RegisterMailModel{
		E: &models.RegisterMailEnhance{Name: anrede,
			ActivationLink: lnk,
			T:              T,
		},
	}

	buffer := new(bytes.Buffer)
	err = registerTemplate.Execute(buffer, &mdl)
	if err != nil {
		dbg.E(TAG, "Error filling registerMail template : ", err)
		return err
	}
	err = tools.SendODLMail([]string{ou.Email()}, T.T("registerMail_subject"), string(buffer.Bytes()), false)

	return
}

// SendResetMail sends an mail to reset the users password.
func SendResetMail(email string, resetKey string, T *Translater) (err error) {
	if resetPasswordTemplate == nil {
		txt, err := ioutil.ReadFile(webfw.Config().RootDir + "/views/resetPasswordMail.htm")
		if err != nil {
			dbg.E(TAG, "Error reading resetPasswordMail template : ", err)
			return err
		}
		resetPasswordTemplate = template.Must(template.New("resetPasswordTemplate").Delims("{[{", "}]}").Parse(string(txt)))
	}
	ou, err := GetUserFromDb(email)
	if err != nil {
		return err
	}
	lnk := webfw.Config().WebUrl + "/" + T.UrlLang + "/odl/newPassword/" + email + "/" + resetKey
	anrede := GetAnrede(ou, T)

	mdl := models.ResetPwMailModel{
		E: &models.ResetPwMailEnhance{
			Name:        anrede,
			ResetPwLink: lnk,
			T:           T,
		},
	}

	buffer := new(bytes.Buffer)
	err = resetPasswordTemplate.Execute(buffer, &mdl)
	if err != nil {
		dbg.E(TAG, "Error filling resetPasswordMail template : ", err)
		return err
	}
	err = tools.SendODLMail([]string{email}, T.T("resetPwMail_subject"), string(buffer.Bytes()), false)
	return
}

// GetAnrede gets the Anrede, e.g. "Herr Sonntag" and the Sexswitch vor formatted text (e.g. "Sehr geehrter" in E-Mails)
func GetAnrede(ou *OdlUser, T *Translater) (anrede string) {
	anrede = T.T(ou.Title()) + " " + ou.FirstName() + " " + ou.LastName()

	switch ou.Title() {
	case "Mr.":
		anrede = T.T("Mail_Start_male") + anrede
	case "Ms.":
		anrede = T.T("Mail_Start_female") + anrede
	case "Dr. (female)":
		anrede = T.T("Mail_Start_female") + anrede
	case "Dr. (male)":
		anrede = T.T("Mail_Start_male") + anrede
	case "Prof. (female)":
		anrede = T.T("Mail_Start_female") + anrede
	case "Prof. (male)":
		anrede = T.T("Mail_Start_male") + anrede
	default:
		anrede = T.T("Mail_Start_neutral") + anrede
	}

	return
}

// ActivateUser activates the user if the activation-key is correct, else returns WrongActivationKey-error.
func ActivateUser(mailAddress string, key string) (err error) {
	ou, err := GetUserFromDb(mailAddress)
	if err != nil {
		return err
	}
	actKey := ou.ActivationKey()
	if actKey == "" { // Account already activated
		return
	}
	if actKey == key { // correct key - activate account
		db := getDbConnection()
		defer db.Close()
		_, err = db.Exec("UPDATE USERS SET activationkey=\"\" WHERE id=?", ou.Id())
		return
	}

	return WrongActivationKey
}

// GetPasswordResetKey returns the current passwordResetKey for the given email. If no key is existing,
// a new one will be created.
func GetPasswordResetKey(email string) (key string, err error) {
	usr, err := GetUserFromDb(email)
	if err != nil {
		return
	}
	db := getDbConnection()
	defer db.Close()

	var validity int64
	err = db.QueryRow("SELECT resetkey,validuntil FROM PWRESETS WHERE userid=?", usr.Id()).Scan(&key, &validity)
	if err != sql.ErrNoRows { // We already have a reset key in the database
		if err != nil {
			dbg.E(TAG, "Error asking for resetkey : ", err)
			return
		}

		if validity > time.Now().Unix() {
			err = nil
			return
		} else {
			db.Exec("DELETE FROM PWRESETS WHERE userid=?", usr.Id())
		}
	}
	key = tools.RandSeq(256)
	dbg.D(TAG, "ResetKey generated : ", key)
	// key is valid for 3 hours
	_, err = db.Exec("INSERT INTO PWRESETS (userid,resetkey,validuntil,lowercasemail) VALUES (?,?,?,?)", usr.Id(), key, time.Now().Unix()+60*60*3, usr.LoginName())

	return
}

// ChangePassword changes the users password, if the resetKey is correct.
func ChangePassword(email string, newPassword string, resetKey string) bool {

	valid, uid := CheckResetKeyValidity(resetKey, email)

	if !valid || uid == 0 {
		return false
	}
	db := getDbConnection()
	defer db.Close()
	if newPassword == "" {
		return false
	}
	start := time.Now()

	// hash passwd
	passwdByte, err := bcrypt.GenerateFromPassword([]byte(newPassword), 13)

	if err != nil {
		dbg.E(TAG, " bcrypting failed :( : ", err)
		return false
	}

	pwHash := string(passwdByte)

	elapsed := time.Since(start)
	dbg.I(TAG, "Pw encrypt took %s", elapsed)

	res, err := db.Exec("UPDATE USERS SET pwHash=? WHERE lowercasemail=? and id=?", pwHash, strings.ToLower(email), uid)

	if err != nil {
		dbg.E(TAG, "error updating USERS with new pwhash : ", err)
		return false
	}
	if aff, _ := res.RowsAffected(); aff == 0 {

		return false
	}

	db.Exec("DELETE FROM PWRESETS WHERE userid=?", uid)

	return true

}

// CheckResetKeyValidity checks if the given resetKey is valid for the given Mailaddress and returns if it is valid
// and to which userId it belongs.
func CheckResetKeyValidity(resetKey string, email string) (valid bool, userid int64) {
	db := getDbConnection()
	defer db.Close()
	var validity int64
	var key string

	err := db.QueryRow("SELECT ID FROM USERS WHERE lowercasemail=?", strings.ToLower(email)).Scan(&userid)
	// The current users mailAddress has to be the same as the one when the key was generated
	err = db.QueryRow("SELECT resetkey,validuntil FROM PWRESETS WHERE userid=? and lowercasemail=?", userid, strings.ToLower(email)).Scan(&key, &validity)
	if err == sql.ErrNoRows { // No resetkey found

		return false, -1
	} else if err != nil {
		dbg.E(TAG, "Error querying PWRESETS in CheckResetKeyValidity : ", err)
		return false, -1
	}
	if validity < time.Now().Unix() || key != resetKey {
		return false, -1
	}

	return true, userid
}

// mailRegExp is a regular expression to check for the validity of an E-Mail-Address.
const mailRegExp = "(?i)^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,4}$"

var compiledMailRegExp *regexp.Regexp

// verifyRegisterFields check if an valid mail was provided and if the password is long enough and both passwords
// are the same.
func verifyRegisterFields(usr *UserEditableUserData) (success bool, _errors map[string]string) {
	success = true
	_errors = make(map[string]string)
	if succ, errorMsg := VerifyMail(usr.LoginName); !succ {
		_errors["mail"] = errorMsg
		success = false
	}

	if succ, errorMsg := VerifyPassword(usr.Password, usr.RepeatedPassword); !succ {
		_errors["password"] = errorMsg
		success = false
	}

	return
}

// VerifyPassword checks if a password is long enough and both passwords match.
func VerifyPassword(pw string, pwRepeat string) (success bool, errorMessage string) {
	success = true

	if pw != pwRepeat {
		errorMessage = "error_passwordsNotMatch"
		success = false
		return
	}
	if len(pw) < 8 {
		errorMessage = "error_insecurePassword"
		success = false
		return
	}
	return
}

// VerifyMail verifies if it is a valid E-Mail-Address.
func VerifyMail(mail string) (success bool, errorMessage string) {
	success = true
	if compiledMailRegExp == nil {
		compiledMailRegExp, _ = regexp.Compile(mailRegExp)
	}
	if !compiledMailRegExp.MatchString(mail) {
		errorMessage = "error_invalidEmail"
		success = false
	}

	return
}

// UpdateUser updates a user - be aware that this function does check password security and mail validity!
// also sends a notification-Mail about the changes, if sendNotificationMail is true.
func UpdateUser(oldUser *OdlUser, updatedUser *AdminEditableUserData, r *http.Request, T *Translater, canChangeAdminFields bool, sendNotificationMail bool, updateTutorialInfoTable bool) (rowCount int64, err error) {
	if oldUser.Id() != updatedUser.Id {
		dbg.WTF(TAG, "This should not have happened - we tried to update an old user with an different Id than the new user")
		err = errors.New("internal server error")
		return
	}
	isAdmin := false
	hasT := T != nil
	if r != nil {
		u, _, err := GetUserWithSession(r)
		if err != nil {
			dbg.E(TAG, "Error getting user with session", err)
			return -1, err
		}
		isAdmin = u.Level() == 1337
	}

	if (oldUser.Id() != updatedUser.Id || (updatedUser.Level != oldUser.Level() && updatedUser.Level != 0) || updatedUser.ActivationKey != oldUser.ActivationKey()) && !isAdmin {
		dbg.WTF(TAG, "Somebody tried to edit userdata without having permissions - old : %+v \r\n new : %+v", oldUser, updatedUser)
		err = Error_AccessDenied
		return
	}
	udFields := make([]string, 0)
	udParams := make([]interface{}, 0)
	changes := make([]string, 0)
	nameChanged := false
	tutInfoChanged := false
	if updatedUser.Password != "" {
		if succ, errMsg := VerifyPassword(updatedUser.Password, updatedUser.RepeatedPassword); !succ {
			err = errors.New(errMsg)
			return
		}
	}
	if updatedUser.LoginName != "" && updatedUser.LoginName != oldUser.Email() {
		if succ, errMsg := VerifyMail(updatedUser.LoginName); !succ {
			err = errors.New(errMsg)
			return
		}
	}
	if updatedUser.Password != "" {
		errrr := bcrypt.CompareHashAndPassword([]byte(oldUser.PwHash()), []byte(updatedUser.Password))
		if errrr != nil { // new password
			updatedUser.Password, err = GetPwHash(updatedUser.Password)
			if err != nil {
				dbg.E(TAG, "Error hashing password", err)
				return
			}
			udFields = append(udFields, "pwhash")
			udParams = append(udParams, updatedUser.Password)
			if hasT {
				changes = append(changes, T.T("settings_passwordChanged"))
			}
		}
	}

	if updatedUser.LoginName != "" && updatedUser.LoginName != oldUser.Email() {

		udFields = append(udFields, "mail")
		udParams = append(udParams, updatedUser.LoginName)
		udFields = append(udFields, "lowercasemail")
		udParams = append(udParams, strings.ToLower(updatedUser.LoginName))
		if hasT {
			changes = append(changes, fmt.Sprintf(T.T("settings_emailChanged")+"%v", updatedUser.LoginName))
		}
	} else {
		updatedUser.LoginName = oldUser.Email()
	}
	if updatedUser.FirstName != "" && updatedUser.FirstName != oldUser.FirstName() {
		udFields = append(udFields, "firstName")
		udParams = append(udParams, updatedUser.FirstName)
		if hasT {
			changes = append(changes, T.T("settings_firstNameChanged"))
		}
		nameChanged = true
	} else {
		updatedUser.FirstName = oldUser.FirstName()
	}
	if updatedUser.LastName != "" && updatedUser.LastName != oldUser.LastName() {
		udFields = append(udFields, "lastName")
		udParams = append(udParams, updatedUser.LastName)
		if hasT {
			changes = append(changes, T.T("settings_lastNameChanged"))
		}
		nameChanged = true
	} else {
		updatedUser.LastName = oldUser.LastName()
	}
	if updatedUser.Title != "" && updatedUser.Title != "user" && updatedUser.Title != oldUser.Title() {
		udFields = append(udFields, "title")
		udParams = append(udParams, updatedUser.Title)
		if hasT {
			changes = append(changes, T.T("settings_titleChanged"))
		}
		nameChanged = true
	} else {
		updatedUser.Title = oldUser.Title()
	}

	if updatedUser.TutorialDisabled != 0 && updatedUser.TutorialDisabled != oldUser.TutorialDisabled() {
		udFields = append(udFields, "tutorialDisabled")
		udParams = append(udParams, updatedUser.TutorialDisabled)
		if hasT {
			changes = append(changes, T.T("settings_tutorialDisabledChanged"))
		}
		tutInfoChanged = true
	} else {
		updatedUser.TutorialDisabled = oldUser.TutorialDisabled()
	}
	if updatedUser.NotificationsEnabled != 0 && updatedUser.NotificationsEnabled != oldUser.NotificationsEnabled() {
		udFields = append(udFields, "notificationsEnabled")
		udParams = append(udParams, updatedUser.NotificationsEnabled)
		if hasT {
			changes = append(changes, T.T("settings_notificationsEnabledChanged"))
		}
	} else {
		updatedUser.NotificationsEnabled = oldUser.NotificationsEnabled()
	}

	if updatedUser.NextNotificationTime != 0 && updatedUser.NextNotificationTime != oldUser.NextNotificationTime() {
		udFields = append(udFields, "nextNotificationTime")
		udParams = append(udParams, updatedUser.NextNotificationTime)

	} else {
		updatedUser.NextNotificationTime = oldUser.NextNotificationTime()
	}

	if canChangeAdminFields {
		if updatedUser.Level != 0 && updatedUser.Level != oldUser.Level() {
			udFields = append(udFields, "Level")
			udParams = append(udParams, updatedUser.Level)
			if hasT {
				changes = append(changes, T.T("settings_LevelChanged"))
			}
			nameChanged = true
		} else {
			updatedUser.Level = oldUser.Level()
		}
		if updatedUser.ActivationKey != "" && updatedUser.ActivationKey != oldUser.ActivationKey() {
			if updatedUser.ActivationKey == "-" {
				updatedUser.ActivationKey = ""
			}
			udFields = append(udFields, "ActivationKey")
			udParams = append(udParams, updatedUser.ActivationKey)
			if hasT {
				changes = append(changes, T.T("settings_ActivationKeyChanged"))
			}
			nameChanged = true
		} else {
			updatedUser.ActivationKey = oldUser.ActivationKey()
		}
	}

	fieldList := ""
	first := true
	for _, v := range udFields {
		if first {
			first = false
		} else {
			fieldList += ","
		}
		fieldList += v + "=?"
	}

	if first {
		dbg.I(TAG, "No fields changed")
		return
	}
	updQry := "UPDATE USERS SET " + fieldList + " WHERE ID=?"

	db := getDbConnection()
	defer db.Close()
	udParams = append(udParams, oldUser.Id())
	res, err := db.Exec(updQry, udParams...)
	if err != nil {
		dbg.E(TAG, "I failed to update users with query : ", updQry)
		return
	}

	if err != nil {
		return
	}
	rowCount, err = res.RowsAffected()
	if err != nil {
		return
	}
	if len(changes) == 0 && hasT { // no fields changed
		return
	}
	if rowCount == 0 {
		dbg.WTF(TAG, "Someone tried to edit a non-existent user - wtf?")
		err = errors.New("Internal server error")
		return
	}

	if tutInfoChanged {
		var dbCon *sql.DB
		dbCon, err = GetLocationDb(updatedUser.Id)
		if err != nil {
			dbg.E(TAG, "Error getting locationDb : ", err)
			return
		}
		_, err := dbCon.Exec("Update TutorialInfo SET disabled=?", updatedUser.TutorialDisabled)
		if err != nil {
			dbg.E(TAG, "Error updating tutorialInfo : ", err)
		}
	}
	if sendNotificationMail {

		var newUser *OdlUser
		newUser, err = GetUserById(updatedUser.Id)
		if err != nil {
			dbg.E(TAG, "Error getting updated user", err)
			return
		}
		curAnrede := GetAnrede(newUser, T)

		mailData := models.EditUserMailModel{
			E: &models.EditUserMailEnhance{
				Name:    curAnrede,
				Changes: changes,
				T:       T,
			},
		}

		if nameChanged {
			prevAnrede := GetAnrede(oldUser, T)
			mailData.E.PreviousName = prevAnrede
		}
		if len(changes) == 0 {
			// nothing to update
			err = errors.New("NoChangesFound")
			return
		}
		buffer := new(bytes.Buffer)
		if editUserTemplate == nil {
			var txt []byte
			txt, err = ioutil.ReadFile(webfw.Config().RootDir + "/views/editUserMail.htm")
			if err != nil {
				dbg.E(TAG, "Error reading editUseMmail template : ", err)
				return
			}
			editUserTemplate = template.Must(template.New("editUserTemplate").Delims("{[{", "}]}").Parse(string(txt)))
		}
		dbg.D(TAG, "Filling edit user mail template")
		err = editUserTemplate.Execute(buffer, &mailData)
		if err != nil {
			dbg.E(TAG, "Error filling editUserMail template : ", err)
			return
		}
		dbg.D(TAG, "Filled user mail template")
		mailAddr := []string{oldUser.Email()}
		if newUser.Email() != oldUser.Email() && newUser.Email() != "" {
			mailAddr = append(mailAddr, newUser.Email())
		}
		err = tools.SendODLMail(mailAddr, T.T("editUserMail_Subject"), string(buffer.Bytes()), false)
		if err != nil {
			return
		}
	}
	return
}

// GetDecryptedPw takes a password and tries to decrypt it by the given request data.
// InitEncryptionHandler needs to have been called in a previous request for this to work.
func GetDecryptedPw(passwd string, r *http.Request) (pw string, err error) {
	dcuser, err := getOpenUserFromRequest(r)
	dckey := dcuser.Nonce().PrivKey
	if err != nil {
		dbg.E(TAG, "Error getting open user from request", err)
		return
	}
	pw, err = getDecryptedPw(passwd, dckey)
	if err != nil {
		dbg.E(TAG, "Error decrypting password : ", err)
		return
	}
	return
}

// getDecryptedPw decrypts the given password with the given private key.
func getDecryptedPw(passwd string, key *rsa.PrivateKey) (npasswd string, err error) {

	// decrypt passwd https://medium.com/@tikiatua/symmetric-and-asymmetric-encryption-with-javascript-and-go-240043e56daf
	cipheredValue, err := base64.StdEncoding.DecodeString(passwd)
	if err != nil {
		dbg.E(TAG, "error: decoding string (rsa)", err)
		return
	}
	//b, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, usr.Nonce().PrivKey, []byte(passwd), nil)
	b, err := rsa.DecryptPKCS1v15(rand.Reader, key, cipheredValue)

	if err != nil {
		dbg.E(TAG, "Error decrypting user data :( ", err)
		return

	}
	npasswd = string(b)
	return
}

// getDbConnection gets a connection to the userDb.db
func getDbConnection() (con *sql.DB) {
	dbPath := webfw.Config().SharedDir + "/userDb.db"
	dbg.I(TAG, "UserDBPath : %s", dbPath)
	if !userDbExists {

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			CreateNewUserDb(dbPath)
		}

		userDbExists = true
	}

	return openDbCon(dbPath)
}

// GetUserWorkDir gets the private directory for the given user.
func GetUserWorkDir(usrId int64) string {
	return webfw.Config().SharedDir + "/upload/" + fmt.Sprintf("%d", usrId)
}

// openDbCon opens a SQLITE-connection to the given path.
func openDbCon(dbPath string) *sql.DB {
	database, err := sql.Open("SQLITE", dbPath)//+"?Pooling=true")

	if err != nil {
		dbg.E(TAG, "Failed to create DB handle at openDbCon()")
		return nil
	}
	if err2 := database.Ping(); err2 != nil {
		dbg.E(TAG, "Failed to keep connection alive at openDbCon()")
		return nil
	}
	var mode string

	err = database.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		dbg.E(TAG, "Error getting journal_mode!", err)
	} else if mode != "wal" {
		dbg.W(TAG, "Setting journal_mode for %s to WAL!", dbPath)
		_, err = database.Exec("PRAGMA journal_mode=WAL")
		if err != nil {
			dbg.E(TAG, "Error setting journal_mode!", err)
		}
	}
	dbg.D(TAG, "Databaseconnection established", dbPath)
	return database
}

// CreateNewUser creates a new ODLUser, without checking anything. In normal use cases please use CreateUser!
func CreateNewUser(usr *OdlUser) (err error) {
	db := getDbConnection()
	defer db.Close()
	res, err := db.Exec("INSERT INTO USERS (firstName,lastName,pwhash,title,mail,lowercasemail,activationkey,level) VALUES(?,?,?,?,?,?,?,?)",
		usr.FirstName(), usr.LastName(),
		usr.PwHash(), usr.Title(), usr.Email(), strings.ToLower(usr.Email()), usr.ActivationKey(), usr.Level())
	if err != nil {

		if strings.Contains(err.Error(), "UNIQUE constraint") {
			dbg.I(TAG, "User for mail %s already exists", usr.Email())
			return
		}
		dbg.E(TAG, "Error creating user : ", err)
		return
	}
	id, _ := res.LastInsertId()
	usr.SetId(id)
	savePath := webfw.Config().SharedDir + "/upload/" + fmt.Sprintf("%d", id) + "/"
	os.MkdirAll(savePath, 0755)

	if usr.TutorialDisabled() == 1 { // We need to disable tutorial here because now we have user-specific db
		usr.SetTutorialDisabled(0)
		upd := &AdminEditableUserData{
			TutorialDisabled: 1,
			Id:               id,
		}
		_, err = UpdateUser(usr, upd, nil, nil, false, false, true)
		if err != nil {
			dbg.E(TAG, "Error disabling tutorial on user creation : ", err)
			err = nil
		}
	}
	return
}

// CreateNewUserDb creates a new userDb.db
func CreateNewUserDb(dbPath string) (err error) {
	db := openDbCon(dbPath)
	dbg.I(TAG, "Creating new User-DB at ", dbPath)
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE USERS(
id INTEGER PRIMARY KEY AUTOINCREMENT,
firstName TEXT NOT NULL,
lastName TEXT NOT NULL,
pwhash TEXT NOT NULL,
title TEXT NOT NULL,
mail TEXT NOT NULL,
lowercasemail TEXT NOT NULL,
activationkey STRING,
level INTEGER DEFAULT 0,
timeEnter        DATE,
 CONSTRAINT uc_Mail UNIQUE(lowercasemail)
)`)

	if err == nil {
		_, err = db.Exec("CREATE INDEX IDX_MAIL_USR ON USERS(lowercasemail)")
	}
	if err != nil {
		dbg.E(TAG, "Error creating User-DB. That is bad : ", err)

		return
	}

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS INVITES(
		id INTEGER PRIMARY KEY,
		invitekey STRING NOT NULL
		)`); err != nil {
		dbg.E(TAG, "Error creating Invites-DB. That is bad : ", err)
		return
	}

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS LOGINCOUNTS(
		userid INTEGER PRIMARY KEY,
		timeInterval INTEGER,
		failedLoginCount INTEGER
		)`); err != nil {
		dbg.E(TAG, "Error creating LoginCounts-DB. That is bad : ", err)
		return
	}
	if _, err = db.Exec("CREATE INDEX IF NOT EXISTS IDX_LoginCounts_USR_Interval ON LOGINCOUNTS(userid,timeInterval)"); err != nil {
		dbg.E(TAG, "Error creating LoginCounts-Index. That is bad : ", err)
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS PWRESETS(
		id INTEGER PRIMARY KEY,
		userid INTEGER,
		resetkey STRING NOT NULL,
		validuntil INTEGER,
		lowercasemail string
		)`); err != nil {
		dbg.E(TAG, "Error creating PWResets-DB. That is bad : ", err)

		return
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS USERS_History(
		id INTEGER PRIMARY KEY,
		useridOLD INTEGER,
		useridNEW INTEGER,
		firstNameOLD,
		firstNameNEW,
		lastNameOLD,
		lastNameNEW,
		titleOLD,
		titleNEW,
		mailOLD,
		mailNEW,
		lowercasemailOLD,
		lowercasemailNEW,
		levelOLD,
		levelNEW,
		sqlAction VARCHAR(15),
        usertimeEnter    DATE,
        usertimeUpdate   DATE,
        timeEnter        DATE)`)
	if err != nil {
		return
	}
	_, err = db.Exec(`
--  Create an update trigger
CREATE TRIGGER update_userhistory AFTER UPDATE  ON USERS
BEGIN

  INSERT INTO USERS_History  (useridNEW,useridOLD,firstNameOLD,firstNameNEW,lastNameOLD,
                        lastNameNEW,titleOLD,titleNEW,mailOLD,mailNEW,
                        lowercasemailOLD,lowercasemailNEW,levelOLD,levelNEW,
                        sqlAction,usertimeEnter,
                        usertimeUpdate,timeEnter)

          values (old.id,new.id,old.firstName,new.firstName,old.lastName,
                  new.lastName,old.title, new.title,old.mail,
                  new.mail,old.lowercasemail,new.lowercasemail,old.level,new.level, 'UPDATE',old.timeEnter,
                  DATETIME('NOW'),DATETIME('NOW') );

END;
--
--  Also create an insert trigger
--    NOTE  AFTER keyword ------v
CREATE TRIGGER insert_userhistory AFTER INSERT ON USERS
BEGIN
INSERT INTO USERS_History  (useridNEW,firstNameNEW,lastNameNEW,titleNEW,mailNEW,lowercasemailNEW,levelNEW,
                      sqlAction,usertimeEnter,timeEnter)

          values (new.id,new.firstName,new.lastName,new.title,new.mail,new.lowercasemail,new.level,
                  'INSERT',new.timeEnter,DATETIME('NOW') );

END;

--  Also create a DELETE trigger
CREATE TRIGGER delete_userhistory DELETE ON USERS
BEGIN

INSERT INTO USERS_History  (useridOLD,firstNameOLD,lastNameOLD,titleOLD,mailOLD,lowercasemailOLD,levelOLD,
                      sqlAction,timeEnter)

          values (old.id,old.firstName,old.lastName,old.title,old.mail,old.lowercasemail,old.level,
                  'DELETE',DATETIME('NOW') );

END;
		`)
	if err != nil {
		return
	}
	if _, err = db.Exec(`
CREATE TRIGGER insert_user_timeEnter AFTER  INSERT ON USERS
BEGIN

UPDATE USERS SET timeEnter = DATETIME('NOW')
         WHERE rowid = new.rowid;
END;
			`); err != nil {
		return
	}
	if _, err = db.Exec(`
CREATE TABLE  IF NOT EXISTS FailedLoginLogs(
		id INTEGER PRIMARY KEY,
		userid INTEGER,
		timeStamp INTEGER)
			`); err != nil {
		return
	}
	version := 6
	db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version))
	return
}

// CheckInviteKey checks if the given invite-key is valid.
func CheckInviteKey(key string) (exists bool, err error) {
	db := getDbConnection()
	defer db.Close()
	err = db.QueryRow("SELECT invitekey FROM INVITES WHERE invitekey=?", key).Scan(&key)
	if err != nil {
		if err == sql.ErrNoRows { // No error while querying, but row does not exist
			err = nil
			return
		}
	} else {
		exists = true
	}
	return
}

// RemoveInviteKey removes/disables the given InviteKey.
func RemoveInviteKey(key string) (err error) {
	db := getDbConnection()
	defer db.Close()
	_, err = db.Exec("DELETE FROM INVITES WHERE invitekey = ?", key)
	if err != nil {
		return err
	}
	return
}

// CreateNewInviteKey generates a new random InviteKey.
func CreateNewInviteKey() (key string, err error) {
	db := getDbConnection()
	defer db.Close()
	key = tools.RandSeq(32)
	_, err = db.Exec("INSERT INTO INVITES (invitekey) VALUES (?)", key)
	if err != nil {
		return "", err
	}
	return
}

// UpdateDbsIfNeeded updates the userDb.db if it is necessary.
func UpdateDbsIfNeeded(dbPath string) (err error) {
	db := openDbCon(dbPath)
	if db == nil {
		err = CreateNewUserDb(dbPath)
		db = openDbCon(dbPath)
	}
	if db == nil {
		dbg.E(TAG, "UpdateDbsIfNeeded could not get a user db, but I tried it so hard :( ", err)
		return
	}
	defer db.Close()
	if err != nil {
		dbg.E(TAG, "UpdateDbsIfNeeded could not get a user db, but I tried it so hard :( With error :", err)
		return
	}

	_, err = db.Exec("Select id from USERS WHERE 1=0")
	if err != nil {
		err = CreateNewUserDb(dbPath)
		if err != nil {
			dbg.E(TAG, "UpdateDbsIfNeeded could not create a new user db, but I tried it so hard :( With error :", err)
			return
		}
	}
	var version int64

	db.QueryRow("PRAGMA user_version").Scan(&version)

	if version < 1 {
		if _, err = db.Exec("ALTER TABLE USERS ADD COLUMN activationkey STRING"); err != nil {
			return
		}
		version = 1
		if _, err = db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version)); err != nil {
			return
		}
	}
	if version < 2 {
		if _, err = db.Exec("UPDATE USERS set activationkey=\"\""); err != nil {
			return
		}
		version = 2
		if _, err = db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version)); err != nil {
			return
		}
	}
	if version < 3 { // Change primary key column by adding AUTOINCREMENT to make sure deleted IDs are not assigned again
		// we need to recreate the USERS-table.
		if _, err = db.Exec("ALTER TABLE USERS RENAME TO USERS_OLD"); err != nil {
			return
		}
		if _, err = db.Exec("DROP INDEX IF EXISTS IDX_MAIL_USR"); err != nil {
			return
		}
		if err = CreateNewUserDb(dbPath); err != nil {
			return
		}
		if _, err = db.Exec(`
			INSERT INTO USERS(id, firstName, lastName, pwhash,title, mail, lowercasemail, activationkey)
			 SELECT id, firstName, lastName, pwhash,title, mail, lowercasemail, activationkey FROM USERS_OLD`); err != nil {
			return
		}
		if _, err = db.Exec("DROP TABLE USERS_OLD"); err != nil {
			return
		}
		if _, err = db.Exec("UPDATE USERS SET level=1337 WHERE lowercasemail IN ('compufreak345@gmail.com', 'paul@defendtheplanet.net', 'kani@kani.kan')"); err != nil {
			return
		}
		version = 3
		if _, err = db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version)); err != nil {
			return
		}
	}
	if version < 4 {

		if err != nil {
			return
		}
		if _, err = db.Exec("ALTER TABLE USERS RENAME TO USERS_OLD"); err != nil {
			return
		}
		if _, err = db.Exec("DROP INDEX IF EXISTS IDX_MAIL_USR"); err != nil {
			return
		}
		if err = CreateNewUserDb(dbPath); err != nil {
			return
		}
		if _, err = db.Exec(`
			INSERT INTO USERS(id, firstName, lastName, pwhash,title, mail, lowercasemail, activationkey)
			 SELECT id, firstName, lastName, pwhash,title, mail, lowercasemail, activationkey FROM USERS_OLD`); err != nil {
			return
		}
		if _, err = db.Exec("DROP TABLE USERS_OLD"); err != nil {
			return
		}
		// http://souptonuts.sourceforge.net/readme_sqlite_tutorial.html - Logging all inserts, updates, deletes

		version = 4
		_, err = db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version))
		if err != nil {
			return
		}
	}
	if version < 5 {

		if err != nil {
			return
		}
		if _, err = db.Exec(`
CREATE TRIGGER insert_user_timeEnter AFTER  INSERT ON USERS
BEGIN

UPDATE USERS SET timeEnter = DATETIME('NOW')
         WHERE rowid = new.rowid;
END;
			`); err != nil {
			return
		}

		version = 5
		_, err = db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version))
		if err != nil {
			return
		}
	}
	if version < 6 {
		if _, err = db.Exec("ALTER TABLE PWRESETS ADD COLUMN lowercasemail string"); err != nil {
			return
		}

		version = 6
		_, err = db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version))
		if err != nil {
			return
		}
	}
	if version < 7 {
		if _, err = db.Exec("ALTER TABLE USERS ADD COLUMN tutorialDisabled INTEGER NOT NULL DEFAULT 0"); err != nil {
			return
		}

		version = 7
		_, err = db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version))
		if err != nil {
			return
		}
	}
	if version < 8 {
		if _, err = db.Exec(`
		ALTER TABLE USERS ADD COLUMN nextNotificationTime INTEGER NOT NULL DEFAULT 0;
		ALTER TABLE USERS ADD COLUMN notificationsEnabled INTEGER NOT NULL DEFAULT 1;

		`); err != nil {
			return
		}

		version = 8
		_, err = db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version))
		if err != nil {
			return
		}
	}
	if version < 11 {
		if _, err = db.Exec(`
		DROP TABLE IF EXISTS DevicesToUsers;
		DROP TABLE IF EXISTS Devices;

		CREATE TABLE Keys(guid STRING PRIMARY KEY,
		password string,
		userId INTEGER,
		CREATED INTEGER,
		FOREIGN KEY (userId) REFERENCES USERS(id)
)
			`); err != nil {
			return
		}
		version = 11
		_, err = db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version))
		if err != nil {
			return
		}
	}
	return
}


// GetLocationDb creates location DB for the given user if not existent and returns a connection to the db
func GetLocationDb(usrId int64) (database *sql.DB, err error) {
	dbPath, err := GetLocationDbPath(usrId)
	if err != nil {
		dbg.E(TAG," Could not get userDB-path for usr %d : ", usrId, err)
		return nil, err
	}

	database, err = libDbMan.GetLocationDb(dbPath,usrId)
	if err != nil {
		dbg.E(TAG," Could not open userDB for usr %d : ", usrId, err)
	}
	return
}

// GetLocationDbPath gets the database-path for the given user
func GetLocationDbPath(usrId int64) (dbPath string, err error) {
	var dbRoot = GetUserWorkDir(usrId)
	dbPath, err = libtools.GetCleanFilePath("trackrecords.db", dbRoot)
	if err != nil {
		return "", err
	}
	return
}

// GetLocationDbNumbers provides some statistical data for the users locationDb.
func GetLocationDbNumbers(usrId int64) (latestMigration string, devices int, trackrecords int64, tracks int64, keypoints int64, err error) {
	dbPath, err := GetLocationDbPath(usrId)
	if err != nil {
		return
	}
	return libDbMan.GetLocationDbNumbers(dbPath)
}
