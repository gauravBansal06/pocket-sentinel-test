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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type AppDirectories struct {
	WorkingDir   string
	TestInfo     string
	Videos       string
	Assets       string
	AppiumLogs   string
	Screenshots  string
	CommandLogs  string
	Applications string
	BinaryLogs   string
	DiskImages   string
}

var (
	BaseAppiumPort       = "4724"
	AppDirs              AppDirectories
	Adb                  string
	GoIOS                string
	AuthenticateEndpoint = "https://stage-accounts.lambdatestinternal.com/api/user/token/auth"
	SyncEndpoint         = "https://mobile-api-gauravb-byod-dev.lambdatestinternal.com/mobile-automation/api/v1/byod/devices/sync"
	SyncToken            string
	UserInfo             UserDetails
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %s", err)
	}

	AppDirs = AppDirectories{
		WorkingDir:   filepath.Join(homeDir, ".lambdatest"),
		Assets:       filepath.Join(homeDir, ".lambdatest", "assets"),
		TestInfo:     filepath.Join(homeDir, ".lambdatest", "tests"),
		Videos:       filepath.Join(homeDir, ".lambdatest", "videos"),
		CommandLogs:  filepath.Join(homeDir, ".lambdatest", "commandlogs"),
		AppiumLogs:   filepath.Join(homeDir, ".lambdatest", "appiumlogs"),
		BinaryLogs:   filepath.Join(homeDir, ".lambdatest", "binarylogs"),
		Screenshots:  filepath.Join(homeDir, ".lambdatest", "screenshots"),
		Applications: filepath.Join(homeDir, ".lambdatest", "applications"),
		DiskImages:   filepath.Join(homeDir, ".lambdatest", "diskimages"),
	}

	dirs := []string{
		AppDirs.WorkingDir,
		AppDirs.Assets,
		AppDirs.TestInfo,
		AppDirs.Videos,
		AppDirs.BinaryLogs,
		AppDirs.AppiumLogs,
		AppDirs.CommandLogs,
		AppDirs.Screenshots,
		AppDirs.Applications,
		AppDirs.DiskImages,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory '%s': %s", dir, err)
		}
	}

	Adb = fmt.Sprintf("%s/adb", AppDirs.Assets)
	GoIOS = fmt.Sprintf("%s/go-ios", AppDirs.Assets)
}

func OS() string {
	if runtime.GOOS == "darwin" {
		return "macos"
	} else if runtime.GOOS == "linux" {
		return "linux"
	} else {
		return "windows"
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
		log.Fatal(err)
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
	cmd := exec.Command("lsof", "-i", ":"+port)
	output, err := cmd.Output()
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
	killCmd := exec.Command("kill", pid)
	if err := killCmd.Run(); err != nil {
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
		resp, err := http.Get(appPath)
		if err != nil {
			return appPath, err
		}
		defer resp.Body.Close()

		out, err := os.Create(filePath)
		if err != nil {
			return appPath, err
		}
		defer out.Close()
		_, err = io.Copy(out, resp.Body)
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
	if os == "iOS" {
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
		cmd := exec.Command(Adb, "-s", udid, "shell", "pidof", packageName)
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}
	return "", nil
}

func FindDeviceIP(udid, os string) (string, error) {
	if os == "android" {
		cmd := exec.Command(Adb, "-s", udid, "shell", "ip", "route")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		fields := strings.Split(string(out), " ")
		return fields[len(fields)-1], nil
	} else {

	}
	return "", errors.New("IP not found")
}
