package main_test

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	pretty "github.com/tonnerre/golang-pretty"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/tools"
	"github.com/OpenDriversLog/webfw"

	"github.com/OpenDriversLog/goodl-lib/jsonapi/deviceManager"
	conf "github.com/OpenDriversLog/goodl/config"
	"github.com/OpenDriversLog/goodl/utils/userManager"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/driverManager"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/addressManager"
	"github.com/OpenDriversLog/goodl-lib/jsonapi/carManager"
	"github.com/OpenDriversLog/goodl-lib/models/SQLite"
)

const tTAG = dbg.Tag("goodl/t/acpt/suite.go")

var language string // {de|en}
var browser string  //{android|chrome|firefox|htmlunit|internet explorer|iPhone|iPad|opera|safari}
var agoutiDriver *agouti.WebDriver
var Config *webfw.ServerConfig
var ServerUrl string
var SeleniumUrl string
var SeleniumOpts agouti.Option
var LoginUrl string
var AdminUserEmail = "admin@admin.de"
var AdminUserPassword = "CompuIstBlödWeilErFremdePasswörterNichtLesenTut!!!1elfzwölfAdmin"
var AdminUserFirstName = "Admin"
var AdminUserSecondName = "AdminSecond"
var AdminUserTitel = "AdminSecond"
var TestUserEmail = "test@opendriverslog.de"
var RegisterUserEmail = "register_test@opendriverslog.de"
var TestUserPassword = "CompuIstBlödWeilErFremdePasswörterNichtLesenTut!!!1elfzwölf"
var TestUserFirstName = "Rainer"
var TestUserSecondName = "Zufall"
var TestUserTitel = "Seine Lordschaft"
var DefaultWaitForRenderTime = 120.0 * time.Second

func init() {

	args := os.Args[1:]
	pretty.Print("those are args...", args)

	flag.StringVar(&browser, "odl.browser", "chrome", `set desired browser for tests {chrome|firefox} \n
 more browsers should be supported in the future`)
	flag.StringVar(&language, "odl.lang", "de", `set desired language for tests {de|en}`)

	flag.Parse()

	seleniumHubAddress := os.Getenv("SELHUB_PORT_4444_TCP_ADDR") + ":4444"
	SeleniumUrl = "http://" + seleniumHubAddress + "/wd/hub"

	// https://code.google.com/p/selenium/wiki/DesiredCapabilities
	SeleniumOpts = agouti.Desired(agouti.NewCapabilities().
		Browser(browser).
		With("javascriptEnabled").
		With("applicationCacheEnabled"),
	)
	// Possibilities to consider:
	// https://sites.google.com/a/chromium.org/chromedriver/capabilities
	// https://sites.google.com/a/chromium.org/chromedriver/mobile-emulation

	// pretty.Print("F L A A A A G S  given", language, browser)
}

func TestGoodl(t *testing.T) {
	// sql.Register("SQLITE", &sqlite3.SQLiteDriver{})
	tools.RegisterSqlite("SQLITE")
	RegisterFailHandler(Fail)

	//clean image dir before generating new ones
	files, _ := ioutil.ReadDir("./img/")
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".png") && strings.Contains(f.Name(), "."+language+"."+browser) {
			os.Remove("./img/" + f.Name())
		}
	}

	RunSpecs(t, "Goodl Suite")
}

// http://godoc.org/github.com/onsi/ginkgo#SynchronizedBeforeSuite
// var _ = SynchronizedBeforeSuite(func, func)

