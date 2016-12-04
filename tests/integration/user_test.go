package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pretty "github.com/tonnerre/golang-pretty"
	"github.com/OpenDriversLog/goodl/utils/userManager"
)

var _ = Describe("User", func() {

	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	Describe("user in db", func() {

		//it probably opens the wrong database and therefore fails

		It("should be able to get user from DB", func() {
			defer GinkgoRecover()
			var usr, err = userManager.GetUserFromDb("test@opendriverslog.de")
			pretty.Print("GetUserFromDb: \n  ", usr)
			pretty.Print("err: \n  ", usr)

			Expect(err).NotTo(HaveOccurred())
			Expect(usr).NotTo(BeNil())

			Expect(usr.IfirstName).To(Equal("Rainer"))
			Expect(usr.IlastName).To(Equal("von Zufall"))
			Expect(usr.Ititle).To(Equal("Seine Lordschaft"))
		})

		PIt("should be able to create reset password key", func() {
			defer GinkgoRecover()

			key, err := userManager.GetPasswordResetKey("test@opendriverslog.de")
			Expect(err).NotTo(HaveOccurred())
			Expect(key).NotTo(BeNil())

		})
	})
})
