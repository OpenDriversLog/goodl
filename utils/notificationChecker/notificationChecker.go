// Package notificationChecker is responsible for sending notification reminders.
package notificationChecker

import (
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/notificationManager"
	"database/sql"
	"time"
	"github.com/OpenDriversLog/goodl/utils/tools"
	"strings"
	"sync"
)

const TAG = "goodl/NotificationChecker"

// CheckForOverDue checks if there are any overdue notifications among all users and sends an email to those where it is.
// TODO: Test this thoroughly
func CheckForOverDue() (err error) {
	users,err := userManager.GetAllUsers()
	if err != nil {
		dbg.E(TAG, "Error getting users : ", err)
		return
	}
	t := time.Now().Unix()
	for _,u := range users {
		if u.NextNotificationTime > 0 && u.NextNotificationTime <= t {
			_,err = SendOverDueNotifications(u)
			if err != nil {
				dbg.E(TAG,"Error sending overdue notifications : ", err)
				return
			}
		}
	}

	return
}

// SendOverDueNotifications sends all overdue notifications for the given user.
func SendOverDueNotifications(user *userManager.AdminEditableUserData) (sent bool, err error) {
	var nots *[]*notificationManager.Notification
	nots,err = GetActiveNotificationsForUser(user.Id,nil,false)
	if err != nil {
		dbg.E(TAG,"Error getting active notifications for user %d : ", user.Id, err)
		return
	}
	t := time.Now().Unix()
	for _, n := range *nots {
		if t<=n.ExpirationTime {
			n.Message = strings.Replace(n.Message, "_*_Anrede_*_", string(user.Title)+" "+string(user.FirstName + " " + user.LastName), -1)
			err = tools.SendODLMail([]string{user.LoginName},n.Subject,n.Message,false)
			if err != nil {
				dbg.E(TAG,"Error sending reminder mail : ", err)
				return
			}
			sent = true
			n.WasSent = 1
			var uDbCon *sql.DB
			uDbCon,err = userManager.GetLocationDb(user.Id)
			if err != nil {
				dbg.E(TAG,"Error getting UserDb %d for sending notification : ",user.Id, err)
				return
			}
			_, err = notificationManager.UpdateNotification(n,uDbCon)
			if err != nil {
				dbg.E(TAG,"Error updating notification : ", err)
				return
			}
		}
	}
	return

}

// UpdateOverDue updates the time for the next notification for a user by copying it from the users database to the
// global ODL-Userdb.db
func UpdateOverDue(userId int64, dbCon *sql.DB) (err error){
	notifications,err := GetActiveNotificationsForUser(userId,dbCon,true)
	if err != nil {
		dbg.E(TAG,"Error getting active notifications for user %d: ",userId, err)
		return
	}
	nearestNotification := &notificationManager.Notification{ExpirationTime:0x7FFFFFFFFFFFFFFF}
	for _,n := range *notifications {
		if n.ExpirationTime < nearestNotification.ExpirationTime {
			nearestNotification = n
		}
	}
	var u *userManager.OdlUser
	u,err = userManager.GetUserById(userId)
	if err != nil {
		dbg.E(TAG,"Error getting user by id : ", err)
		return
	}
	if u.NextNotificationTime() != nearestNotification.ExpirationTime {
		upd := userManager.AdminEditableUserData{Id:u.Id(),NextNotificationTime:nearestNotification.ExpirationTime}
		_, err = userManager.UpdateUser(u,&upd,nil,nil,true,false,false)
		if err != nil {
			dbg.E(TAG,"Error updating ExpirationTime for user %d",u.Id(),err)
			return
		}
	}
	return
}
var ActiveNotificationsByUser = struct {
	sync.RWMutex
	m map[int64]*[]*notificationManager.Notification
}{m: make(map[int64]*[]*notificationManager.Notification)}

// GetActiveNotificationsForUser gets (and caches) the currently active notifications for the given user.
func GetActiveNotificationsForUser(userId int64,dbCon *sql.DB, forceRefresh bool) (nots *[]*notificationManager.Notification, err error) {
	ActiveNotificationsByUser.RLock()
	notAv := ActiveNotificationsByUser.m[userId] == nil
	ActiveNotificationsByUser.RUnlock()
	if forceRefresh || notAv {
		if dbCon == nil {
			dbCon,err = userManager.GetLocationDb(userId)
			if err != nil {
				dbg.E(TAG,"Error getting location db : ", err)
				return
			}
		}
		var ns []*notificationManager.Notification
		ns,err = notificationManager.GetActiveNotifications(true,dbCon)
		if err != nil {
			dbg.E(TAG," Error refreshing notifications for user %d : ",userId, err)
			return
		}
		ActiveNotificationsByUser.Lock()
		ActiveNotificationsByUser.m[userId] = &ns
		ActiveNotificationsByUser.Unlock()
	}
	ActiveNotificationsByUser.RLock()
	nots = ActiveNotificationsByUser.m[userId]
	ActiveNotificationsByUser.RUnlock()
	return

}