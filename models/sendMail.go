package models

import (
	"html/template"

	"github.com/OpenDriversLog/webfw"
	"github.com/OpenDriversLog/goodl-lib/translate"
)

// SendUSerMailModel is used for sending personal messages to ODL-users.
type SendUserMailModel struct {
	webfw.Model
	E *SendUserMailEnhance
}

// HTML is a shortcut to access the SendUserMailModel Translater.T-method
func (m *SendUserMailModel) HTML(s string, args ...interface{}) template.HTML {
	return template.HTML(m.E.T.T(s, args...))
}

// SendUserMailEnhance enhances SendUserMailModel, used for sending personal messages to ODL-users.
type SendUserMailEnhance struct {
	Mail   string
	Anrede string
	UserId int64
	Text   interface{}
	T      *translate.Translater
}

// C returns an interface-pointer to the SendUserMailEnhance-Part of the SendUserMailModel.
func (r SendUserMailModel) C() interface{} {
	return r.E
}
