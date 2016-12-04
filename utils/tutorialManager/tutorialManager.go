package tutorialManager

import (
	"database/sql"
	"github.com/Compufreak345/dbg"
	. "github.com/OpenDriversLog/goodl-lib/tools"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/goodl-lib/translate"
	"net/http"
)

const TAG = "goodl/utils/tutorialManager.go"

// GetTutorialInfo gets the users current tutorial state.
func GetTutorialInfo(dbCon *sql.DB) (tut *TutorialInfo, err error) {
	tut = &TutorialInfo{}

	row := dbCon.QueryRow("SELECT lastMilestone, disabled FROM TutorialInfo")
	err = row.Scan(&tut.LastMilestone, &tut.Disabled)
	if err != nil {
		dbg.E(TAG, "Error scanning tutorial-row", err)
		return nil, err
	}

	return
}

// UpdateTutorial updates the users current tutorial state.
func UpdateTutorial(c *TutorialInfo, usr *userManager.OdlUser, r *http.Request, T *translate.Translater, dbCon *sql.DB) (rowCount int64, err error) {
	tx, err := dbCon.Begin()
	if err != nil {
		dbg.E(TAG, "Error initialising transaction for UpdateTutorial : ", err)
	}
	tutorial, err := GetTutorialInfo(dbCon)
	if err != nil {
		dbg.E(TAG, "Error getting tutorial", err)
		return
	}
	disabledChanged := false
	vals := []interface{}{}
	firstVal := true
	valString := ""

	if c.LastMilestone != "" {
		AppendStringUpdateField("LastMileStone", &c.LastMilestone, &firstVal, &vals, &valString)
	}
	if c.Disabled != 0 && c.Disabled != tutorial.Disabled {
		AppendInt64UpdateField("disabled", &c.Disabled, &firstVal, &vals, &valString)
		disabledChanged = true
	}

	if firstVal {
		err = ErrNoChanges
		return
	}
	q := "UPDATE TutorialInfo SET " + valString
	var res sql.Result
	res, err = dbCon.Exec(q, vals...)
	if err != nil {
		dbg.E(TAG, "Error in dbCon.Exec for UpdateTutorial: %v ", err)
		return
	}
	rowCount, err = res.RowsAffected()

	if disabledChanged {
		_, err = userManager.UpdateUser(usr, &userManager.AdminEditableUserData{TutorialDisabled: int(c.Disabled), Id: usr.Id()}, r, T, false, false, false)
		if err != nil {
			dbg.E(TAG, "Error setting user TutorialDisabled : ", err)
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		dbg.E(TAG, "Error commiting transaction for UpdateTutorial: ", err)
		return
	}

	return
}
