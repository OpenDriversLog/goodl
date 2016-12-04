// Package tools provides some unsorted tools for creating gitlab issues, sending Mails, creating random strings,
// translating errors and getting deviceIds from a HTTP-Request
package tools

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"net/http"

	"encoding/json"

	"strconv"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/webfw"
	"github.com/OpenDriversLog/goodl-lib/translate"
)

const TAG = "goodl/tools.go"

var gitlabToken string

// GitlabToken() returns the GitLabToken from webfw.Config().RootDir + "/DONTADDTOGIT/gitlab.txt", and caches it for the next call.
func GitlabToken() string {
	if gitlabToken == "" {
		_gitlabToken, err := ioutil.ReadFile(webfw.Config().RootDir + "/DONTADDTOGIT/gitlab.txt")
		if err != nil {
			dbg.E(TAG, "Error in GitlabToken while reading token : ", err)
			return ""
		}
		gitlabToken = strings.Replace(string(_gitlabToken), "\n", "", -1)
	}
	return gitlabToken
}

var gitPath string

// GitPath returns the path of the gitlab from webfw.Config(), and caches it.
func GitPath() string {
	if gitPath == "" {
		gitPath = webfw.Config().GitLabPath
	}
	return gitPath
}

// SetGitPath sets the gitPath
func SetGitPath(path string) {
	gitPath = path
}
// SetGitlabtoken sets the gitlabToken
func SetGitlabToken(token string) {
	gitlabToken = token
}

// SendODLMail sends an mail from info@opendriverslog.de to the given addresses (or to test@opendriverslog.de if environment is development)
// TODO: Make this more configurable for other mailservers.
// !!!TO BE DISCUSSED!!!
func SendODLMail(to []string, subject string, message string, withIssue bool) (err error) {
	dbg.D(TAG, "Start SendMail")
	//because we can't send emails from our local machines, we skip the complete process
	if webfw.Config().Environment == "development" /*|| webfw.Config().Environment == "test"*/ {
		//dbg.W(TAG, "not sending any mails because we are in "+webfw.Config().Environment)
		dbg.W(TAG,"We are in " + webfw.Config().Environment + " - sending all mails to test@opendriverslog.de")
		to = []string{"test@opendriverslog.de"}
		//return
	}
	if webfw.Config().Environment == "test" {
		dbg.W(TAG,"We are in " + webfw.Config().Environment + " - not sending mails.")
		return
	}
	defer func() {
		if errr := recover(); errr != nil {
			dbg.E(TAG, "panic in SendODLMail: ", errr)
			err = errors.New(fmt.Sprintf("%v", err))
		}
	}()
	_pw, err := ioutil.ReadFile(webfw.Config().RootDir + "/DONTADDTOGIT/mail.txt")

	pw := strings.Replace(string(_pw), "\n", "", -1)

	if err != nil {
		dbg.E(TAG, "Error in SendODLMail while reading pw : ", err)
		return
	}

	if withIssue {
		url, err := CreateIssue("E-Mail : "+subject, message, "17")
		if err != nil {
			dbg.E(TAG, "Error creating issue - I will still send Email! ", err)
			err = nil
		}
		message += "<br/><br/> Link to issue : <a href='" + url + "' target='_blank'>" + url + "</a>"
	}

	var smtpHost = webfw.Config().SmtpHost
	var smtpPort, _ = strconv.Atoi(webfw.Config().SmtpPort)
	err = SendEmail(smtpHost, smtpPort, "info@opendriverslog.de", string(pw), to, subject, message)
	if err != nil {
		dbg.E(TAG, "Error in SendEmail : ", err)
		return
	}

	return
}