var _ = BeforeSuite(func() {

	os.Setenv("ENVIRONMENT", "test")

	Config = conf.GetConfig()

	// TODO: FS: we could set differnt pathes to `webfw.SharedDir` to different userDb's here,
	// for concurrent testing in multiple browser nodes (chrome, ff, android, chromium w/ mobile resolution)
	// @see config/config.go ~L51 does that already
	// this would not work if we want to run tests in _goodl-dev_ container, only in _testing_
	// because in _dev_ the server already runs with the default config.
	// Therefore, only one browser could be targeted in _dev_, multiple in _testing_
	// anyways, we do not need differnt userDb files, only multiple users. their trackRecords dbs are created
	// on user creation, and we test data upload/procession on their own DBs.

	// fuuuu, user_register_test heavily works on userDb, either we multiple DBs or give them browser/lang based unique data
	_, filename, _, _ := runtime.Caller(1)
	Config.RootDir = path.Dir(filename)
	webfw.SetConfig(Config)

	// find IP of the container running our server
	addrs, err := net.InterfaceAddrs()
	var myContainerIP string
	if err != nil {
		pretty.Print("failed to get Ips", err)
	}
	for _, v := range addrs {
		pretty.Print("those are , are my IPs: %s", v.String(), v.Network())
		if strings.HasPrefix(v.String(), "172.17") {
			myContainerIP = strings.Split(v.String(), "/")[0]
		}
	}

	ServerUrl = "http://" + myContainerIP + ":4000" + Config.SubDir + "/" + language
	LoginUrl = ServerUrl + "/odl/login"

	createTestAdminUser()
	createTestUser(TestUserEmail)
	createTestUser(RegisterUserEmail)
})

var _ = AfterSuite(func() {
	// Expect(agoutiDriver.Stop()).To(Succeed())
	//deleteTestAdminUser()
	//deleteTestUser()
})

func deleteTestAdminUser() bool {
	userManager.DeleteUserFromDb(AdminUserEmail)
	return true
}
func deleteTestUser() bool {
	userManager.DeleteUserFromDb(TestUserEmail)
	userManager.DeleteUserFromDb(RegisterUserEmail)
	userManager.DeleteUserFromDb("tester_changedMail@opendriverslog.de")
	return true
}

func createTestAdminUser() bool {
	userManager.DeleteUserFromDb(AdminUserEmail)
	var admin_usr = &userManager.OdlUser{}
	admin_usr.SetFirstName(AdminUserFirstName)
	admin_usr.SetLastName(AdminUserSecondName)
	admin_usr.SetLevel(1337)
	admin_usr.SetTitle(AdminUserTitel)
	admin_usr.SetEmail(AdminUserEmail)
	admin_usr.SetLoginName(AdminUserEmail)
	admin_usr.SetTutorialDisabled(1)
	passwd := AdminUserPassword

	passwdByte, err := bcrypt.GenerateFromPassword([]byte(passwd), 13)

	if err != nil {
		dbg.E(tTAG, " bcrypting failed :( : ", err)
		return false
	}

	passwd = string(passwdByte)
	admin_usr.SetPwHash(passwd)
	userManager.CreateNewUser(admin_usr)

	return true
}

