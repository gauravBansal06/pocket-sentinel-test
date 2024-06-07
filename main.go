package main

import (
	"byod/common"
	"byod/remote"
	"byod/services"
	"byod/watcher"
	"context"
	"encoding/base64"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// main orchestrates the starting sequence of the application.
func main() {
	user, key := parseFlags()               // Retrieve user credentials from command-line flags.
	userInfo := authenticateUser(user, key) // Authenticate the user with the provided credentials.

	remote.LaunchTunnel(user, key) //launch tunnel
	time.Sleep(4 * time.Second)    //to get tunnel up and runing

	//create stop channel for graceful shutdown
	stopChan := make(chan struct{})
	mainExit := make(chan struct{})
	go shutdownListener(stopChan, mainExit)

	initializeServices(userInfo) // Initialize the necessary services with authenticated user information.

	watcher.SyncBinaryHost(1) //this is to mark previously connected devices disconnected and clear any tests if running as binary is started now

	startDeviceWatcher(stopChan)                         // Start the device watcher to monitor device activities.
	go services.ResetAuthenticatedJwtUsersCron(stopChan) //to reset jwt token map after 30 mins

	services.StartServer() // Start the main server at end to handle incoming requests.

	//wait on main exit post graceful shutdown in shutdownListener
	<-mainExit
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
		log.Println("Unable to authenticate username and access key: ", err)
		os.Exit(1) // Exit the program if authentication fails.
	}
	return userInfo // Return the authenticated user's details.
}

// initializeServices initializes application services and global state with the user's details.
func initializeServices(userInfo common.UserDetails) {
	log.Println("staring services initialization")

	services.Initialize() // Initialize basic services.

	// Set global user information and synchronization token for the session.
	common.UserInfo = userInfo
	common.SyncToken = base64.StdEncoding.EncodeToString([]byte(userInfo.Username + ":" + userInfo.ApiToken))

	log.Println("services initialization complete")
}

// startDeviceWatcher initializes and starts a device watcher to monitor connected devices.
func startDeviceWatcher(stopChan chan struct{}) {
	log.Println("starting device watcher process....")
	deviceWatcher, err := watcher.NewDeviceWatcher() // Create a new device watcher.
	if err != nil {
		log.Println("Error initializing device watcher: ", err)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT) // Exit the program if the device watcher cannot be initialized.
	}
	common.WG.Add(1)
	go deviceWatcher.Watch(stopChan) // Run the device watcher in a new goroutine.
}

// function for graceful shutdown
func shutdownListener(stopChan, mainExit chan struct{}) {
	log.Println("starting shutdown listener...waiting on signal")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	signalReceived := <-sigChan
	log.Println("shutdownListener :: termination signal received: ", signalReceived.String())
	log.Println("starting shutdown of binary....")

	close(stopChan)

	// Create a context with a timeout to ensure the server shuts down gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	services.KillServer(ctx)

	common.WG.Wait() //wait for all go routines to finish
	log.Println("shutdownListener: all go routines finished")

	watcher.SyncBinaryHost(1) //to mark all devices disconnected and clear any running tests without waiting for keep alive timeout
	remote.KillTunnel()
	log.Println("binary shutdown complete..")
	close(mainExit)
}
