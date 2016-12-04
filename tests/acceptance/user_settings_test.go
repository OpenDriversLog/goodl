package main_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	pretty "github.com/tonnerre/golang-pretty"

	. "github.com/OpenDriversLog/goodl/utils/userManager"
)

var _ = Describe("user_settings_test", func() {
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

	It("should be able to login and navigate to settings", func() {

		By("should have working login", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)
			Expect(RequireRegisterLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
		})
	})

	It("should be able to change user settings", func() {

		By("should be able to access settings page", func() {
			Expect(page.Navigate(ServerUrl + "/odl/settings")).To(Succeed())
			Expect(RequireLogin(page, ServerUrl+"/odl/settings")).To(BeTrue())
			Expect(WaitForRender(page)).To(Succeed())
		})

		By("should be able to find settings fields", func() {
			Expect(page.Find("#settings #uEdit #mail input")).To(BeFound())
			Expect(page.Find("#settings #uEdit #password input")).To(BeFound())
			Expect(page.Find("#settings #uEdit #password2 input")).To(BeFound())
			Expect(page.Find("#settings #uEdit #firstName input")).To(BeFound())
			Expect(page.Find("#settings #uEdit #lastName input")).To(BeFound())
			Expect(page.Find("#settings #uEdit #title")).To(BeFound())
		})

	})

	It("should be able to change the users first name", func() {

		By("should be able to access settings page", func() {
			Expect(page.Navigate(ServerUrl + "/odl/settings")).To(Succeed())
			Expect(RequireLogin(page, ServerUrl+"/odl/settings")).To(BeTrue())
			Expect(WaitForRender(page)).To(Succeed())
		})

		firstNameField := page.Find("#settings #uEdit #firstName input")
		initial_firstName := TestUserFirstName
		new_firstName := initial_firstName + " Maria"

		By("should be able to change first name in GUI", func() {
			pretty.Print("new_firstName: \n  ", new_firstName)
			pretty.Print("initial_firstName: \n  ", initial_firstName)
			time.Sleep(500 * time.Millisecond)
			Expect(firstNameField.Click()).To(Succeed())
			Expect(firstNameField.Fill(new_firstName)).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			Expect(page.Screenshot("./img/user_settings_test_firstName_beforeClick." + language + "." + browser + ".png")).To(Succeed())

			Expect(page.Find("#settings #uEdit #saveButton").Click()).To(Succeed())
			Eventually(func() error {
				visible, _ := page.Find("#toast").Visible()
				if !visible {
					return errors.New("toast not visible")
				}
				return nil
			}, 30000*time.Millisecond).Should(Succeed())
			time.Sleep(20 * time.Millisecond)
			Expect(page.Screenshot("./img/user_settings_test_firstName." + language + "." + browser + ".png")).To(Succeed())
			Expect(page.Find(".warning").Visible()).To(BeFalse(), "warning not expected")
			Expect(page.Find(".error").Visible()).To(BeFalse(), "error not expected")
			Expect(page.Find("#toast").Visible()).To(BeTrue(), "status expected")
		})

		By("FirstName should have changed in database", func() {
			var testUser, err = GetUserFromDb(TestUserEmail)
			Expect(err).NotTo(HaveOccurred())
			Expect(testUser.IfirstName).To(Equal(new_firstName))
		})

	})

	It("should be able to change last name", func() {
		lastNameField := page.Find("#settings #uEdit #lastName input")
		initial_lastName := TestUserSecondName
		new_lastName := "von " + initial_lastName

		By("should be able to access settings page", func() {
			Expect(page.Navigate(ServerUrl + "/odl/settings")).To(Succeed())
			Expect(RequireLogin(page, ServerUrl+"/odl/settings")).To(BeTrue())
			Expect(WaitForRender(page)).To(Succeed())

		})

		By("should be able to change last name in GUI", func() {
			time.Sleep(500 * time.Millisecond)
			Expect(lastNameField.Click()).To(Succeed())
			Expect(lastNameField.Fill(new_lastName)).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			Expect(page.Find("#settings #uEdit #saveButton").Click()).To(Succeed())
			Eventually(func() error {
				visible, _ := page.Find("#toast").Visible()
				if !visible {
					return errors.New("toast not visible")
				}
				return nil
			}, 30000*time.Millisecond).Should(Succeed())
			time.Sleep(20 * time.Millisecond)
			Expect(page.Screenshot("./img/user_settings_test_lastName." + language + "." + browser + ".png")).To(Succeed())
			Expect(page.Find(".warning").Visible()).To(BeFalse(), "warning not expected")
			Expect(page.Find(".error").Visible()).To(BeFalse(), "error not expected")
			Expect(page.Find("#toast").Visible()).To(BeTrue(), "status expected")
		})

		By("LastName should have changed in database", func() {
			var testUser, err = GetUserFromDb(TestUserEmail)
			Expect(err).NotTo(HaveOccurred())
			Expect(testUser.IlastName).To(Equal(new_lastName))
		})
	})

	It("should be able to change email", func() {

		By("should be able to access settings page", func() {
			Expect(page.Navigate(ServerUrl + "/odl/settings")).To(Succeed())
			Expect(RequireLogin(page, ServerUrl+"/odl/settings")).To(BeTrue())
			Expect(WaitForRender(page)).To(Succeed())
		})

		emailField := page.Find("#settings #uEdit #mail input")
		initial_email := TestUserEmail
		new_email := "tester_changedMail@opendriverslog.de"

		By("should be able to change email adress in GUI", func() {
			time.Sleep(500 * time.Millisecond)
			Expect(emailField.Click()).To(Succeed())
			Expect(emailField.Fill(new_email)).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Find("#settings #uEdit #saveButton").Click()).To(Succeed())
			Expect(page.Screenshot("./img/user_settings_test_email1." + language + "." + browser + ".png")).To(Succeed())

			Eventually(func() error {
				visible, _ := page.Find("#toast").Visible()
				if !visible {
					Expect(page.Screenshot("./img/user_settings_test_email_back_no_toast." + language + "." + browser + ".png")).To(Succeed())
					return errors.New("toast not visible")

				}
				return nil
			}, 30000*time.Millisecond).Should(Succeed())
			time.Sleep(20 * time.Millisecond)
			Expect(page.Screenshot("./img/user_settings_test_email2." + language + "." + browser + ".png")).To(Succeed())
			Expect(page.Find(".warning").Visible()).To(BeFalse(), "warning not expected")
			Expect(page.Find(".error").Visible()).To(BeFalse(), "error not expected")
			Expect(page.Find("#toast").Visible()).To(BeTrue(), "status expected")
		})

		By("email should have changed in database", func() {
			var testUser, err = GetUserFromDb(new_email)
			pretty.Print("testUserChanged: \n  ", testUser)
			Expect(err).NotTo(HaveOccurred())
			Expect(testUser != nil).To(BeTrue())
			Expect(testUser.Iemail).To(Equal(new_email))
		})

		By("should be able to access settings page", func() {
			Expect(page.Navigate(ServerUrl + "/odl/settings")).To(Succeed())
			Expect(RequireLogin(page, ServerUrl+"/odl/settings")).To(BeTrue())
		})

		By("should be able to change email adress back", func() {

			Expect(WaitForRender(page)).To(Succeed())
			time.Sleep(500 * time.Millisecond)

			Expect(emailField.Click()).To(Succeed())
			Expect(emailField.Fill(initial_email)).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Find("#settings #uEdit #saveButton").Click()).To(Succeed())

			Eventually(func() error {
				visible, _ := page.Find("#toast").Visible()
				if !visible {
					Expect(page.Screenshot("./img/user_settings_test_email_back_no_toast." + language + "." + browser + ".png")).To(Succeed())
					return errors.New("toast not visible")

				}
				return nil
			}, 30000*time.Millisecond).Should(Succeed())
			time.Sleep(20 * time.Millisecond)
			Expect(page.Screenshot("./img/user_settings_test_email_back." + language + "." + browser + ".png")).To(Succeed())
			Expect(page.Find(".warning").Visible()).To(BeFalse(), "warning not expected")
			Expect(page.Find(".error").Visible()).To(BeFalse(), "error not expected")
			Expect(page.Find("#toast").Visible()).To(BeTrue(), "status expected")
		})

		By("email should have changed in database", func() {
			var testUser, err = GetUserFromDb(initial_email)
			pretty.Print("testUserChangedBack: \n  ", testUser)
			Expect(err).NotTo(HaveOccurred())
			Expect(testUser != nil).To(BeTrue())
			Expect(testUser.Iemail).To(Equal(initial_email))
		})

	})

	It("should be able to change users password", func() {

		By("should be able to access settings page", func() {
			Expect(page.Navigate(ServerUrl + "/odl/settings")).To(Succeed())
			Expect(RequireLogin(page, ServerUrl+"/odl/settings")).To(BeTrue())
			Expect(WaitForRender(page)).To(Succeed())
		})

		passwordField := page.Find("#settings #uEdit #password input")
		passwordField2 := page.Find("#settings #uEdit #password2 input")
		//initial_email, _ := passwordField.Text()
		new_password := TestUserPassword

		By("should be able to change user password in GUI", func() {

			time.Sleep(500 * time.Millisecond)
			Expect(passwordField.Click()).To(Succeed())
			Expect(passwordField.Fill(new_password)).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			Expect(passwordField2.Click()).To(Succeed())
			Expect(passwordField2.Fill(new_password)).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			Expect(page.Find("#settings #uEdit #saveButton").Click()).To(Succeed())
			Eventually(func() error {
				visible, _ := page.Find("#toast").Visible()
				if !visible {
					return errors.New("toast not visible")
				}
				return nil
			}, 30000*time.Millisecond).Should(Succeed())
			time.Sleep(20 * time.Millisecond)
			Expect(page.Screenshot("./img/user_settings_test_password." + language + "." + browser + ".png")).To(Succeed())
			Expect(page.Find(".warning").Visible()).To(BeFalse(), "warning not expected")
			Expect(page.Find(".error").Visible()).To(BeFalse(), "error not expected")
			Expect(page.Find("#toast").Visible()).To(BeTrue(), "status expected")
		})

		/* TODO - check if password really has changed in database
			fails at testUser.ipwHash which is not accessible?
		By("password hash should have changed in database", func() {
			var testUser, err = GetUserFromDb(RegisterUserEmail)
			Expect(err == nil).To(BeTrue())
			passwdByte, err := bcrypt.GenerateFromPassword([]byte(new_password), 13)
			pretty.Print("testUser: \n  ", testUser)
			Expect(err == nil).To(BeTrue())
			passwd := string(passwdByte)
			Expect(testUser.ipwHash).To(Equal(passwd))
		})
		*/

	})

	/*
		IloginName:         "register_test@opendriverslog.de",
		IfirstName:         "Rainer",
		IlastName:          "Zufall",
		Ititle:             "Mr.",
		Iemail:             "register_test@opendriverslog.de",
		Iid:                129,
		Iip:                "",
		lastNonce:          (*login.PrivKeyAndSalt)(nil),
		nonceWasUsed:       false,
		notFirstNonce:      false,
		IisLoggedIn:        false,
		 IopenSessionIds:    {},
		openSessionIdMutex: {},
		 ipwHash:            "$2a$13$zr9cbaYWKXPVvNiHwfnTIuCpymUQP4/VVHTq9K8QQRYbPlW5DALwm",
	*/

})
