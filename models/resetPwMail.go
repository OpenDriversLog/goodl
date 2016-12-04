package models

import (
	"html/template"

	"github.com/OpenDriversLog/webfw"
	"github.com/OpenDriversLog/goodl-lib/translate"
)

// ResetPwMailModel is used for sending Reset-Password-E-Mails.
type ResetPwMailModel struct {
	webfw.Model
	E *ResetPwMailEnhance
}

// HTML is a shortcut to access the ResetPwMailModel Translater.T-method
func (m *ResetPwMailModel) HTML(s string, args ...interface{}) template.HTML {
	return template.HTML(m.E.T.T(s, args...))
}

// ResetPwMailEnhance enhances ResetPwMailModel, used for sending Reset-Password-E-Mails.
type ResetPwMailEnhance struct {
	ResetPwLink string
	Name        string
	T           *translate.Translater
}

// C returns an interface-pointer to the ResetPwMailEnhance-part of the ResetPwMailModel.
func (r ResetPwMailModel) C() interface{} {
	return r.E
}
