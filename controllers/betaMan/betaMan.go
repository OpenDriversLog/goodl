package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/OpenDriversLog/webfw"
	. "github.com/OpenDriversLog/goodl-lib/translate"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/models"
	. "github.com/OpenDriversLog/goodl-lib/tools"

	"github.com/OpenDriversLog/goodl/utils/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
)

const TAG = dbg.Tag("goodl/BetaManController.go")
const limeSurveyURL = "http://YourServer:7201"
const limeSurveyPublicURL = "https://YourServer/limesurvey"

// BetaManController serves for editing beta-users and sending beta-invite-Mails. It can also create invites for
// a limesurvey-survey.
type BetaManController struct {
}

const NoDataGiven = "Please fill at least one entry."
// GetViewData CRUDs beta-users depending on the "action"-parameter.
// read : Return list of all betaUsers
// create : Creates the given betaUser
// update : Updates the given betaUser
// delete : Deletes the given betaUser
// sendMail : Sends the given mail in parameter "data", optionally also creating a new survey link and replacing placeholders in the mail

func (BetaManController) GetViewData(ctx context.Context, r *http.Request) (vd webfw.ViewData, viewPath string, vShared string, err error) {
	defer func() {
		if err := recover(); err != nil {
			vd, viewPath, vShared, err = webfw.RecoverForTag(TAG, "GetViewData", r, errors.New(fmt.Sprintf("%s", err)), true)
		}
	}()
	dbg.D(TAG, "BetaManController called")
	T := ctx.Value("T").(*Translater)

	vd = webfw.ViewData{
		T:    T,
		Data: make(map[string]interface{}),
	}
	viewPath = "views/showDataMessage.htm"
	usr, _, _ := userManager.GetUserWithSession(r)
	if usr == nil || !usr.IsLoggedIn() {
		return webfw.GetErrorViewData(TAG, 500, dbg.WTF(TAG, "This is impossible. How did I get to BetaManController without logging in?"), "", nil, false)
	}
	if usr.Level() != 1337 {
		return webfw.GetErrorViewData(TAG, 403, dbg.W(TAG, "User with id "+strconv.FormatInt(usr.Id(), 10)+" tried to open BetaMan without permission..."), "", nil, false)
	}
	dbCon, err := GetBetaDbCon()
	defer dbCon.Close()
	var marshaled []byte

	if strings.TrimSpace(r.FormValue("action")) == "" {
		marshaled, err = json.Marshal(models.GetBadJSONAnswer("Please define an action."))
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal models.GetBadJSONAnswer(\"Unknown action\") \n error: %v", err), "", err, true)
		}
		vd.Data["Message"] = template.HTML(string(marshaled))
		return
	}
	if r.FormValue("action") == "sendMail" {
		var res models.JSONAnswer
		res, err = JSONSendMail(r.FormValue("data"), dbCon, T)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get BetaMan.JSONGetBetaUsers :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	if err != nil {
		return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get BetaDb :( ", err), "", err, true)
	}

	if r.FormValue("action") == "read" {
		var res JSONBetaManAnswer
		res, err = JSONGetBetaUsers(dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get BetaMan.JSONGetBetaUsers :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "create" {
		var res models.JSONInsertAnswer
		res, err = JSONCreateBetaUser(r.FormValue("betaUser"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get betaUserManager.JSONCreateBetaUser :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "update" {
		var res models.JSONUpdateAnswer
		res, err = JSONUpdateBetaUser(r.FormValue("betaUser"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get BetaMan.JSONUpdateAnswer :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}
	if r.FormValue("action") == "delete" {
		var res models.JSONDeleteAnswer
		res, err = JSONDeleteBetaUser(r.FormValue("betaUser"), dbCon)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "Could not get BetaMan.JSONDeleteBetaUser :( ", err), "", err, true)
		}
		tools.TranslateErrors(&res.ErrorMessage, res.Errors, T)

		marshaled, err = json.Marshal(res)
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal result: %v \n error: %v", res, err), "", err, true)
		}
	}

	mString := string(marshaled)
	if mString == "" {
		marshaled, err = json.Marshal(models.GetBadJSONAnswer("Unknown action"))
		if err != nil {
			return webfw.GetErrorViewData(TAG, 500, dbg.E(TAG, "unable to marshal models.GetBadJSONAnswer(\"Unknown action\") \n error: %v", err), "", err, true)
		}
		mString = string(marshaled)
	}
	vd.Data["Message"] = template.HTML(mString)

	return
}

// SendThisMail sends the given mails by the given SendMailRequest. Replaces some text parts like Title
// and can create survey invite links that get inserted by replacing as well.
func SendThisMail(data SendMailRequest, dbCon *sql.DB, T *Translater) (status string) {
	usrs, err := GetBetaUsers(dbCon)
	if err != nil {
		status = "Error getting users :("
	}
	usrsById := make(map[string]*BetaUser)

	betreff := data.Subject
	if betreff == "" {
		status = "Error : No Betreff"
		return
	}

	message := data.Message
	message = strings.Replace(message, "_*_Abschied_*_", T.T("Mail_End"), -1)
	message = strings.Replace(message, "\n", "<br/>", -1)
	for _, u := range usrs {
		usrsById[strconv.FormatInt(u.Id, 10)] = u
	}
	surveyId := data.SurveyId
	var sessionKey string
	if surveyId != "" {
		_pw, err := ioutil.ReadFile(webfw.Config().RootDir + "/DONTADDTOGIT/limesurvey.txt")

		pw := strings.Replace(string(_pw), "\n", "", -1)

		if err != nil {
			dbg.E(TAG, "Error in SendThisMail while reading limesurvey pw : ", err)
			return
		}
		call := JSONRPCCall{
			Method: "get_session_key",
			Params: map[string]interface{}{"1": "admin", "2": pw},
			Id:     "OpenSession",
		}
		jsoncall, err := json.Marshal(call)

		resp, err := http.Post(limeSurveyURL+"/index.php/admin/remotecontrol", "application/json", strings.NewReader(string(jsoncall)))
		defer resp.Body.Close()
		if err != nil {
			status += "Error while opening survey session : " + err.Error()
			err = nil
			return
		}
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			status += "Error while reading survey session : " + err.Error()
			err = nil
			return
		}
		jsonResp := &JSONRPCResponse{}
		err = json.Unmarshal(contents, &jsonResp)
		if err != nil {
			status += "Error while unmarshaling survey session : " + err.Error()
			err = nil
			return
		}

		sessionKey = jsonResp.Result.(string)

		defer func() {
			call := JSONRPCCall{
				Method: "release_session_key",
				Params: map[string]interface{}{"1": sessionKey},
				Id:     "OpenSession",
			}
			jsoncall, _ := json.Marshal(call)
			http.Post(limeSurveyURL+"/index.php/admin/remotecontrol", "application/json", strings.NewReader(string(jsoncall)))

		}()
	}
	for _, usrId := range data.UsrIds {
		status += "<br/>"
		usr := usrsById[usrId]
		if usr == nil {
			status += "Error : Could not find user with ID " + usrId
		} else {
			status += string(usr.Email) + " : "
			umsg := strings.Replace(message, "_*_Anrede_*_", string(usr.Anrede)+" "+string(usr.Name), -1)
			var surveyKey string
			if surveyId != "" {
				// Generate a survey user with token!

				data := make(map[string]map[string]interface{}, 0)

				data["primary"] = map[string]interface{}{"email": usr.Email,
					"lastname":  usr.Name,
					"firstname": usr.Vorname,
				}

				call := JSONRPCCall{
					Method: "add_participants",
					Params: map[string]interface{}{"1": sessionKey, "2": surveyId, "3": data, "4": true},
					Id:     "requestParticipant" + strconv.FormatInt(usr.Id, 10),
				}
				jsoncall, err := json.Marshal(call)

				resp, err := http.Post(limeSurveyURL+"/index.php/admin/remotecontrol", "application/json", strings.NewReader(string(jsoncall)))
				defer resp.Body.Close()
				if err != nil {
					status += "Error while opening add_participants : " + err.Error()
					err = nil
					return
				}
				contents, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					status += "Error while reading add_participants : " + err.Error()
					err = nil
					return
				}
				jsonResp := &JSONRPCResponse{}
				err = json.Unmarshal(contents, &jsonResp)
				if err != nil {
					status += "Error while unmarshaling add_participants : " + err.Error()
					err = nil
					return
				}

				respp := jsonResp.Result.(map[string]interface{})
				surveyKey = respp["primary"].(map[string]interface{})["token"].(string)
				surveyLink := limeSurveyPublicURL + "/index.php/" + surveyId + "/lang-de/token/" + surveyKey
				umsg = strings.Replace(umsg, "_*_SurveyLink_*_", surveyLink, -1)
			}

			err = tools.SendODLMail([]string{string(usr.Email)}, betreff, umsg, false)
			if err != nil {
				status += "Error : " + err.Error()
			} else {
				status += "Success"
			}
		}
	}
	return
}

// JSONSendMail takes an JSON-SendMailRequest-objects and calls SendThisMail with the given parameters.
func JSONSendMail(jsonMail string, dbCon *sql.DB, T *Translater) (res models.JSONAnswer, err error) {
	m := SendMailRequest{}
	if jsonMail == "" {
		res = models.GetBadJSONAnswer(NoDataGiven)
		return
	}
	err = json.Unmarshal([]byte(jsonMail), &m)
	if err != nil {
		dbg.W(TAG, "Could not read JSON %v in JSONSendMail : ", jsonMail, err)
		res = models.GetBadJSONAnswer("Invalid format")
		err = nil
		return
	}
	status := SendThisMail(m, dbCon, T)
	if strings.Contains(status, "Error :") {
		res.Success = false
		res.Error = true
	} else {
		res.Success = true
	}
	res.ErrorMessage = status
	return
}

// JSONGetBetaUsers Gets the current beta users as JSON.
func JSONGetBetaUsers(dbCon *sql.DB) (res JSONBetaManAnswer, err error) {
	res = JSONBetaManAnswer{}

	res.BetaUsers, err = GetBetaUsers(dbCon)
	if err != nil {
		dbg.E(TAG, "Error getting beta users : ", err)
		err = nil
		res = GetBadJsonBetaManAnswer("Unknown error while getting contacts")
		return
	}

	res.Success = true
	return
}

// GetBadJsonBetaManAnswer returns an JSONBetaManAnswer representing an error
func GetBadJsonBetaManAnswer(message string) JSONBetaManAnswer {
	return JSONBetaManAnswer{
		JSONAnswer: models.GetBadJSONAnswer(message),
	}
}

// GetBetaUsers returns an array of current beta users.
func GetBetaUsers(dbCon *sql.DB) (users []*BetaUser, err error) {
	res, err := dbCon.Query("SELECT id,name,vorname,email,wants2bePilot,wantsNewsletter,anrede FROM BetaUser")
	if err != nil {
		return
	}
	users = make([]*BetaUser, 0)
	for res.Next() {
		usr := &BetaUser{}
		res.Scan(&usr.Id, &usr.Name, &usr.Vorname, &usr.Email, &usr.Wants2BePilot, &usr.WantsNewsletter, &usr.Anrede)
		if err != nil {
			return
		}
		users = append(users, usr)
	}

	return
}

// JSONUpdateBetaUser takes a JSON-string representing a BetaUser and updates the BetaUser with the given ID to match.
func JSONUpdateBetaUser(jsonUsr string, dbCon *sql.DB) (res models.JSONUpdateAnswer, err error) {
	c := &BetaUser{}
	if jsonUsr == "" {
		res = models.GetBadJSONUpdateAnswer(NoDataGiven, -1)
		return
	}
	err = json.Unmarshal([]byte(jsonUsr), c)
	if err != nil {
		dbg.W(TAG, "Could not read JSON %v in UpdateContactJSON : ", jsonUsr, err)
		res = models.GetBadJSONUpdateAnswer("Invalid format", -1)
		err = nil
		return
	}
	var rowCount int64
	rowCount, err = UpdateBetaUser(c, dbCon)
	if err != nil {
		if err == ErrNoChanges {
			err = nil
			res = models.GetBadJSONUpdateAnswer(NoDataGiven, -1)
			return
		}
		dbg.E(TAG, "Error in JSONUpdateBetaUser UpdateBetaUser: ", err)
		err = nil
		res = models.GetBadJSONUpdateAnswer("Internal server error", c.Id)
		return
	}
	res.Id = c.Id
	res.RowCount = rowCount
	res.Success = true
	return
}

// UpdateBetaUser updates the betaUser with the given ID with the given fields in the BetaUser-object
func UpdateBetaUser(u *BetaUser, dbCon *sql.DB) (rowCount int64, err error) {
	vals := []interface{}{}
	firstVal := true
	valString := ""

	if u.WantsNewsletter != 0 {
		AppendNInt64UpdateField("wantsNewsletter", &u.WantsNewsletter, &firstVal, &vals, &valString)
	}
	if u.Wants2BePilot != 0 {
		AppendNInt64UpdateField("wants2BePilot", &u.Wants2BePilot, &firstVal, &vals, &valString)
	}
	if u.Name != "" {
		AppendNStringUpdateField("name", &u.Name, &firstVal, &vals, &valString)
	}
	if u.Vorname != "" {
		AppendNStringUpdateField("vorname", &u.Vorname, &firstVal, &vals, &valString)
	}
	if u.Email != "" {
		AppendNStringUpdateField("email", &u.Email, &firstVal, &vals, &valString)
	}
	if u.Anrede != "" {
		AppendNStringUpdateField("anrede", &u.Anrede, &firstVal, &vals, &valString)
	}

	if firstVal {
		err = ErrNoChanges
		return
	}
	q := "UPDATE BetaUser SET " + valString + " WHERE id=?"
	vals = append(vals, u.Id)
	var res sql.Result
	res, err = dbCon.Exec(q, vals...)
	if err != nil {
		dbg.E(TAG, "Error in 2nd dbCon.Exec for UpdateContact: %v ", err)

		return
	}
	rowCount, err = res.RowsAffected()

	return
}

// JSONInsertBetaUser creates a new beta user by the given json-string
func JSONInsertBetaUser(jsonUsr string, dbCon *sql.DB) (res models.JSONInsertAnswer, err error) {
	u := &BetaUser{}
	if jsonUsr == "" {
		res = models.GetBadJSONInsertAnswer(NoDataGiven)
		return
	}

	dbg.D(TAG, "got jsonusr string", jsonUsr)
	err = json.Unmarshal([]byte(jsonUsr), u)
	if err != nil {
		dbg.W(TAG, "Could not read JSON %v in DeleteContactJSON : ", jsonUsr, err)
		res = models.GetBadJSONInsertAnswer("Invalid format")
		err = nil
		return
	}
	var lastKey int64
	lastKey, err = InsertBetaUser(u, dbCon)
	if err != nil {
		dbg.E(TAG, "Error in InsertBetaUserJSJON InsertBetaUser: ", err)
		err = nil
		res = models.GetBadJSONInsertAnswer("Internal server error")
		return
	}
	res.LastKey = lastKey
	res.Success = true
	return
}

// InsertBetaUser creates a new beta user
func InsertBetaUser(user *BetaUser, dbCon *sql.DB) (lastKey int64, err error) {
	var res sql.Result

	dbg.D(TAG, "golang betauser:", user)

	res, err = dbCon.Exec(`INSERT INTO BetaUser (name, vorname, email, wants2bePilot, wantsNewsletter, anrede) 
VALUES (?, ?, ?, ?, ?, "Sehr geehrte(r) Herr/Frau")`,
		user.Name, user.Vorname, user.Email, user.Wants2BePilot, user.WantsNewsletter)
	if err != nil {
		dbg.E(TAG, "Error in InsertBetaUser : ", err)
	} else {
		lastKey, err = res.LastInsertId()
		if err != nil {
			dbg.E(TAG, "Error in InsertBetaUser get RowsAffected : ", err)
		}
	}

	return
}

// JSONDeleteBetaUser deletes the given beta user by the ID given in the JSON object
func JSONDeleteBetaUser(jsonUsr string, dbCon *sql.DB) (res models.JSONDeleteAnswer, err error) {
	u := &BetaUser{}
	if jsonUsr == "" {
		res = models.GetBadJSONDeleteAnswer(NoDataGiven, -1)
		return
	}
	err = json.Unmarshal([]byte(jsonUsr), u)
	if err != nil {
		dbg.W(TAG, "Could not read JSON %v in DeleteContactJSON : ", jsonUsr, err)
		res = models.GetBadJSONDeleteAnswer("Invalid format", -1)
		err = nil
		return
	}
	var rowCount int64
	rowCount, err = DeleteBetaUser(u.Id, dbCon)
	if err != nil {
		dbg.E(TAG, "Error in DeleteContactJSJON DeleteContact: ", err)
		err = nil
		res = models.GetBadJSONDeleteAnswer("Internal server error", u.Id)
		return
	}
	res.RowCount = rowCount
	res.Success = true
	return
}

// DeleteBetaUser deletes the betaUser with the given ID
func DeleteBetaUser(id int64, dbCon *sql.DB) (rowCount int64, err error) {
	var res sql.Result
	res, err = dbCon.Exec("DELETE FROM BetaUser WHERE id=?", id)
	if err != nil {
		dbg.E(TAG, "Error in DeleteBetaUser : ", err)
	} else {
		rowCount, err = res.RowsAffected()
		if err != nil {
			dbg.E(TAG, "Error in DeleteBetaUser get RowsAffected : ", err)
		}
	}

	return
}

// GetBetDabCon returns an open DB-connection to the betaUsers-DB
func GetBetaDbCon() (con *sql.DB, err error) {
	dbPath := webfw.Config().SharedDir + "/betaanmeldung.sqlite3"
	dbg.I(TAG, "BetaUserPath : %s", dbPath)

	return openDbCon(dbPath)
}

// openDbCon opens a database connection to the given DB-Path and updates the DB if necessary
func openDbCon(dbPath string) (*sql.DB, error) {
	database, err := sql.Open("SQLITE", dbPath)

	if err != nil {
		dbg.E(TAG, "Failed to create DB handle at openDbCon()")
		return nil, err
	}
	if err2 := database.Ping(); err2 != nil {
		dbg.E(TAG, "Failed to keep connection alive at openDbCon()")
		return nil, err
	}
	dbg.D(TAG, "Databaseconnection established")
	dbg.D(TAG, "Testing if we got field anrede")

	_, err = database.Exec("UPDATE BetaUser SET anrede='Sehr geehrte(r) Frau/Herr' WHERE 0=1")
	if err != nil {
		dbg.I(TAG, "Adding anrede-field")
		_, err = database.Exec("ALTER TABLE BetaUser ADD anrede TEXT NOT NULL Default 'Sehr geehrte(r) Herr/Frau'")
	}
	return database, err
}

// CreateBetaUser creates the given betaUser
func CreateBetaUser(betaUser *BetaUser, dbCon *sql.DB) (key int64, err error) {

	vals := []interface{}{betaUser.Email, betaUser.Wants2BePilot, betaUser.WantsNewsletter}
	valString := "?,?,?"
	insFields := "Email,Wants2BePilot,WantsNewsletter"

	if betaUser.Anrede != "" {
		insFields += ",Anrede"
		valString += ",?"
		vals = append(vals, betaUser.Anrede)
	}
	if betaUser.Name != "" {
		insFields += ",Name"
		valString += ",?"
		vals = append(vals, betaUser.Name)
	}
	if betaUser.Vorname != "" {
		insFields += ",Vorname"
		valString += ",?"
		vals = append(vals, betaUser.Vorname)
	}
	q := "INSERT INTO BetaUser(" + insFields + ") VALUES(" + valString + ")"
	var res sql.Result
	res, err = dbCon.Exec(q, vals...)
	if err != nil {
		dbg.E(TAG, "Error in dbCon.Exec for CreateBetaUsers: %v ", err)
		return
	}
	key, err = res.LastInsertId()
	return
}

// JSONCreateBetaUser creates the BetaUser given in the JSON string
func JSONCreateBetaUser(betaUserJson string, dbCon *sql.DB) (res models.JSONInsertAnswer, err error) {
	c := &BetaUser{}
	if betaUserJson == "" {
		res = models.GetBadJSONInsertAnswer(NoDataGiven)
		return
	}
	err = json.Unmarshal([]byte(betaUserJson), c)
	if err != nil {
		dbg.W(TAG, "Could not read JSON %v in JSONCreateBetaUser : ", betaUserJson, err)
		res = models.GetBadJSONInsertAnswer("Invalid format")
		err = nil
		return
	}
	var key int64
	key, err = CreateBetaUser(c, dbCon)
	if err != nil {
		dbg.E(TAG, "Error in JSONCreateBetaUser CreateBetaUser: ", err)
		err = nil
		res = models.GetBadJSONInsertAnswer("Internal server error")
		return
	}
	res.LastKey = key
	c.Id = res.LastKey
	res.Success = true
	return

}
