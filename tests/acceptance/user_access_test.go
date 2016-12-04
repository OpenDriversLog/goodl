package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	"io/ioutil"
	"strings"
	"time"
	//"time"
	//"github.com/Compufreak345/dbg"
)

var user_pages = []string{
	"lhm/trips/list/-1",
	"settings",
	"upload/upload",
	"welcome",
}

var admin_pages = []string{
	"betaMan/sendMail/list/-1",
	"usrMan/list/-1",
}

var _ = Describe("user_access_test", func() {
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

	It("should redirect from user pages to login when user is logged out", func() {

		By("should have working logout", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)
		})

		for _, element := range user_pages {

			By("redirect from "+element+" to login if logged out", func() {
				Expect(page.Navigate(ServerUrl + "/odl/" + element)).To(Succeed())
				Expect(WaitForRender(page)).To(Succeed())
				time.Sleep(3 * time.Second)
				Expect(page).To(HaveURL(ServerUrl + "/odl/login"))
			})

		}

	})

	It("should be able to show user pages when logged in", func() {

		By("should have working login", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)
			Expect(RequireLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
		})

		for _, element := range user_pages {

			By("allow access to "+element+" when logged in", func() {

				Expect(NavigateAndRender(page, ServerUrl+"/odl/"+element)).To(Succeed())
				Expect(page).To(HaveURL(ServerUrl + "/odl/" + element))
				// it is a user page, therefore there should be the user nav!
				// Expect(page.Find("#nav_welcome")).To(Succeed())
				// Expect(page.Find("#nav_welcome")).To(BeFound())

				//check for javascript errors
				Expect(hasUnignoreableJavascriptErrors(page)).To(BeFalse())

				Expect(page.Screenshot("./img/user_access_test." + language + "." + browser + "_logged_in_" + strings.Replace(element, "/", "_", -1) + ".png")).To(Succeed())
			})

		}

	})

	It("should be not show admin pages to users even if user is logged in", func() {

		By("should have working user login", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)
			Expect(RequireLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
		})

		for _, element := range admin_pages {

			By("deny access to "+element+" if user level to low", func() {

				Expect(page.Navigate(ServerUrl + "/odl/" + element)).To(Succeed())
				Expect(page).To(HaveURL(ServerUrl + "/odl/" + element))
				Expect(page.Screenshot("./img/user_access_test." + language + "." + browser + "_logged_in_user_" + strings.Replace(element, "/", "_", -1) + ".png")).To(Succeed())
			})

		}

	})

	It("should allow access to static files if user is logged in", func() {

		By("should have working user login", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)
			Expect(RequireLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
		})

		static_files, _ := ioutil.ReadDir(Config.RootDir + "/views/static/")

		for _, element := range static_files {

			By("deny access to "+element.Name()+" if user level to low", func() {

				Expect(page.Navigate(ServerUrl + "/odl/static/" + element.Name())).To(Succeed())
				Expect(page).To(HaveURL(ServerUrl + "/odl/static/" + element.Name()))
				Expect(page.Screenshot("./img/user_access_test." + language + "." + browser + "_logged_in_user_" + element.Name() + ".png")).To(Succeed())
			})

		}

	})

	It("should be able show admin pages to admins if logged in", func() {

		By("should have working admin login", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)
			Expect(RequireAdminLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
		})

		for _, element := range admin_pages {
			By("allow access to admin page "+element+" if user_level is admin", func() {
				Expect(NavigateAndRender(page, ServerUrl+"/odl/"+element)).To(Succeed())
				Expect(page).To(HaveURL(ServerUrl + "/odl/" + element))

				//it is an admin page, therefore there should be the admin nav!
				// Expect(page.First("#nav_usrMan")).To(BeFound())
				// Expect(page.First("#nav_betaMan")).To(BeFound())

				//check for javascript errors
				Expect(hasUnignoreableJavascriptErrors(page)).To(BeFalse())

				Expect(page.Screenshot("./img/user_access_test." + language + "." + browser + "_logged_in_admin_" + strings.Replace(element, "/", "_", -1) + ".png")).To(Succeed())
			})

		}

		By("should have working admin logout", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)
			Expect(RequireAdminLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
		})

	})

})
