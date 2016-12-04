package models

import (
	"html/template"

	"github.com/OpenDriversLog/webfw"
	"github.com/OpenDriversLog/goodl-lib/translate"
)

// RegisterMailModel is used for sending ActivationLinks by E-Mail
type RegisterMailModel struct {
	webfw.Model
	E *RegisterMailEnhance
}

// HTML is a shortcut to access the RegisterMailModel Translater.T-method
func (m *RegisterMailModel) HTML(s string, args ...interface{}) template.HTML {
	return template.HTML(m.E.T.T(s, args...))
}

// RegisterMailEnhance enhances RegisterMailModel, used for sending ActivationLinks by E-Mail.
type RegisterMailEnhance struct {
	ActivationLink string
	Name           string
	T              *translate.Translater
}

// C returns an interface-pointer to the RegisterMailEnhance-part of the RegisterMailModel.
func (r RegisterMailModel) C() interface{} {
	return r.E
}
