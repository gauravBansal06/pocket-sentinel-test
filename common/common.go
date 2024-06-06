package common

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/mholt/archiver"
)

type AppDirectories struct {
	WorkingDir, Assets, TestInfo, Videos, CommandLogs, AppiumLogs,
	BinaryLogs, Screenshots, Applications, DiskImages string
}

var (
	WG sync.WaitGroup

	BaseAppiumPort = "4724"
	AppDirs        AppDirectories
	Adb            string
	GoIOS          string
	Appium         string

	SanitisatioEndpoint = "https://prod-mobile-automation-artefects.lambdatest.com/byod-assets"

	BasicAuthenticateEndpoint  = "https://stage-accounts.lambdatestinternal.com/api/user/token/auth"
	BearerAuthenticateEndpoint = "https://stage-accounts.lambdatestinternal.com/api/user/auth"
	UserContextKey             = "userInfo"

	SyncEndpoint          = "https://mobile-api-gauravb-byod-dev.lambdatestinternal.com/mobile-automation/api/v1/byod/devices/sync"
	BinaryStartupEndpoint = "https://mobile-api-gauravb-byod-dev.lambdatestinternal.com/mobile-automation/api/v1/byod/host/startup"

	SyncToken string
	UserInfo  UserDetails
)

func OS() string {
	if runtime.GOOS == "darwin" {
		return "macos"
	} else if runtime.GOOS == "linux" {
		return "linux"
	} else {
		return "windows"
	}
}

func GetDeviceCommand(os string) string {
	switch os {
	case "android":
		return Adb // Assuming 'Adb' is the path or command to run Android ADB
	case "ios":
		return GoIOS // Assuming 'GoIOS' is the path or command for the iOS management tool
	default:
		log.Printf("Unsupported OS: %s\n", os)
		return "" // Return empty if the OS is not supported
	}
}

func Execute(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(out), "\n"), err
}

func ExecuteAsync(command string) (*exec.Cmd, error) {
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Start()
	return cmd, err
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Println("error fetching outboubd ip: ", err)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func IsPortAvailable(port string) bool {
	addr := fmt.Sprintf("127.0.0.1:%s", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func ForwardLocalPortToProxy(port string, inconn net.Conn) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "13.126.37.58:1536")
	if err != nil {
		log.Println("ResolveTCPAddr failed")
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Println("DialTCP failed", port)
		return
	}

	_, err = conn.Write([]byte(fmt.Sprintf("CONNECT localhost:%s HTTP/1.1\r\nHost: localhost\r\n\r\n", port)))
	if err != nil {
		log.Println("Write failed")
		return
	}

	buf := make([]byte, 4096)

	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Read failed")
		return
	}
	response := string(buf[0:n])
	if !strings.Contains(response, "Connection establi") {
		log.Printf("%s\n", response)
	}

	done := make(chan bool)

	go func() {
		defer inconn.Close()
		defer conn.Close()
		io.Copy(conn, inconn)
		done <- true
	}()

	go func() {
		defer inconn.Close()
		defer conn.Close()
		io.Copy(inconn, conn)
		done <- true
	}()

	<-done
	<-done
}

func KillProcessOnPort(port string) error {
	output, err := Execute(fmt.Sprintf("lsof -i:%s", port))
	if err != nil {
		return fmt.Errorf("failed to execute lsof: %s", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	scanner.Scan()
	scanner.Scan()
	if scanner.Text() == "" {
		return fmt.Errorf("no process found on port %s", port)
	}
	columns := strings.Fields(scanner.Text())
	if len(columns) < 2 {
		return fmt.Errorf("unexpected lsof output format")
	}
	pid := columns[1]
	_, err = Execute(fmt.Sprintf("kill -9 %s", pid))
	if err != nil {
		return fmt.Errorf("failed to kill process: %s", err)
	}
	return nil
}

func DownloadAppIfRequired(appPath string) (string, error) {
	if strings.HasPrefix(appPath, "http://") || strings.HasPrefix(appPath, "https://") {
		parsedURL, err := url.Parse(appPath)
		if err != nil {
			return appPath, err
		}
		filePath := fmt.Sprintf("%s/%s", AppDirs.Applications, path.Base(parsedURL.Path))
		err = Download(appPath, filePath)
		return filePath, err
	}
	return appPath, nil
}

func GetPidByBundleId(bundleId, udid string) int {
	cmd := exec.Command("pymobiledevice3", "developer", "dvt", "proclist", "--no-color", "--udid", udid)
	out, err := cmd.Output()
	if err != nil {
		return -1
	}

	var processes []ProcessInfo
	err = json.Unmarshal(out, &processes)
	if err != nil {
		return -1
	}

	for _, process := range processes {
		if process.BundleIdentifier == bundleId {
			return process.PID
		}
	}
	return -1
}

func GetForegroundApp(udid, packageName, os string) (string, error) {
	if os == "ios" {
		cmd := exec.Command("pymobiledevice3", "developer", "dvt", "proclist", "--no-color", "--udid", udid)
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}

		var processes []ProcessInfo
		err = json.Unmarshal(out, &processes)
		if err != nil {
			return "", err
		}

		for _, process := range processes {
			// if process.IsApplication == true && process.ForegroundRunning == true {
			// 	return process.BundleIdentifier, strconv.Itoa(process.PID)
			// }
			if process.BundleIdentifier == packageName {
				return strconv.Itoa(process.PID), nil
			}
		}
	} else {
		command := fmt.Sprintf("%s -s %s shell pidof %s", Adb, udid, packageName)
		out, err := Execute(command)
		if err != nil {
			return "", err
		}
		return strings.Trim(string(out), "\n"), nil
	}
	return "", nil
}

func FindDeviceIP(udid, os string) (string, error) {
	if os == "android" {
		out, err := Execute(fmt.Sprintf("%s -s %s shell ip route", Adb, udid))
		if err != nil {
			return "", err
		}
		response := strings.Trim(string(out), "\n")
		fields := strings.Fields(response)
		return fields[len(fields)-1], nil
	} else {

	}
	return "", errors.New("IP not found")
}

func Download(source, target string) error {
	log.Println("Downloading", source, "at", target)
	resp, err := http.Get(source)
	if err != nil {
		log.Println("Failed to download", source)
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	out.Chmod(0755)
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func Unzip(source, dest string) error {
	zip := archiver.NewZip()
	err := zip.Unarchive(source, dest)
	os.Remove(source)
	if err != nil {
		log.Println("UnpackIPA: Couldn't unzip the file:", err)
		return err
	}
	os.Remove(source)
	return nil
}