// CreateIssue creates a new issue in gitlab.
func CreateIssue(title string, details string, projectId string) (issueUrl string, err error) {
	defer func() {
		if errr := recover(); errr != nil {
			dbg.E(TAG, "Panic in CreateIssue : ", errr)
			err = errors.New(fmt.Sprintf("", errr))
			return
		}
	}()

	var vals url.Values = map[string][]string{
		"id":          {projectId},
		"title":       {title},
		"description": {details},
		"labels":      {"Mail"},
	}
	r, err := http.PostForm(GitPath()+"/api/v3/projects/"+projectId+"/issues?private_token="+GitlabToken(), vals)
	if err != nil {
		dbg.E(TAG, "Error in CreateIssue while posting : ", err)
		return
	}
	defer r.Body.Close()
	if r.StatusCode != 201 {
		dbg.E(TAG, "Status was not 201 created :( ", r)
		err = errors.New("CreateIssue did not get 201 response")
		return
	}

	respTxt, err := ioutil.ReadAll(r.Body)

	if err != nil {
		dbg.E(TAG, "Error in CreateIssue while reading response body : ", err)
		return
	}

	parsedResp := make(map[string]interface{})
	err = json.Unmarshal(respTxt, &parsedResp)

	issueUrl = GitPath() + "/odl/goodl/issues/" + strconv.FormatFloat(parsedResp["id"].(float64), 'f', 0, 32)

	return
}

// The following is from :https://github.com/plouc/go-gitlab-client/blob/master/issue.go

// Milestone represents a gitlab-Milestone.
type Milestone struct {
	Id          int    `json:"id,omitempty"`
	IId         int    `json:"iid,omitempty"`
	ProjectId   int    `json:"project_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	State       string `json:"state,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// Issue represents a gitlab-issue
type Issue struct {
	Id          int        `json:"id"`
	IId         int        `json:"iid"`
	ProjectId   int        `json:"project_id,omitempty"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	Labels      []string   `json:"labels,omitempty"`
	Milestone   *Milestone `json:"milestone,omitempty"`
	Assignee    *User      `json:"assignee,omitempty"`
	Author      *User      `json:"author,omitempty"`
	State       string     `json:"state,omitempty"`
	CreatedAt   string     `json:"created_at,omitempty"`
	UpdatedAt   string     `json:"updated_at,omitempty"`
}

// IssueRequest represents a gitlab-issueRequest
type IssueRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	AssigneeId  int    `json:"assignee_id,omitempty"`
	MilestoneId int    `json:"milestone_id,omitempty"`
	Labels      string `json:"labels,omitempty"`
}

// User represents a gitlab-user
type User struct {
	Id            int    `json:"id,omitempty"`
	Username      string `json:"username,omitempty"`
	Email         string `json:"email,omitempty"`
	Name          string `json:"name,omitempty"`
	State         string `json:"state,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	Bio           string `json:"bio,omitempty"`
	Skype         string `json:"skype,omitempty"`
	LinkedIn      string `json:"linkedin,omitempty"`
	Twitter       string `json:"twitter,omitempty"`
	ExternUid     string `json:"extern_uid,omitempty"`
	Provider      string `json:"provider,omitempty"`
	ThemeId       int    `json:"theme_id,omitempty"`
	ColorSchemeId int    `json:"color_scheme_id,color_scheme_id"`
}

// CheckIfIssueExists searches for an issue with the given title in the given projectId
func CheckIfIssueExists(title string, projectId string) (issue *Issue, err error) {

	r, err := http.Get(GitPath() + "/api/v3/projects/" + projectId + "/issues?private_token=" + GitlabToken())
	if err != nil {
		dbg.E(TAG, "Error contacting gitlab :( ", err)
		return
	}

	var issues []Issue
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		dbg.E(TAG, "Status was not OK :( ", r)
		err = errors.New("Did not get OK response")
		return
	}

	respTxt, err := ioutil.ReadAll(r.Body)

	if err != nil {
		dbg.E(TAG, "Error in CheckIfIssueExists while reading response body : ", err)
		return
	}
	err = json.Unmarshal(respTxt, &issues)
	for _, i := range issues {
		if i.Title == title {
			issue = &i
			return
		}
	}
	return
	dbg.D(TAG, "End CheckIfIssueExists")
	return
}

// AddCommentToIssue adds an comment to the given issue.
func AddCommentToIssue(comment string, projectId string, issueId int) (err error) {
	var vals url.Values = map[string][]string{
		"id":       {projectId},
		"body":     {comment},
		"issue_id": {strconv.Itoa(issueId)},
	}
	r, err := http.PostForm(GitPath()+"/api/v3/projects/"+projectId+"/issues/"+strconv.Itoa(issueId)+"/notes?private_token="+GitlabToken(), vals)
	if err != nil {
		return
	}
	if r.StatusCode != 201 {
		dbg.E(TAG, "Status was not 201 created :( ", r)
		err = errors.New("CreateIssue did not get 201 response")
		return
	}
	return
}

