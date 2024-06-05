package main

import (
	"byod/common"
	"byod/remote"
	"byod/services"
	"byod/watcher"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// main orchestrates the starting sequence of the application.
func main() {
	user, key := parseFlags()               // Retrieve user credentials from command-line flags.
	userInfo := authenticateUser(user, key) // Authenticate the user with the provided credentials.

	remote.LaunchTunnel(user, key)

	initializeServices(userInfo) // Initialize the necessary services with authenticated user information.
	startDeviceWatcher()         // Start the device watcher to monitor device activities.
	services.StartServer()       // Start the main server to handle incoming requests.

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	fmt.Println("terminating signal received")
	remote.KillTunnel()
	services.KillServer()
}

// parseFlags parses and validates command-line flags for user credentials.
func parseFlags() (string, string) {
	user := flag.String("user", "", "Username for the application")
	key := flag.String("key", "", "Key for the application")

	env := flag.String("env", "stage", "env: stage/prod, default 'prod'")
	tunnel := flag.String("tunnel", "./LT", "LT Tunnel Binary Path, default './LT'")

	flag.Parse() // Parse all command-line flags.

	if *user == "" || *key == "" {
		log.Println("Both --user and --key flags are required.")
		flag.PrintDefaults() // Display default help messages for flags.
		os.Exit(1)           // Exit the program with an error code.
	}
	remote.SetTunnelArgs(*tunnel, *env)
	return *user, *key // Return the parsed username and key.
}

// authenticateUser attempts to authenticate a user with the given username and key.
func authenticateUser(user, key string) common.UserDetails {
	userInfo, err := services.AuthenticateUser(user, key)
	if err != nil {
		log.Println("Unable to authenticate username and access key:", err)
		os.Exit(1) // Exit the program if authentication fails.
	}
	return userInfo // Return the authenticated user's details.
}

// initializeServices initializes application services and global state with the user's details.
func initializeServices(userInfo common.UserDetails) {
	services.Initialize() // Initialize basic services.
	// Set global user information and synchronization token for the session.
	common.UserInfo = userInfo
	common.SyncToken = base64.StdEncoding.EncodeToString([]byte(userInfo.Username + ":" + userInfo.ApiToken))
}

// startDeviceWatcher initializes and starts a device watcher to monitor connected devices.
func startDeviceWatcher() {
	deviceWatcher, err := watcher.NewDeviceWatcher() // Create a new device watcher.
	if err != nil {
		log.Println("Error initializing device watcher:", err)
		os.Exit(1) // Exit the program if the device watcher cannot be initialized.
	}
	go deviceWatcher.Watch() // Run the device watcher in a new goroutine.
}
