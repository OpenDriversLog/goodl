// Package tutorialManager is used for getting and setting the tutorial-state of the current user.
package tutorialManager

import (
	"database/sql"
	"encoding/json"
	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/models"
	"github.com/OpenDriversLog/goodl-lib/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/goodl-lib/translate"
	"net/http"
)

// JSONTutorialAnswer is the default JSON-answer when requesting TutorialInfo.
type JSONTutorialAnswer struct {
	models.JSONAnswer
	TutorialInfo *TutorialInfo
}

// JSONGetTutorialInfo gets the current state of the users tutorial.
func JSONGetTutorialInfo(dbCon *sql.DB) (res JSONTutorialAnswer, err error) {
	res.TutorialInfo, err = GetTutorialInfo(dbCon)
	if err != nil {
		dbg.E(TAG, "Error getting tutorialInfo", err)
		res = GetBadJsonTutorialAnswer("internal server error")
		return
	}

	res.Success = true
	return
}

// GetBadJsonTutorialAnswer gets a bad JSONTutorialAnswer in case of an error.
func GetBadJsonTutorialAnswer(message string) JSONTutorialAnswer {
	return JSONTutorialAnswer{
		JSONAnswer: models.GetBadJSONAnswer(message),
	}
}

const NoDataGiven = "Please fill at least one entry."

// JSONUpdateTutorial updates the users tutorial state.
func JSONUpdateTutorial(tutJson string, usr *userManager.OdlUser, r *http.Request, T *translate.Translater, dbCon *sql.DB) (res models.JSONUpdateAnswer, err error) {

	if tutJson == "" {
		res = models.GetBadJSONUpdateAnswer(NoDataGiven, -1)
		return
	}
	t := &TutorialInfo{}
	err = json.Unmarshal([]byte(tutJson), t)
	if err != nil {
		dbg.W(TAG, "Could not read JSON %v in JSONUpdateTutorial : ", tutJson, err)
		res = models.GetBadJSONUpdateAnswer("Invalid format", -1)
		err = nil
		return
	}

	_, err = UpdateTutorial(t, usr, r, T, dbCon)

	if err != nil && err != tools.ErrNoChanges {
		dbg.E(TAG, "Error calling UpdateTutorial : ", err)
		res = models.GetBadJSONUpdateAnswer("Internal server error", -1)
		err = nil
		return
	}
	err = nil
	res.Success = true
	return
}
