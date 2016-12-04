package main_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	pretty "github.com/tonnerre/golang-pretty"

	"github.com/OpenDriversLog/goodl/utils/userManager"
)

var _ = Describe("user_register_test", func() {
	var page *agouti.Page

	BeforeEach(func() {
		var err error
		page, err = agouti.NewPage(SeleniumUrl, SeleniumOpts)

		Expect(page.Size(1280, 800)).To(Succeed())
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
	})

	It("should be able to register a user", func() {

		By("should not have RegisterUserEmail in database", func() {
			var err = userManager.DeleteUserFromDb(RegisterUserEmail)
			Expect(err).NotTo(HaveOccurred())
			err = userManager.DeleteUserFromDb("tester_changedMail@opendriverslog.de")
			Expect(err).NotTo(HaveOccurred())
		})

		By("should be able to get inviteKey", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)
			Expect(page.Navigate(ServerUrl + "/odl/login")).To(Succeed())
			Expect(RequireAdminLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
			Expect(NavigateAndRender(page, ServerUrl+"/odl/invite")).To(Succeed())
			time.Sleep(300 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test." + language + "." + browser + ".invitekey.png")).To(Succeed())
		})

		inviteKeyLink, _ := page.Find("#inviteKey").Text()
		inviteKey := strings.Replace(inviteKeyLink, Config.WebUrl+"/de", "", 10)

		pretty.Print("testUser InviteKey: \n", inviteKey)

		By("should be logging out as an admin", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)

			Expect(page.Navigate(ServerUrl + "/odl/login")).To(Succeed())
		})

		By("invitekey should not be empty", func() {
			Expect(len(strings.TrimSpace(inviteKey))).Should(BeNumerically(">=", 0))
		})

		By("should redirect to login page", func() {
			Expect(page.Navigate(ServerUrl + "/odl/lhm/trips/list/-1")).To(Succeed())
			time.Sleep(3 * time.Second)
			Expect(page).To(HaveURL(ServerUrl + "/odl/login"))
		})

		By("invitekey should be available for register", func() {
			Expect(NavigateAndRender(page, ServerUrl+inviteKey)).To(Succeed())
			Expect(page.Screenshot("./img/user_register_test." + language + "." + browser + ".userWithInvitekeyPage.png")).To(Succeed())
			//Expect(page.Find("#login_register").Click()).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + inviteKey))

		})

		emailField := page.Find("#register #uEdit #mail input")
		passwdField := page.Find("#register #uEdit #password input")
		passwd2Field := page.Find("#register #uEdit #password2 input")
		firstNameField := page.Find("#register #uEdit #firstName input")
		lastNameField := page.Find("#register #uEdit #lastName input")
		titleField := page.Find("#register #uEdit #title")
		//secondOptionText, _ := titleField.All("option").At(1).Text()

		By("should offer at least two options in title field", func() {
			Expect(titleField.All("paper-item").Count()).Should(BeNumerically(">=", 2))
		})

		By("should accept all input fields to be filled or used", func() {

			Expect(NavigateAndRender(page, ServerUrl+inviteKey)).To(Succeed())
			pretty.Print("testUser InviteKey: \n", ServerUrl+inviteKey)

			Expect(emailField.Fill(RegisterUserEmail)).Should(Succeed())
			Expect(passwdField.Fill(TestUserPassword)).Should(Succeed())
			Expect(passwd2Field.Fill(TestUserPassword)).Should(Succeed())
			Expect(firstNameField.Fill(TestUserFirstName)).Should(Succeed())
			Expect(lastNameField.Fill(TestUserSecondName)).Should(Succeed())

			//Selecting the first Option by it's Text requires to optain text
			//Expect(titleField.Select(secondOptionText)).Should(Succeed())

		})

		By("should be reject incorrect passwordvalues to accept user data values", func() {

			Expect(passwd2Field.Fill(TestUserPassword + "42")).Should(Succeed())
			time.Sleep(300 * time.Millisecond)
			Expect(page.Find("#register #uEdit #saveButton").Click()).To(Succeed())
			time.Sleep(300 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_filled_incorrect." + language + "." + browser + ".png")).To(Succeed())
			Expect(page.Find("#error")).To(BeVisible())
			Expect(page.Find("#error #closeHelpButton").Click()).To(Succeed())
			time.Sleep(100 * time.Millisecond)
		})

		By("should reject too short passwords", func() {
			Expect(passwdField.Fill("Pen1sCS")).Should(Succeed())
			Expect(passwd2Field.Fill("Pen1sCS")).Should(Succeed())
			time.Sleep(100 * time.Millisecond)
			Expect(page.Find("#register #uEdit #saveButton").Click()).To(Succeed())
			time.Sleep(300 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_filled." + language + "." + browser + "_too_short.png")).To(Succeed())
			Expect(page.Find("#error")).To(BeVisible())
			Expect(page.Find("#error #closeHelpButton").Click()).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			//DIDN'T work Eventually(page.Find(".warning").Visible()).Should(BeTrue())
		})

		By("should reject passwords without numbers", func() {
			Expect(passwdField.Fill("PeeeeeeeeeeeeeeeenisPP")).Should(Succeed())
			Expect(passwd2Field.Fill("PeeeeeeeeeeeeeeeenisPP")).Should(Succeed())
			time.Sleep(100 * time.Millisecond)
			Expect(page.Find("#register #uEdit #saveButton").Click()).To(Succeed())
			time.Sleep(300 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_filled." + language + "." + browser + "_without_numbers.png")).To(Succeed())
			Expect(page.Find("#error")).To(BeVisible())
			Expect(page.Find("#error #closeHelpButton").Click()).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			//DIDN'T work Eventually(page.Find(".warning").Visible()).Should(BeTrue())
		})

		By("should be able to accept user data values", func() {

			//PP failed to increase config by testUser and testPassword, pls edit
			Expect(emailField.Fill(RegisterUserEmail)).Should(Succeed())
			Expect(passwdField.Fill(TestUserPassword)).Should(Succeed())
			Expect(passwd2Field.Fill(TestUserPassword)).Should(Succeed())
			Expect(firstNameField.Fill(TestUserFirstName)).Should(Succeed())
			Expect(lastNameField.Fill(TestUserSecondName)).Should(Succeed())
			//TODO Expect(titleField.Fill(TestUserTitel)).Should(Succeed())

			Expect(page.Screenshot("./img/user_register_test_filled." + language + "." + browser + ".png")).To(Succeed())
		})
		By("should be able to delete old user", func() {

			err := userManager.DeleteUserFromDb(RegisterUserEmail)
			time.Sleep(1000 * time.Millisecond)
			Expect(err).NotTo(HaveOccurred())
		})
		By("should be able to save new user", func() {
			Expect(page.Find("#register #uEdit #saveButton").Click()).To(Succeed())

			time.Sleep(1000 * time.Millisecond)
			Expect(WaitForRender(page)).To(Succeed())
			time.Sleep(300 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_filled." + language + "." + browser + "_submitted.png")).To(Succeed())

			Expect(page.Find(".error").Visible()).To(BeFalse())
			Expect(page.Find(".warning").Visible()).To(BeFalse())
			Expect(page.Find("#toast").Visible()).To(BeTrue())
		})

		var testUser, err = userManager.GetUserFromDb(RegisterUserEmail)

		By("should have created testuser", func() {
			pretty.Print("testUser: \n  ", err)
			Expect(err).NotTo(HaveOccurred())
			pretty.Print("testUser: \n  ", testUser)
			pretty.Print("IActivationKey: \n  ", testUser.ActivationKey())
		})

		By("should be able to activate new user", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/register?mail="+RegisterUserEmail+"&activateKey="+testUser.ActivationKey())).To(Succeed())
			Expect(page.Screenshot("./img/user_register_test_activation." + language + "." + browser + "_done.png")).To(Succeed())
			//Expect(page.Find(".statusMessage").Visible()).To(BeTrue())
		})

		//TODO implement!
		//By("should login User after sucessfull registration", func() {
		//	Expect(isLoggedIn(page)).To(BeTrue())
		//})

		//TODO implement!
		//By("should redirect User after sucessfull registration", func() {
		//	Expect(page).To(HaveURL(ServerUrl + "/odl/login/"))
		//})

		By("redirecting the user to the login form from the home page", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/odl/login")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/login"))
		})

		By("should have visible navigation", func() {
			time.Sleep(1 * time.Second)
			Expect(page.Find("#side-nav")).To(BeVisible())
		})

		By("should have working login", func() {
			if isLoggedIn(page) == true { //shouldn't be the case
				Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
				time.Sleep(3000 * time.Millisecond)
			}
			Expect(RequireRegisterLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
		})

		By("should make lhm available", func() {
			Expect(page.Navigate(ServerUrl + "/odl/lhm/trips/list/-1")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/lhm/trips/list/-1"))
		})

		By("should make settings available", func() {
			Expect(page.Navigate(ServerUrl + "/odl/settings")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/settings"))
		})

		By("Should be able to disable tutorial", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/odl/settings")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/settings"))
			time.Sleep(1000 * time.Millisecond)
			page.Find("#status #closeHelpButton").Click()
			time.Sleep(500 * time.Millisecond)
			tutorialEnableCB := page.Find("#tutorialEnableCB")
			Expect(tutorialEnableCB.Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_disableTutorial." + language + "." + browser + "_beforeSave.png")).To(Succeed())
			Expect(page.Find("#uEdit #saveButton").Click()).To(Succeed())
			time.Sleep(1000 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_disableTutorial." + language + "." + browser + "_finished.png")).To(Succeed())
			Expect(page.Find("#tutorialEnableCB")).NotTo(BeSelected())
		})
		By("should be able to create driver from upload page")
		{
			Expect(NavigateAndRender(page, ServerUrl+"/odl/upload/upload")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/upload/upload"))
			upload_key := page.Find("#upload_key iron-icon")
			Expect(upload_key.Click()).To(Succeed())
			time.Sleep(1000 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_deviceDropDown.png")).To(Succeed())
			newDeviceButton := page.Find("#upload_key paper-item[device-id=\"0\"]")
			Expect(newDeviceButton.Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_editDevice_empty.png")).To(Succeed())

			Expect(page).To(HaveURL(ServerUrl + "/odl/upload/editDevice"))
			Expect(page.Find("#upload #editDevice #formEditDevice car-selector iron-icon").Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			newCarButton := page.Find("#upload #editDevice #formEditDevice car-selector paper-item[car-id=\"0\"]")
			Expect(newCarButton.Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_editCar_empty.png")).To(Succeed())

			Expect(page.Find("#upload #editDevice #formEditCar driver-selector iron-icon").Click()).To(Succeed())
			time.Sleep(1000 * time.Millisecond)
			newDriverButton := page.Find("#upload #editDevice #formEditCar driver-selector paper-item[driver-id=\"0\"]")
			Expect(newDriverButton.Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_editDriver_empty.png")).To(Succeed())
			Expect(page.Find("#upload #editDevice #formEditDriver #driverName input").Fill("Christoph Sonntag")).To(Succeed())
			Expect(page.Find("#upload #editDevice #formEditDriver #driverAdditional input").Fill("Bester Fahrer der Welt")).To(Succeed())
			Expect(page.Find("#upload #editDevice #formEditDriver #adrEdit #address_edit_postal input").Fill("09599")).To(Succeed())
			Expect(page.Find("#upload #editDevice #formEditDriver #adrEdit #address_edit_city input").Fill("Freiberg")).To(Succeed())
			Expect(page.Find("#upload #editDevice #formEditDriver #adrEdit #address_edit_street input").Fill("Chemnitzer Straße")).To(Succeed())
			Expect(page.Find("#upload #editDevice #formEditDriver #adrEdit #address_edit_houseNumber input").Fill("89")).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_editDriver_filled.png")).To(Succeed())

			Expect(page.Find("#upload #editDevice #driverEdit #saveButton").Click()).To(Succeed())
			time.Sleep(1000 * time.Millisecond)
			Expect(WaitForRender(page)).To(Succeed())
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_editCar_newDriver.png")).To(Succeed())
			var driver string
			Expect(page.RunScript("return $(\"#upload #editDevice #formEditCar driver-selector paper-dropdown-menu\")[0].selectedItemLabel", nil, &driver)).To(Succeed())
			Expect(driver).To(Equal("Christoph Sonntag"))
		}
		By("should be able to create car from upload page")
		{
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_editCar_newOwner.png")).To(Succeed())
			Expect(page.Find("#upload #editDevice #formEditCar #car_edit_plate input").Fill("FG-CS 629")).To(Succeed())
			Expect(page.Find("#upload #editDevice #formEditCar #car_edit_type input").Fill("VW Passat CC")).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_editCar_filled.png")).To(Succeed())

			Expect(page.First("#upload #editDevice #carEdit #saveButton").Click()).To(Succeed())
			time.Sleep(1000 * time.Millisecond)

			Expect(WaitForRender(page)).To(Succeed())
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_editDevice_newCar.png")).To(Succeed())
			var car string
			Expect(page.RunScript("return $(\"#upload #editDevice #formEditDevice car-selector paper-dropdown-menu\")[0].selectedItemLabel", nil, &car)).To(Succeed())
			Expect(car).To(ContainSubstring("FG-CS 629"))
			Expect(car).To(ContainSubstring("VW Passat CC"))
		}
		By("should be able to create device from upload page")
		{
			Expect(page.Find("#upload #editDevice #formEditDevice #deviceDesc input").Fill("NMEATest")).To(Succeed())
			time.Sleep(500 * time.Millisecond)

			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_filled.png")).To(Succeed())
			Expect(page.First("#upload #editDevice #deviceEdit #saveButton").Click()).To(Succeed())
			time.Sleep(1000 * time.Millisecond)
			Expect(WaitForRender(page)).To(Succeed())
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_newDevice.png")).To(Succeed())
			var device string
			Expect(page.RunScript("return $(\"#upload device-selector paper-dropdown-menu\")[0].selectedItemLabel", nil, &device)).To(Succeed())
			Expect(device).To(Equal("NMEATest"))

		}
		By("should accept uploaded NMEA data", func() {

			upload_externalData := page.Find("#upload_externalData")

			buf := bytes.NewBuffer(nil)
			f, err := os.Open("./files/gps.log")
			Expect(err).NotTo(HaveOccurred())
			io.Copy(buf, f)
			f.Close()
			gpsData := string(buf.Bytes())
			pms := make(map[string]interface{})
			pms["gpsData"] = string(gpsData)

			Expect(upload_externalData.Fill("Crap")).Should(Succeed())
			page.RunScript("document.getElementById('upload_externalData').value=gpsData;", pms, nil)

			time.Sleep(300 * time.Millisecond)
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_filled.png")).To(Succeed())
			Expect(page.Find("#submit").Click()).To(Succeed())
			Expect(page.Screenshot("./img/user_register_test_upload." + language + "." + browser + "_submitted.png")).To(Succeed())

		})

		By("should have data in lhm after upload", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/odl/lhm/trips/list/-1")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/lhm/trips/list/-1"))
			time.Sleep(5 * time.Second)
			Expect(page.Screenshot("./img/user_register_test_lhm_with_data." + language + "." + browser + ".png")).To(Succeed())
		})

	})

})
