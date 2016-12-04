package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	conf "github.com/OpenDriversLog/goodl/config"
	"github.com/OpenDriversLog/webfw"
	"os"
)

var _ = Describe("Config", func() {

	var (
		config *webfw.ServerConfig
	)

	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	Describe("config sets dirs for different environment variables", func() {
		It("should set production", func() {
			defer GinkgoRecover()

			os.Setenv("ENVIRONMENT", "production")
			config = conf.GetConfig()

			Expect(config.Environment).To(Equal("production"))
			// Expect(config.WebUrl).To(Equal("https://opendriverslog.de/beta"))
			// TODO: FS: switch to beta!
			Expect(config.WebUrl).To(Equal("https://opendriverslog.de/alpha"))
		})

		It("should set development", func() {
			defer GinkgoRecover()

			os.Setenv("ENVIRONMENT", "development")
			config = conf.GetConfig()

			Expect(config.Environment).To(Equal("development"))
			Expect(config.WebUrl).To(Equal("http://localhost:4000/alpha"))
		})

		It("should set test", func() {
			defer GinkgoRecover()

			os.Setenv("ENVIRONMENT", "test")
			config = conf.GetConfig()

			Expect(config.Environment).To(Equal("test"))
			Expect(config.WebUrl).To(Equal("http://localhost:4000/test"))
		})

		It("should set intern", func() {
			defer GinkgoRecover()

			os.Setenv("ENVIRONMENT", "intern")
			config = conf.GetConfig()

			Expect(config.Environment).To(Equal("intern"))
			Expect(config.WebUrl).To(Equal("https://opendriverslog.de/beta-intern"))
		})

		It("should set beta-dev (dev-server)", func() {
			defer GinkgoRecover()

			os.Setenv("ENVIRONMENT", "dev-server")
			config = conf.GetConfig()

			Expect(config.Environment).To(Equal("dev-server"))
			Expect(config.WebUrl).To(Equal("https://opendriverslog.de/beta-dev"))
		})

		It("should set default: test", func() {
			defer GinkgoRecover()

			os.Setenv("ENVIRONMENT", "")
			config = conf.GetConfig()

			Expect(config.Environment).To(Equal("test"))
			Expect(config.WebUrl).To(Equal("http://localhost:4000/alpha"))
		})
	})

})
