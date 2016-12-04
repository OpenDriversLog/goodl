package userManager

import "github.com/OpenDriversLog/goodl-lib/models/SQLite"

// UserEditableUserData contains the user-fields a normal user can edit by himself.
type UserEditableUserData struct {
	Title            string
	FirstName        string
	LastName         string
	LoginName        string
	Password         string
	RepeatedPassword string
	TutorialDisabled int
	NotificationsEnabled int

}

// AdminEditableUserData contains the user-fields that an admin can edit.
type AdminEditableUserData struct {
	Title            string
	FirstName        string
	LastName         string
	LoginName        string
	Password         string
	RepeatedPassword string
	Level            int
	ActivationKey    string
	Id               int64
	TutorialDisabled int
	NotificationsEnabled int
	NextNotificationTime int64

}

// Key represents a device-key for the automatic upload process.
type Key struct {
	GUID string
	Password string
	UserId models.NInt64
	Created int64
}