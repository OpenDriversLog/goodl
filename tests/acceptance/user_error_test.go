package main_test

import (
	//"bytes"
	//"io"
	//"os"
	//"strings"
	//"golang.org/x/crypto/bcrypt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	"time"
)

var _ = Describe("user_error_test", func() {
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

	It("should be able to login and navigate to welcome page", func() {

		By("should have working login", func() {
			Expect(page.Navigate(ServerUrl + "/odl/welcome")).To(Succeed())
			Expect(RequireRegisterLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
		})
	})

	It("should be able to report an error", func() {

		By("should be able to access settings page", func() {
			Expect(page.Navigate(ServerUrl + "/odl/welcome")).To(Succeed())
			Expect(RequireLogin(page, ServerUrl+"/odl/welcome")).To(BeTrue())
			Expect(WaitForRender(page)).To(Succeed())
			Expect(page.Find("#reportErrorDialog").Visible()).To(BeFalse(), "modal be visible expected")
		})

		By("should be able to find report error top nav button", func() {
			Expect(page.Find("#topnav_nav_report_error")).To(BeFound())
			Expect(page.Find("#topnav_nav_report_error").Visible()).To(BeTrue(), "topnav button to be visible expected")
		})

		By("should be able to find report error side nav button", func() {
			Expect(page.Find("#nav_nav_report_error")).To(BeFound())
		})

		By("should be able to click report error button", func() {
			Expect(page.Find("#topnav_nav_report_error").Click()).To(Succeed())
			Expect(page.Screenshot("./img/user_error_test_first_click." + language + "." + browser + ".png")).To(Succeed())

			time.Sleep(500 * time.Millisecond)
			Expect(page.Find("#reportErrorDialog").Visible()).To(BeTrue(), "modal be visible expected")
		})

	})

})
