package controllers

import (
	"github.com/OpenDriversLog/goodl-lib/models"
	S "github.com/OpenDriversLog/goodl-lib/models/SQLite"
)

type BetaUser struct {
	Id              int64
	Name            S.NString
	Vorname         S.NString
	Email           S.NString
	Wants2BePilot   S.NInt64
	WantsNewsletter S.NInt64
	Anrede          S.NString
}
type JSONBetaManAnswer struct {
	models.JSONAnswer
	BetaUsers []*BetaUser
}

type JSONRPCCall struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
	Id     string                 `json:"id"`
}

type JSONRPCResponse struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	Id     string      `json:"id"`
}

type SendMailRequest struct{
	Message string
	UsrIds []string
	Subject string
	SurveyId string
}