package integration_test

import (
	"database/sql"
	"os"
	"path"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pretty "github.com/tonnerre/golang-pretty"

	"github.com/Compufreak345/dbg"
	"github.com/OpenDriversLog/goodl-lib/tools"
	"github.com/OpenDriversLog/goodl/config"
	"github.com/OpenDriversLog/webfw"
)

func TestIntegration(t *testing.T) {
	tools.RegisterSqlite("SQLITE")
	RegisterFailHandler(Fail)

	RunSpecs(t, "Integration Test Suite")
}

const TAG = dbg.Tag("goodl/integrationTest")

var language = "de"
var Config *webfw.ServerConfig
var DBCon *sql.DB

var _ = BeforeSuite(func() {
	os.Setenv("ENVIRONMENT", "test")
	Config = config.GetConfig()

	_, filename, _, _ := runtime.Caller(1)
	Config.RootDir = path.Dir(filename)

	webfw.SetConfig(Config)
	dbg.I(TAG, "Server starting up with config : \n %# v\n", pretty.Formatter(Config))
})
