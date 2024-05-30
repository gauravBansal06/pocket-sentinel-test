package main

import (
	"byod/common"
	"byod/services"
	"byod/watcher"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
)

func main() {
	user := flag.String("user", "", "Username for the application")
	key := flag.String("key", "", "Key for the application")

	flag.Parse()

	if *user == "" || *key == "" {
		fmt.Println("Both --user and --key flags are required.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	userInfo, err := services.AuthenticateUser(*user, *key)
	if err != nil {
		fmt.Println("Unable to authenticate username and accesskey")
		os.Exit(1)
	}
	common.UserInfo = userInfo
	common.SyncToken = base64.StdEncoding.EncodeToString([]byte(userInfo.Username + ":" + userInfo.ApiToken))

	deviceWatcher, err := watcher.NewDeviceWatcher()
	if err != nil {
		fmt.Println("Error initializing device watcher:", err)
	}
	go deviceWatcher.Watch()
	services.StartServer()
}
