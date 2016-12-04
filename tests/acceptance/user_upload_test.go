package main_test

import (
	"bytes"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	//"github.com/Compufreak345/dbg"
)

var _ = Describe("user_upload_test", func() {
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

	It("should be able to accept uploaded data", func() {
		By("should have working login", func() {
			Expect(page.Navigate(ServerUrl + "/logout")).To(Succeed())
			time.Sleep(3 * time.Second)
			Expect(RequireLogin(page, ServerUrl+"/odl/login")).To(BeTrue())
		})

		By("should accept uploaded KML data", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/odl/upload/upload")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/upload/upload"))

			//upload_key := page.First("#upload_key input")
			upload_externalData := page.Find("#upload_externalData")
			//upload_type := page.Find("#upload_dataType input")
			time.Sleep(500 * time.Millisecond)

			upload_key := page.Find("#upload_key iron-icon")
			Expect(upload_key.Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/user_upload_test_upload." + language + "." + browser + "_deviceDropdown.png")).To(Succeed())

			device1Button := page.Find("#upload_key paper-item[device-id=\"1\"]")
			Expect(device1Button.Click()).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			upload_dataType := page.Find("#upload_dataType iron-icon")
			Expect(upload_dataType.Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			Expect(page.Screenshot("./img/user_upload_test_upload." + language + "." + browser + "_typeDropdown.png")).To(Succeed())

			kmlButton := page.Find("#upload_dataType paper-item[name='KML']")
			Expect(kmlButton.Click()).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			Expect(page.Screenshot("./img/user_upload_test_upload." + language + "." + browser + "_KML_filled.png")).To(Succeed())

			//Expect(upload_key.Fill("TestDeviceKML")).Should(Succeed())
			//Expect(upload_type.Fill("KML")).Should(Succeed())

			//var gpsData, err = ReadFile("./files/gps.log")

			buf := bytes.NewBuffer(nil)
			f, err := os.Open("./files/gps.kml")
			Expect(err == nil).To(BeTrue())
			io.Copy(buf, f)
			f.Close()

			gpsData := string(buf.Bytes())

			pms := make(map[string]interface{})
			pms["gpsData"] = string(gpsData)
			Expect(upload_externalData.Fill("Crap")).Should(Succeed())
			page.RunScript("document.getElementById('upload_externalData').value=gpsData;", pms, nil)

			time.Sleep(5000 * time.Millisecond)
			Expect(page.Screenshot("./img/user_upload_test_upload." + language + "." + browser + "_filled-KML.png")).To(Succeed())
			Expect(page.Find("#submit").Click()).To(Succeed())
			time.Sleep(15000 * time.Millisecond)
			// TODO: Find out why Jenkins does not like that toast detection.
			/*
				Eventually(func() error {
					visible, _ := page.Find("#toast").Visible()
					if !visible {
						return errors.New("toast not visible")
					}
					return nil
				}, 40000*time.Millisecond).Should(Succeed())*/
			Expect(page.Screenshot("./img/user_upload_test_upload." + language + "." + browser + "_submitted-KML.png")).To(Succeed())
		})

		By("should accept uploaded NMEA data", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/odl/upload/upload")).To(Succeed())
			Expect(page).To(HaveURL(ServerUrl + "/odl/upload/upload"))

			//upload_dataType := page.Find("#upload_dataType")
			//upload_key := page.First("#upload_key input")
			upload_externalData := page.Find("#upload_externalData")
			time.Sleep(500 * time.Millisecond)
			//Expect(upload_dataType.Select("NMEA/GPRMC")).Should(Succeed())
			//Expect(upload_key.Fill("TestDevice")).Should(Succeed())
			upload_key := page.Find("#upload_key iron-icon")
			Expect(upload_key.Click()).To(Succeed())
			time.Sleep(1000 * time.Millisecond)
			Expect(page.Screenshot("./img/user_upload_test_upload." + language + "." + browser + "_deviceSelect.png")).To(Succeed())

			device1Button := page.Find("#upload_key paper-item[device-id=\"1\"]")
			Expect(device1Button.Click()).To(Succeed())
			time.Sleep(100 * time.Millisecond)
			/*upload_dataType := page.Find("#upload_dataType iron-icon")
			Expect(upload_dataType.Click()).To(Succeed())
			time.Sleep(500 * time.Millisecond)
			nmeaButton := page.Find("#upload_dataType paper-item[name='NMEA/GPRMC']")
			Expect(nmeaButton.Click()).To(Succeed())*/
			time.Sleep(100 * time.Millisecond)
			Expect(page.Screenshot("./img/user_upload_test_upload." + language + "." + browser + "_NMEA_filled.png")).To(Succeed())

			//var gpsData, err = ReadFile("./files/gps.log")

			buf := bytes.NewBuffer(nil)
			f, err := os.Open("./files/gps.log")
			Expect(err == nil).To(BeTrue())
			io.Copy(buf, f)
			f.Close()

			gpsData := string(buf.Bytes())

			pms := make(map[string]interface{})
			pms["gpsData"] = string(gpsData)
			Expect(upload_externalData.Fill("Crap")).Should(Succeed())
			page.RunScript("document.getElementById('upload_externalData').value=gpsData;", pms, nil)

			time.Sleep(5000 * time.Millisecond)
			Expect(page.Screenshot("./img/user_upload_test_upload." + language + "." + browser + "_filled-NMEA.png")).To(Succeed())
			Expect(page.Find("#submit").Click()).To(Succeed())
			time.Sleep(15000 * time.Millisecond)

			// TODO: Find out why this fails when Jenkins runs the tests...
			/*Eventually(func() error {
				visible, _ := page.Find("#toast").Visible()
				if !visible {
					return errors.New("toast not visible")
				}
				return nil
			}, 40000*time.Millisecond).Should(Succeed())*/

			Expect(page.Screenshot("./img/user_upload_test_upload." + language + "." + browser + "_submitted-NMEA.png")).To(Succeed())
		})

		By("should load lhm after upload", func() {
			Expect(NavigateAndRender(page, ServerUrl+"/odl/lhm/trips/list/-1")).To(Succeed())
			time.Sleep(20 * time.Second)
			Expect(page).To(HaveURL(ServerUrl + "/odl/lhm/trips/list/-1"))
			Expect(page.Screenshot("./img/user_upload_test_lhm_with_data." + language + "." + browser + ".png")).To(Succeed())
		})

		// TODO: @CS @compu didnt work for me, please see how to fix that
		/*By("should have exactly one visible trip from NMEA data", func() {
					WaitForRender(page)
					Expect(page.Find("#tripIronList")).NotTo(BeNil())
					var itemsInTripList bool
					Expect(page.RunScript(`
						if(document.getElementById('tripIronList').children[1].children.length == 2) return true; return false;
		`, nil, &itemsInTripList)).To(Succeed())
					Expect(itemsInTripList).To(BeTrue()) // 2 trips + 1 template element

					Expect(page.Find("#tripMapItem1")).NotTo(BeNil())
					// 	var singleTripFound bool
					// 	Expect(page.RunScript(`
					// 		if(document.getElementById('tripMapItem1').trip.DeviceId != 1) return false;
					// 		if(document.getElementById('tripMapItem2').trip.DeviceId != 1) return false;
					// 		if(document.getElementById('tripMapItem3') != null) return false;
					// 		if(document.getElementById('tripMapItem2').children[0].nodeName != "LEAFLET-GEOJSON") return false;
					// 		return true;
					// 	`, nil, &singleTripFound)).To(Succeed())
					// 	Expect(singleTripFound).To(BeTrue())
				})*/

		// TODO: reimplement
		/*By("should have visible keypoints from NMEA data", func() {
			var keyPointsFound bool
			Expect(page.RunScript(`
				if(document.getElementById('tripMapItem1').children[2].children[0].id != "noContactMarker") return false;
				return true;
			`, nil, &keyPointsFound)).To(Succeed())
			Expect(keyPointsFound).To(BeTrue())
		})*/
		/*
			//TODO: change filters...
			By("should have visible trips from KML data", func() {

			})

			By("should have visible keypoints from KML data", func() {

			})*/

	})

})
