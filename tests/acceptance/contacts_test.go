package main_test

import (
	//"fmt"
	//"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	//pretty "github.com/tonnerre/golang-pretty"

	//"github.com/Compufreak345/dbg"
	"strings"
)

var TestContactTitle = "Paul!"
var TestContactDesc = "tcDesc"
var TestContactCity = "Freiberg"
var TestContactPostal = "09599"
var TestContactStreet = "Am Obermarkt"
var TestContactNumber = "24"
var TestContactTitleEdit = "Klaus!"
var TestContactDescEdit = "edDesc"
var TestContactCityEdit = "Dresden"
var TestContactPostalEdit = "01127"
var TestContactStreetEdit = "Konkordienstraße"
var TestContactNumberEdit = "27"

var _ = Describe("contacts_test", func() {
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

	It("should be able to create and show contact", func() {

		By("redirecting the user to the login form from the lhm if logged out", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3000 * time.Millisecond)
			Expect(page.Navigate(ServerUrl + "/odl/lhm/contacts/list/1")).To(Succeed())
			time.Sleep(3000 * time.Millisecond)
			Expect(page).To(HaveURL(ServerUrl + "/odl/login"))
		})

		By("should have working login", func() {
			Expect(RequireLogin(page, ServerUrl+"/odl/lhm/contacts/list/-1")).To(BeTrue())
			Expect(page).To(HaveURL(ServerUrl + "/odl/lhm/contacts/list/-1"))
			time.Sleep(2200 * time.Millisecond)
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + ".png")).To(Succeed())
		})

		By("not redirecting the user after sucessfull login", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/odl/lhm/contacts/list/-1")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/lhm/contacts/list/-1"))
		})

		By("should have visible navigation on mapselect", func() {
			Expect(page.Find("#side-nav")).To(BeVisible())
		})

		By("should have logout & upload link", func() {
			Expect(page.First("#nav_logout")).To(BeFound())
			Expect(page.First("#nav_upload")).To(BeFound())
		})

		By("Should be able to open contacts tab", func() {
			time.Sleep(1 * time.Second)
			Expect(page.Find("#nav_contacts").Click()).To(Succeed())
		})

		By("should not have broken javascript on lhm/contacts/list/-1", func() {
			Expect(hasUnignoreableJavascriptErrors(page)).To(BeFalse())
		})

		By("should be showing contacts list", func() {
			time.Sleep(3000 * time.Millisecond)
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_default_view.png")).To(Succeed())
		})

		By("should have newContactBtn to click on", func() {
			Expect(page.First("#newContactBtn").Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_newContact_clicked.png")).To(Succeed())
		})

		By("should find inputs for contactForm", func() {
			titleField := page.Find("#contacts #contactEdit #contact_edit_title input")
			descriptionField := page.Find("#contacts #contactEdit #contact_edit_description input")
			cityField := page.Find("#contacts #contactEdit #address_edit_city input")
			postalField := page.Find("#contacts #contactEdit #address_edit_postal input")
			streetField := page.Find("#contacts #contactEdit #address_edit_street input")
			numberField := page.Find("#contacts #contactEdit #address_edit_houseNumber input")

			Expect(titleField.Fill(TestContactTitle)).Should(Succeed())
			Expect(descriptionField.Fill(TestContactDesc)).Should(Succeed())
			Expect(cityField.Fill(TestContactCity)).Should(Succeed())
			Expect(postalField.Fill(TestContactPostal)).Should(Succeed())
			Expect(streetField.Fill(TestContactStreet)).Should(Succeed())
			Expect(numberField.Fill(TestContactNumber)).Should(Succeed())

			time.Sleep(100 * time.Millisecond)
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_newContact_filled.png")).To(Succeed())
		})

		By("should be able to save", func() {
			Expect(page.First("#contacts #contactEdit #saveButton").Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_newContact_submitted.png")).To(Succeed())
		})

		By("should find new contact", func() {
			time.Sleep(400*time.Millisecond)
			searchField := page.Find("#contacts #searchBox input")
			Expect(searchField.Fill(TestContactStreet)).To(Succeed())
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_highlighted_new_contact.png")).To(Succeed())
			Expect(page.Find("#contacts .highlighted")).To(BeFound())
			t,_err:= page.Find("#contacts .highlighted > div:first-of-type").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestContactTitle))
			t,_err = page.Find("#contacts .highlighted > span:nth-of-type(1)").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestContactPostal + " " + TestContactCity))
			t,_err = page.Find("#contacts .highlighted > span:nth-of-type(2)").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestContactStreet + " " + TestContactNumber))
		})

		By("should be able to open edit for new contact", func() {
			id, _err := page.Find("#contacts .highlighted").Attribute("id")
			Expect(_err).To(BeNil())
			id = string(id[strings.Index(id,"_")+1:])
			editButton := page.Find("#contacts #contactIronList #items #item_"+id+" paper-icon-button[icon='icons:create']")
			Expect(editButton.Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			time.Sleep(400 * time.Millisecond)
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_edit_new_contact.png")).To(Succeed())
			titleField := page.Find("#contacts #contactEdit #contact_edit_title input")
			descriptionField := page.Find("#contacts #contactEdit #contact_edit_description input")
			cityField := page.Find("#contacts #contactEdit #address_edit_city input")
			postalField := page.Find("#contacts #contactEdit #address_edit_postal input")
			streetField := page.Find("#contacts #contactEdit #address_edit_street input")
			numberField := page.Find("#contacts #contactEdit #address_edit_houseNumber input")

			Expect(titleField.Fill(TestContactTitleEdit)).Should(Succeed())
			Expect(descriptionField.Fill(TestContactDescEdit)).Should(Succeed())
			Expect(cityField.Fill(TestContactCityEdit)).Should(Succeed())
			Expect(postalField.Fill(TestContactPostalEdit)).Should(Succeed())
			Expect(streetField.Fill(TestContactStreetEdit)).Should(Succeed())
			Expect(numberField.Fill(TestContactNumberEdit)).Should(Succeed())
			time.Sleep(133 * time.Millisecond)
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_newContact_filled.png")).To(Succeed())

		})
		By("should be able to save edited contact", func() {
			Expect(page.First("#contacts #contactEdit #saveButton").Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_newContact_submitted.png")).To(Succeed())
		})

		By("should find edited contact", func() {
			time.Sleep(400*time.Millisecond)
			searchField := page.Find("#contacts #searchBox input")
			Expect(searchField.Fill(TestContactStreetEdit)).To(Succeed())
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_highlighted_edited_contact.png")).To(Succeed())
			Expect(page.Find("#contacts .highlighted")).To(BeFound())
			t,_err:= page.Find("#contacts .highlighted > div:first-of-type").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestContactTitleEdit))
			t,_err = page.Find("#contacts .highlighted > span:nth-of-type(1)").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestContactPostalEdit + " " + TestContactCityEdit))
			t,_err = page.Find("#contacts .highlighted > span:nth-of-type(2)").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestContactStreetEdit + " " + TestContactNumberEdit))
		})
		By("should be able to delete contact", func() {
			id, _err := page.Find("#contacts .highlighted").Attribute("id")
			Expect(_err).To(BeNil())
			id = string(id[strings.Index(id,"_")+1:])
			deleteButton := page.Find("#contacts #contactIronList #items #item_"+id+" paper-icon-button[icon='icons:delete']")
			Expect(deleteButton.Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			time.Sleep(100*time.Millisecond)
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_delete_confirm_contact.png")).To(Succeed())
			Expect(page.Find("#deleteDialog[key='contact_"+id+"'] #delete_dialog_confirm").Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())


		})
		/*By("should not find deleted contact", func() {
			time.Sleep(3000*time.Millisecond)
			Expect(page.Screenshot("./img/contacts_test." + language + "." + browser + "_deleted_contact.png")).To(Succeed())
			searchField := page.Find("#contacts #searchBox input")
			Expect(searchField.Fill(TestContactStreetEdit)).To(Succeed())
			Expect(page.Find("#contacts .highlighted")).ToNot(BeFound())
		})*/

	})

})
