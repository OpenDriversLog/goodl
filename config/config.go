// Package config provides the ServerConfig for the default GOODL-Server.
package config

import (
	"path"
	"runtime"
	"time"

	"os"

	"github.com/OpenDriversLog/webfw"
	"strings"
)

// GetConfig returns a new config for the GOODL-Server!
func GetConfig() *webfw.ServerConfig {
	config := webfw.NewServerConfig()

	// project root dir, this code can not be put to main func
	// @see https://github.com/QLeelulu/goku/blob/master/examples/mustache-template/app.go
	_, filename, _, _ := runtime.Caller(1)
	config.RootDir = path.Dir(filename)
	config.SharedDir = "/databases"
	config.RedisAddress = os.Getenv("REDIS_PORT_6379_TCP_ADDR") + ":" + os.Getenv("REDIS_PORT_6379_TCP_PORT")
	var port = os.Getenv("HTTP_PORT")

	if port == "" {
		port = ":4000"
	} else if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	config.HttpAddress = port

	//increased for KML import of a month data
	config.MaxResponseTime = 180 * time.Second

	config.GitLabPath = os.Getenv("GITLAB_URL")
	config.SmtpHost = os.Getenv("SMTP_HOST")
	config.SmtpPort = os.Getenv("SMTP_PORT")
	config.Environment = os.Getenv("ENVIRONMENT")

	switch config.Environment {
	case "production": //master branch setup
		config.WebUrl = "https://opendriverslog.de/alpha"
		config.SubDir = "/alpha"
		break
	case "development": //for local enjoyment
		config.WebUrl = "http://localhost:4000/alpha"
		config.SubDir = "/alpha"
		break
	case "test": // runs on TheDeadServer while jenkins tests
		config.WebUrl = "http://localhost:4000/test"
		config.SubDir = "/test"
		// config.SharedDir = path.Dir(filename)
		break
	case "intern": // intern master-branch :9004
		config.WebUrl = "https://opendriverslog.de/beta-intern"
		config.SubDir = "/beta-intern"
		break
	case "dev-server": // current development, after MR GoODL :9003
		config.WebUrl = "https://opendriverslog.de/beta-dev"
		config.SubDir = "/beta-dev"
		break

	default:
		config.Environment = "test"
		config.WebUrl = "http://localhost:4000/alpha"
		config.SubDir = "/alpha"
	}

	return config
}