// SendEmail sends an email to the given to-addresses.
// This is based on http://www.goinggo.net/2013/06/send-email-in-go-with-smtpsendmail.html, extended with stuff from https://golang.org/src/net/smtp/smtp.go
func SendEmail(host string, port int, userName string, password string, to []string, subject string, message string) (err error) {

	parameters := struct {
		From    string
		To      string
		Subject string
		Message template.HTML
	}{
		userName,
		strings.Join([]string(to), ","),
		subject,
		template.HTML(message),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	buffer := new(bytes.Buffer)

	template := template.Must(template.New("emailTemplate").Delims("{[{", "}]}").Parse(emailScript()))
	template.Execute(buffer, &parameters)

	// Inspired / Partially taken from https://golang.org/src/net/smtp/smtp.go (starting line 274)
	c, err := smtp.Dial(addr)
	if err != nil {
		dbg.E(TAG, "Error while dialing : ", err)
		return err
	}
	defer c.Close()
	config := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         addr,
	}
	err = c.StartTLS(config)

	if err != nil {
		dbg.E(TAG, "Error while starting tls : ", err)
		return err
	}

	auth := smtp.PlainAuth("", userName, password, host)
	if err = c.Auth(auth); err != nil {
		dbg.E(TAG, "Error while starting c.Auth : ", err)
		return err
	}

	if err = c.Mail(userName); err != nil {
		dbg.E(TAG, "Error while starting c.Mail : ", err)
		return err
	}
	for _, t := range to {
		if err = c.Rcpt(t); err != nil {
			dbg.E(TAG, "Error while starting c.Rcpt for : ", t, err)
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		dbg.E(TAG, "Error for c.Data() : ", err)
		return err
	}
	_, err = w.Write(buffer.Bytes())
	if err != nil {
		dbg.E(TAG, "Error for w.Write() : ", err)
		return err
	}
	if err = w.Close(); err != nil {
		dbg.E(TAG, "Error for w.Close() : ", err)
		return err
	}

	return c.Quit()
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_")

// RandSeq Generates a random string with given length
// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func RandSeq(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// emailScript is a template-string for starting E-Mails.
func emailScript() (script string) {
	return `From: {[{.From}]}
To: {[{.To}]}
Subject: {[{.Subject}]}
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

{[{.Message}]}`
}

// TranslateErrors translates all errorMessages given with the given translater.
func TranslateErrors(errorMessage *string, errorMessages map[string]string, T *translate.Translater) {
	t := T.T(*errorMessage)
	*errorMessage = t
	for k, m := range errorMessages {
		t = T.T(m)
		errorMessages[k] = t
	}
	return
}

// GetDeviceIds gets an array with deviceIds from either a single "device" or a list of "devices" in a http-requests FormValues.
func GetDeviceIds(r *http.Request) (deviceIds []interface{}, err error) {

	if r.FormValue("device") != "" {
		var deviceId int
		deviceId, err = strconv.Atoi(r.FormValue("device"))
		if err != nil {
			err = errors.New("Could not parse device Id")
			return
		}
		deviceIds = append(deviceIds, int64(deviceId))
	} else if r.FormValue("devices") != "" {
		for _, device := range strings.Split(r.FormValue("devices"), ",") {
			var deviceId int
			deviceId, err = strconv.Atoi(device)
			if err != nil {
				err = errors.New("Could not parse device Id")
				return
			}
			deviceIds = append(deviceIds, int64(deviceId))
		}
	}
	return
}

/*
// modification of smtp.SendMail from https://groups.google.com/forum/#!topic/golang-nuts/W95PXq99uns
func SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		}
		if err = c.StartTLS(config); err != nil {
			return
		}
		c.conn = tls.Client(c.conn, cfg)
		c.Text = textproto.NewConn(c.conn)
		c.tls = true
	}
	if a != nil && c.ext != nil {
		if _, ok := c.ext["AUTH"]; ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
}*/
