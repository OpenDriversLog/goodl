package main_test

import (
	// "fmt"
	// "strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	//"github.com/Compufreak345/dbg"
	"strings"
)

var TestCarPlate = "F-U 1337"
var TestCarType = "TIGER 2"
var TestCarPlateEdit = "FU-UU 1338"
var TestCarTypeEdit = "TIGER 1"

var _ = Describe("cars_test", func() {
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

	It("should be able to create and show car", func() {

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
		})

		By("not redirecting the user after successful login", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/odl/lhm/trips/cars/1")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/lhm/trips/cars/1"))
			time.Sleep(400 * time.Millisecond)
		})

		By("should have visible navigation on mapselect", func() {
			Expect(page.Find("#side-nav")).To(BeVisible())
		})

		By("should have logout & upload link", func() {
			Expect(page.First("#nav_logout")).To(BeFound())
			Expect(page.First("#nav_upload")).To(BeFound())
		})

		By("Should be able to open cars tab", func() {
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + ".png")).To(Succeed())
			Expect(page.Find("#nav_cars").Click()).To(Succeed())
		})

		By("should be showing cars list", func() {
			time.Sleep(3000 * time.Millisecond)
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_default_view.png")).To(Succeed())
		})

		By("should not have broken javascript on lhm/cars/list/1", func() {
			Expect(hasUnignoreableJavascriptErrors(page)).To(BeFalse())
		})

		// logs, _ := page.ReadNewLogs("browser") // we want to check only newer logs further down

		By("should have newCarBtn to click on", func() {
			Expect(page.First("#newCarBtn")).To(BeFound())
			Expect(page.First("#newCarBtn").Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_newCar_clicked.png")).To(Succeed())
		})

		By("should find inputs for carForm", func() {

			// logs, _ = page.ReadNewLogs("browser") // we want to check only newer logs further down
			// fmt.Println(" JavaScript Logs after clicking #newCarBtn: ", len(logs), logs)

			plateField := page.Find("#cars #carEdit #car_edit_plate input")
			typeField := page.Find("#cars #carEdit #car_edit_type input")
			driverField := page.First("#cars #carEdit #driverSelector")
			Expect(plateField.Fill(TestCarPlate)).Should(Succeed())
			Expect(typeField.Fill(TestCarType)).Should(Succeed())
			Expect(driverField.First("iron-icon").Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Find("#cars #carEdit #driverSelector paper-item[driver-id=\"1\"]").Click()).To(Succeed())
			time.Sleep(133 * time.Millisecond)
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_newCar_filled.png")).To(Succeed())
		})

		By("should be able to save new car", func() {
			Expect(page.First("#cars #carEdit #saveButton").Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_newCar_submitted.png")).To(Succeed())
		})

		By("should find new car", func() {
			time.Sleep(400 * time.Millisecond)
			searchField := page.Find("#cars #searchBox input")
			Expect(searchField.Fill(TestCarPlate)).To(Succeed())
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_highlighted_new_car.png")).To(Succeed())
			Expect(page.Find("#cars .highlighted")).To(BeFound())
			t,_err:= page.Find("#cars .highlighted > div:first-of-type").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestCarPlate))
			t,_err = page.Find("#cars .highlighted > div:nth-of-type(2)").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestCarType))
		})

		By("should be able to open edit for new car", func() {
			id, _err := page.Find("#cars .highlighted").Attribute("id")
			Expect(_err).To(BeNil())
			id = string(id[strings.Index(id,"_")+1:])
			editButton := page.Find("#cars #carIronList #items #item_"+id+" paper-icon-button[icon='icons:create']")
			Expect(editButton.Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			time.Sleep(400 * time.Millisecond)
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_edit_new_car.png")).To(Succeed())
			plateField := page.Find("#cars #carEdit #car_edit_plate input")
			typeField := page.Find("#cars #carEdit #car_edit_type input")
			Expect(plateField.Fill(TestCarPlateEdit)).Should(Succeed())
			Expect(typeField.Fill(TestCarTypeEdit)).Should(Succeed())
			time.Sleep(133 * time.Millisecond)
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_newCar_filled.png")).To(Succeed())

		})
		By("should be able to save edited car", func() {
			Expect(page.First("#cars #carEdit #saveButton").Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_newCar_submitted.png")).To(Succeed())
		})

		By("should find edited car", func() {

			time.Sleep(400 * time.Millisecond)
			searchField := page.Find("#cars #searchBox input")
			Expect(searchField.Fill(TestCarPlateEdit)).To(Succeed())
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_highlighted_edited_car.png")).To(Succeed())
			Expect(page.Find("#cars .highlighted")).To(BeFound())
			t,_err:= page.Find("#cars .highlighted > div:first-of-type").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestCarPlateEdit))
			t,_err = page.Find("#cars .highlighted > div:nth-of-type(2)").Text()
			Expect(_err).To(BeNil())
			Expect(t).To(BeEquivalentTo(TestCarTypeEdit))
		})
		By("should be able to delete car", func() {
			id, _err := page.Find("#cars .highlighted").Attribute("id")
			Expect(_err).To(BeNil())
			id = string(id[strings.Index(id,"_")+1:])
			deleteButton := page.Find("#cars #carIronList #items #item_"+id+" paper-icon-button[icon='icons:delete']")
			Expect(deleteButton.Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())
			time.Sleep(100*time.Millisecond)
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_delete_confirm_car.png")).To(Succeed())
			Expect(page.Find("#deleteDialog[key='car_"+id+"'] #delete_dialog_confirm").Click()).To(Succeed())
			Expect(WaitForRender(page)).To(Succeed())


		})
		/*By("should not find deleted car", func() {

			time.Sleep(3000 * time.Millisecond)
			Expect(page.Screenshot("./img/cars_test." + language + "." + browser + "_deleted_car.png")).To(Succeed())
			searchField := page.Find("#cars #searchBox input")
			Expect(searchField.Fill(TestCarPlateEdit)).To(Succeed())
			Expect(page.Find("#cars .highlighted")).ToNot(BeFound())
		})*/

	})

})