func createTestUser(givenEmail string) bool {
	userManager.DeleteUserFromDb(givenEmail)
	var test_usr = &userManager.OdlUser{}
	test_usr.SetFirstName(TestUserFirstName)
	test_usr.SetLastName(TestUserSecondName)
	test_usr.SetLevel(0)
	test_usr.SetTitle(TestUserTitel)
	test_usr.SetEmail(givenEmail)
	test_usr.SetLoginName(TestUserEmail)
	test_usr.SetTutorialDisabled(1)
	passwd := TestUserPassword

	passwdByte, err := bcrypt.GenerateFromPassword([]byte(passwd), 13)

	if err != nil {
		dbg.E(tTAG, " bcrypting failed :( : ", err)
		return false
	}

	passwd = string(passwdByte)
	test_usr.SetPwHash(passwd)
	userManager.CreateNewUser(test_usr)

	u, _ := userManager.GetUserFromDb(TestUserEmail)
	dbCon, _ := userManager.GetLocationDb(u.Id())
	testDriver := &driverManager.Driver{
		Priority: 1,
		Address : addressManager.Address{
			Street : "Chemnitzer Straße",
			HouseNumber : "89",
			Postal : "09599",
			City : "Freiberg",
		},
		Name : "Test Driver 1",
		Additional : "Coding Monkey",
	}
	var id int64
	id,err = driverManager.CreateDriver(testDriver, dbCon)
	if err != nil {
		dbg.E(tTAG, " Driver creation failed :( : ", err)
		return false
	}
	testDriver.Id = models.NInt64(id)
	/*
	Id           S.NInt64
	Type         S.NString
	Owner        driverManager.Driver
	Plate        S.NString
	FirstMileage S.NInt64
	Mileage      S.NInt64
	FirstUseDate S.NInt64
	 */
	/*
	Id          S.NInt64
	Description S.NString
	Color		*colorManager.Color
	Checked     S.NInt64
	CarId 		S.NInt64
	 */
	testCar := &carManager.Car{
		Type : "Porsche Panamera",
		Owner: *testDriver,
		Plate: "FG-CS 627",
	}

	id, err = carManager.CreateCar(testCar,dbCon)

	if err != nil {
		dbg.E(tTAG, " Car creation failed :( : ", err)
		return false
	}
	testCar.Id = models.NInt64(id)
	testCar2 := &carManager.Car{
		Type : "Opel Omega",
		Owner: *testDriver,
		Plate: "FG-JG 626",
	}
	id, err = carManager.CreateCar(testCar2,dbCon)
	if err != nil {
		dbg.E(tTAG, " Car2 creation failed :( : ", err)
		return false
	}
	testCar2.Id = models.NInt64(id)
	_, err = deviceManager.CreateDevice(&deviceManager.Device{
		Description: "TestDevice",
		Checked: 1,
		CarId: testCar.Id,
	}, dbCon)
	if err != nil {
		dbg.E(tTAG, " Device creation failed :( : ", err)
		return false
	}
	_,err = deviceManager.CreateDevice(&deviceManager.Device{
		Description: "TestDevice2",
		Checked: 1,
		CarId: testCar2.Id,
	}, dbCon)
	if err != nil {
		dbg.E(tTAG, " Device2 creation failed :( : ", err)
		return false
	}
	return true
}

func isLoggedIn(page *agouti.Page) bool {
	//var count, _ = page.Find("#nav_logout").Count()
	WaitForRender(page, 60*time.Second)
	var number int
	page.RunScript("return userData.Id;", map[string]interface{}{"number": 100}, &number)
	fmt.Println(number)
	if number != 0 {
		//dbg.D("isLoggedIn", "true %v", number)
		return true
	}
	//dbg.D("isLoggedIn", "false %v", number)
	return false

}

var AdminMode = false
var RegisterMode = false

func RequireAdminLogin(page *agouti.Page, pageUrl string) bool {
	AdminMode = true
	RequireLogin(page, pageUrl)
	AdminMode = false
	return true //not pretty
}

func RequireRegisterLogin(page *agouti.Page, pageUrl string) bool {
	RegisterMode = true
	RequireLogin(page, pageUrl)
	RegisterMode = false
	return true //not pretty
}

func NavigateAndRender(page *agouti.Page, url string, waitTime ...time.Duration) error {
	Expect(page.Navigate(url)).To(Succeed())
	dbg.D("NavigateAndRender", "Called with url : %v", url)

	Expect(WaitForRender(page, waitTime...)).To(Succeed())
	return nil
}

func WaitForRender(page *agouti.Page, waitTime ...time.Duration) error {

	wt := DefaultWaitForRenderTime
	if len(waitTime) != 0 {
		wt = waitTime[0]
	}

	i := 0
	Eventually(func() error {
		i++
		el := page.Find("#loadFinished")
		cnt, err := el.Count()
		if(i%20==0) {
			page.Screenshot("./img/currentLoading.png")
		}
		if err != nil {
			return err
		}
		if cnt == 0 {
			return errors.New("loadFinished not found.")
		}

		var loading bool
		page.RunScript("return !odl || odl==undefined || odl.loading;", map[string]interface{}{"boolean": false}, &loading)
		if loading {
			return errors.New("odl is not defined or odl.loading is true")

		}

		return nil
	}, wt).Should(Succeed())
	return nil
}

