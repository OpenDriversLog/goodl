package integration

import (
	"fmt"
	. "github.com/OpenDriversLog/goodl/config"
)

func PrintConfig() {
	config := GetConfig()
	fmt.Sprint(config)
}
