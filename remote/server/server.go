package server

// import (
// 	"byod/common"
// 	"bytes"
// 	"fmt"
// 	"log"
// 	"net"
// 	"os/exec"
// 	"strings"
// 	"sync"
// 	"time"
// )

// func Start(wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	// go android.StartTCPListner("6037", "5037")
// 	// go android.CheckAutomator("6037")
// 	StartSocketListener("9037")
// }

// func StartSocketListener(target string) {
// 	listener, err := net.Listen("unix", "/var/run/usbmuxd")
// 	if err != nil {
// 		log.Println("Listen failed", err.Error())
// 		return
// 	}
// 	for {
// 		inconn, err := listener.Accept()
// 		if err != nil {
// 			log.Println("Accept failed")
// 			return
// 		}
// 		go common.ForwardLocalPortToProxy(target, inconn)
// 	}
// }

// func CheckAutomator(port string) {
// 	for {
// 		devices := device.ListDevices(port)
// 		if len(devices) > 0 {
// 			for _, device := range devices {
// 				targetPort := getAutomatorPort(device.UDID, port)
// 				if targetPort != "" && common.IsPortAvailable(targetPort) {
// 					fmt.Println("target port", targetPort)
// 					go StartTCPListner(targetPort, targetPort)
// 					time.Sleep(200 * time.Millisecond)
// 				}
// 			}
// 		}
// 	}
// }

// func getAutomatorPort(udid, port string) string {
// 	cmd := exec.Command(common.Adb, "-P", port, "-s", udid, "forward", "--list")
// 	var out bytes.Buffer
// 	cmd.Stdout = &out
// 	err := cmd.Run()
// 	if err != nil {
// 		fmt.Println("Error executing adb command:", err)
// 		return ""
// 	}

// 	lines := strings.Split(out.String(), "\n")
// 	for _, line := range lines {
// 		if strings.Contains(line, "tcp:6790") {
// 			fields := strings.Fields(line)
// 			return strings.Split(fields[1], ":")[1]
// 		}
// 	}
// 	return ""
// }

// func StartTCPListner(source string, target string) {
// 	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", source))
// 	if err != nil {
// 		log.Println("Listen failed")
// 		return
// 	}
// 	for {
// 		inconn, err := listener.Accept()
// 		if err != nil {
// 			log.Println("Accept failed")
// 			return
// 		}
// 		go common.ForwardLocalPortToProxy(target, inconn)
// 	}
// }
