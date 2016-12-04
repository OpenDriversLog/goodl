package main_test

import (
	//"fmt"
	//"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	//"github.com/Compufreak345/dbg"
	"strings"
)

var TestDriverName = "Rainer Zufall"
var TestDriverAdditional = "FAST"
var TestDriverCity = "Freiberg"
var TestDriverPostal = "09599"
var TestDriverPostalEdited = "12345"
var TesDriverStreet = "Moritzstraße"
var TestDriverNumber = "1"
var TestDriverEditChar = "1"

var _ = Describe("drivers_test", func() {
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

	It("should be able to create and show driver", func() {

		By("redirecting the user to the login form from the lhm if logged out", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3000 * time.Millisecond)
			Expect(page.Navigate(ServerUrl + "/odl/lhm/trips/list/-1")).To(Succeed())
			time.Sleep(3000 * time.Millisecond)
			Expect(page).To(HaveURL(ServerUrl + "/odl/login"))
		})

		By("should have working login", func() {
			Expect(RequireLogin(page, ServerUrl+"/odl/lhm/trips/list/-1")).To(BeTrue())
			Expect(page).To(HaveURL(ServerUrl + "/odl/lhm/trips/list/-1"))
			Expect(page.Screenshot("./img/drivers_test." + language + "." + browser + ".png")).To(Succeed())
		})

		By("not redirecting the user after sucessfull login", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/odl/lhm/trips/list/-1")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/lhm/trips/list/-1"))
		})

		By("should have visible navigation on mapselect", func() {
			Expect(page.Find("#side-nav")).To(BeVisible())
		})

		By("should have logout & upload link", func() {
			Expect(page.First("#nav_logout")).To(BeFound())
			Expect(page.First("#nav_upload")).To(BeFound())
		})

		By("Should be able to open drivers tab", func() {
			Expect(page.Screenshot("./img/drivers_nav_test." + language + "." + browser + ".png")).To(Succeed())
			Expect(page.Find("#nav_drivers").Click()).To(Succeed())
		})

		By("should be showing drivers list", func() {
			time.Sleep(3000 * time.Millisecond)
			Expect(page.Screenshot("./img/drivers_test." + language + "." + browser + "_default_view.png")).To(Succeed())
		})

		By("should not have broken javascript on lhm/drivers/list/1", func() {
			Expect(hasUnignoreableJavascriptErrors(page)).To(BeFalse())
		})

		By("should have newDriverBtn to click on", func() {
			Expect(page.First("#newDriverBtn")).To(BeFound())
			Expect(page.First("#newDriverBtn").Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/drivers_test." + language + "." + browser + "_newDriver_clicked.png")).To(Succeed())
		})



		By("should find inputs for driverForm", func() {
			nameField := page.Find("#drivers #driverEdit input[name='name']")
			commentField := page.Find("#drivers #driverEdit input[name='additional']")
			cityField := page.Find("#drivers #driverEdit input[name='city']")
			postalField := page.Find("#drivers #driverEdit input[name='postal']")
			streetField := page.Find("#drivers #driverEdit input[name='street']")
			numberField := page.Find("#drivers #driverEdit input[name='number']")
			Expect(nameField.Fill(TestDriverName)).Should(Succeed())
			Expect(commentField.Fill(TestDriverAdditional)).Should(Succeed())
			Expect(cityField.Fill(TestDriverCity)).Should(Succeed())
			Expect(postalField.Fill(TestDriverPostal)).Should(Succeed())
			Expect(streetField.Fill(TesDriverStreet)).Should(Succeed())
			Expect(numberField.Fill(TestDriverNumber)).Should(Succeed())

			time.Sleep(100 * time.Millisecond)
			Expect(page.Screenshot("./img/drivers_test." + language + "." + browser + "_newDriver_filled.png")).To(Succeed())
		})

		By("should be able to save", func() {
			Expect(page.First("#drivers #driverEdit #saveButton").Click()).To(Succeed())
			time.Sleep(1000 * time.Millisecond)
			Expect(WaitForRender(page)).To(Succeed())
			Expect(page.Screenshot("./img/drivers_test." + language + "." + browser + "_newDriver_submitted.png")).To(Succeed())
		})

		By("should show created driver in drivers tab", func() {
			time.Sleep(400 * time.Millisecond)
			searchField := page.Find("#drivers #searchBox input")
			Expect(searchField.Fill(TestDriverName)).To(Succeed())
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_highlighted_new_driver.png")).To(Succeed())
			Expect(page.Find("#drivers .highlighted")).To(BeFound())
			t,_err:= page.Find("#drivers .highlighted > div.title").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestDriverName))
		})



		By("should be able to navigate to edit", func() {
			id, _err := page.Find("#drivers .highlighted").Attribute("id")
			Expect(_err).To(BeNil())
			id = string(id[strings.Index(id,"_")+1:])
			editButton := page.Find("#drivers #driverIronList #items #item_"+id+" paper-icon-button[icon='icons:create']")
			Expect(editButton.Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/drivers_test." + language + "." + browser + "_clicked_edit2_driver.png")).To(Succeed())
		})

		By("should accept modification values", func() {
			nameField := page.Find("#drivers #driverEdit input[name='name']")
			commentField := page.Find("#drivers #driverEdit input[name='additional']")
			cityField := page.Find("#drivers #driverEdit input[name='city']")
			postalField := page.Find("#drivers #driverEdit input[name='postal']")
			streetField := page.Find("#drivers #driverEdit input[name='street']")
			numberField := page.Find("#drivers #driverEdit input[name='number']")
			Expect(nameField.Fill(TestDriverName + TestDriverEditChar)).Should(Succeed())
			Expect(commentField.Fill(TestDriverAdditional + TestDriverEditChar)).Should(Succeed())
			Expect(cityField.Fill(TestDriverCity + TestDriverEditChar)).Should(Succeed())
			Expect(postalField.Fill(TestDriverPostalEdited)).Should(Succeed())
			Expect(streetField.Fill(TesDriverStreet + TestDriverEditChar)).Should(Succeed())
			Expect(numberField.Fill(TestDriverNumber + TestDriverEditChar)).Should(Succeed())

			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/drivers_test." + language + "." + browser + "_editDriver_filled.png")).To(Succeed())

		})

		By("should be able to save modification", func() {
			Expect(page.First("#drivers #saveButton").Click()).To(Succeed())
			time.Sleep(1500 * time.Millisecond)
		})

		By("Should be able to search for edited driver", func() {

			time.Sleep(400 * time.Millisecond)
			searchField := page.Find("#drivers #searchBox input")
			Expect(searchField.Fill(TestDriverName+TestDriverEditChar)).To(Succeed())
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_highlighted_edited_driver.png")).To(Succeed())
			Expect(page.Find("#drivers .highlighted")).To(BeFound())
			t,_err:= page.Find("#drivers .highlighted > div.title").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestDriverName+TestDriverEditChar))
		})

		By("should be able to delete driver", func() {
			time.Sleep(100*time.Millisecond)

			id, _err := page.Find("#drivers .highlighted").Attribute("id")
			Expect(_err).To(BeNil())
			id = string(id[strings.Index(id,"_")+1:])
			deleteButton := page.Find("#drivers #driverIronList #items #item_"+id+" paper-icon-button[icon='icons:delete']")
			Expect(deleteButton.Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			time.Sleep(200 * time.Millisecond)
			Expect(page.Screenshot("./img/drivers_test." + language + "." + browser + "_delete_confirm_driver.png")).To(Succeed())
			Expect(page.Find("#deleteDialog[key='driver_"+id+"'] #delete_dialog_confirm").Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())


		})

		/*By("should not find driver anymore", func() {

			time.Sleep(3000 * time.Millisecond)
			Expect(page.Screenshot("./img/drivers_test." + language + "." + browser + "_delete_confirmed_driver.png")).To(Succeed())

			searchField := page.Find("#drivers #searchBox input")
			Expect(searchField.Fill(TestDriverName+TestDriverEditChar)).To(Succeed())
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_highlighted_deleted_driver.png")).To(Succeed())
			Expect(page.Find("#drivers .highlighted")).ToNot(BeFound())
		})*/

	})

})