func RequireLogin(page *agouti.Page, pageUrl string) bool {
	if isLoggedIn(page) == false {

		Expect(NavigateAndRender(page, LoginUrl)).To(Succeed())
		Expect(page.Screenshot("./img/user_loginpage." + language + "." + browser + ".png")).To(Succeed())

		userField := page.Find("#login_email input")
		passwordField := page.Find("#login_password input")
		time.Sleep(500 * time.Millisecond)

		if AdminMode {
			Expect(userField.Fill(AdminUserEmail)).Should(Succeed())
			Expect(passwordField.Fill(AdminUserPassword)).Should(Succeed())
		} else if RegisterMode {
			Expect(userField.Fill(RegisterUserEmail)).Should(Succeed())
			Expect(passwordField.Fill(TestUserPassword)).Should(Succeed())
		} else {
			Expect(userField.Fill(TestUserEmail)).Should(Succeed())
			Expect(passwordField.Fill(TestUserPassword)).Should(Succeed())
		}
		time.Sleep(300 * time.Millisecond)
		Expect(page.Find("#login_submit").Click()).To(Succeed())
		//Expect(page.Screenshot("./img/RequireLoginPostClickInsta." + language + "." + browser + ".png")).To(Succeed())

		dbg.D("TestsRequireLogin", "submitted login form.", AdminMode)
		// Expect(page.Screenshot("./img/RequireLoginPostClickWait." + language + "." + browser + ".png")).To(Succeed())

		WaitForRender(page)
		time.Sleep(1000 * time.Millisecond)
		var nowLoggedIn = isLoggedIn(page)
		if nowLoggedIn == false {
			if AdminMode {
				Expect(page.Screenshot("./img/user_loginpage_failed_AdminMode." + language + "." + browser + ".png")).To(Succeed())
			} else if RegisterMode {
				Expect(page.Screenshot("./img/user_loginpage_failed_RegisterMode." + language + "." + browser + ".png")).To(Succeed())
			} else {
				Expect(page.Screenshot("./img/user_loginpage_failed_NormalMode." + language + "." + browser + ".png")).To(Succeed())
			}
			dbg.E("TestsRequireLogin", "login failed. ", AdminMode)
			return false
		} else {
			dbg.D("TestsRequireLogin", "sucessfully finished login", AdminMode)
		}
		Expect(nowLoggedIn).To(BeTrue())

		//now navigating to target page
		Expect(page.Navigate(pageUrl)).To(Succeed())
		Expect(page).To(HaveURL(pageUrl))
		return true
	}
	return true
}

func hasUnignoreableJavascriptErrors(page *agouti.Page) bool {
	//PP apparently console.log and console.error both log with the "INFO" log level.
	//So only real JavaScript errors are matched by following line
	hasRealErrors := false

	logs, _ := page.ReadAllLogs("browser")
	fmt.Println(" Found javascript errors/warnings/info Total: ", len(logs))
	// TODO: PP: reenable when polymer javascript logging is cleaned up
	for index, log := range logs {
		//we ignore silly jquery warnings
		if strings.Contains(log.Location, "jquery") {
			continue
		}
		//offline mode
		if strings.Contains(log.Message, "ERR_NAME_RESOLUTION_FAILED") {
			continue
		}
		//offline mode
		if strings.Contains(log.Message, "favicon.ico") {
			continue
		}

		//we ignore warnings for now
		if log.Level == "WARNING" {
			continue
		}

		//we ignore infos for now
		if log.Level == "INFO" {
			continue
		}
		// workaround for an issue.
		if strings.Contains(log.Message, "de/betaMan?action=read") {
			continue
		}

		pretty.Print("Log:  ", log)
		fmt.Print(index)
		fmt.Println(" Found unignoreable JavaScript Error: ")
		fmt.Print(log)
		fmt.Printf("%+v", log)
		hasRealErrors = true
	}

	return hasRealErrors
}
