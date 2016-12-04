// Package Models contains some Data-Models, mainly for Goodl-Emails.
package models

import (
	"html/template"

	"github.com/OpenDriversLog/webfw"
	"github.com/OpenDriversLog/goodl-lib/translate"
)

// EditUserMailModel is used for sending an mail on user-data-changes (e.g. Name changes, password changes...)
type EditUserMailModel struct {
	webfw.Model
	E *EditUserMailEnhance
}

// HTML is a shortcut to access the EditUserMailModels Translater.T-method
func (m *EditUserMailModel) HTML(s string, args ...interface{}) template.HTML {
	return template.HTML(m.E.T.T(s, args...))
}

// EditUserMailEnhance is used in EditUserMailModel that is used for sending an mail on user-data-changes (e.g. Name changes, password changes...)
type EditUserMailEnhance struct {
	Name         string
	PreviousName string
	Changes      []string
	T            *translate.Translater
}

// C returns an interface-pointer to the EditUserMailEnhance-Part of the EditUserMailModel.
func (r EditUserMailModel) C() interface{} {
	return r.E
}
