package watcher

import (
	"byod/common"
	"byod/services"
	"byod/storage"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	adb "github.com/zach-klippenstein/goadb"
)

type HostInfo struct {
	IsSyncHost                bool         `json:"is_sync_host"`
	HostIP                    string       `json:"host_ip"`
	HostPort                  int          `json:"host_port"`
	DiscoveryTunnelIdentifier string       `json:"discovery_tunnel_identifier"`
	HostType                  string       `json:"host_type"`
	HostUserID                string       `json:"host_user_id"`
	DedicatedOrg              string       `json:"dedicated_org"`
	Devices                   []DeviceInfo `json:"devices"`
}

type DeviceInfo struct {
	OS            string `json:"os"`
	Name          string `json:"name"`
	UDID          string `json:"udid"`
	Brand         string `json:"brand"`
	Status        string `json:"status"`
	OSVersion     string `json:"os_version"`
	FullOSVersion string `json:"full_os_version"`
}

type DeviceWatcher struct {
	HostIP     string
	AdbClient  *adb.Adb
	OldDevices map[string]DeviceInfo
}

func NewDeviceWatcher() (*DeviceWatcher, error) {
	client, _ := adb.NewWithConfig(adb.ServerConfig{Port: 5037})
	return &DeviceWatcher{
		HostIP:     common.GetOutboundIP(),
		OldDevices: make(map[string]DeviceInfo),
		AdbClient:  client,
	}, nil
}

func (dw *DeviceWatcher) Watch() {
	go dw.launchTunnel()
	go dw.watchDevices()

	dw.keepAlive()
}

func (dw *DeviceWatcher) watchDevices() {
	for {
		dw.HostIP = common.GetOutboundIP() // Update IP if needed
		newDevices := make(map[string]DeviceInfo)
		devices, err := ios.ListDevices()
		if err != nil {
			log.Println("Failed to list iOS devices:", err)
		} else {
			for _, device := range devices.DeviceList {
				udid := device.Properties.SerialNumber
				deviceInfo := DeviceInfo{
					OS:     "ios",
					UDID:   udid,
					Status: "connected",
				}
				values, _ := ios.GetValues(device)
				deviceInfo.Name = values.Value.DeviceName
				deviceInfo.Brand = values.Value.DeviceClass
				deviceInfo.FullOSVersion = values.Value.ProductVersion
				deviceInfo.OSVersion = strings.Split(deviceInfo.FullOSVersion, ".")[0]
				newDevices[udid] = deviceInfo
				err = dw.syncDiskImages(udid, deviceInfo.FullOSVersion)
				if err == nil {
					deviceInfo.Status = "ready"
				}
			}
		}

		androidDevices, err := dw.AdbClient.ListDeviceSerials()
		if err != nil {
			log.Println("Failed to list Android devices:", err)
		} else {
			for _, udid := range androidDevices {
				deviceInfo := DeviceInfo{
					OS:     "android",
					UDID:   udid,
					Status: "connected",
				}

				device := dw.AdbClient.Device(adb.DeviceWithSerial(udid))
				state, _ := device.State()
				if state.String() == "StateOnline" {
					deviceInfo.Status = "ready"
				}
				deviceInfo.Name, _ = device.RunCommand("getprop ro.product.model")
				deviceInfo.Name = strings.Trim(deviceInfo.Name, "\n")

				deviceInfo.Brand, _ = device.RunCommand("getprop ro.product.brand")
				deviceInfo.Brand = strings.Trim(deviceInfo.Brand, "\n")

				deviceInfo.OSVersion, _ = device.RunCommand("getprop ro.build.version.release")
				deviceInfo.OSVersion = strings.Trim(deviceInfo.OSVersion, "\n")

				deviceInfo.FullOSVersion = deviceInfo.OSVersion
				newDevices[udid] = deviceInfo
			}
		}

		for udid, device := range dw.OldDevices {
			if _, ok := newDevices[udid]; !ok {
				log.Println("Disconnected:", udid)
				device.Status = "disconnected"
				go dw.sync(false, []DeviceInfo{device})
			}
		}

		for udid, device := range newDevices {
			oldDevice, ok := dw.OldDevices[udid]
			if !ok {
				dw.setAppiumPort(udid)
				log.Println("Connected:", udid)
				go dw.sync(false, []DeviceInfo{device})
				if device.OS == "ios" {
					go dw.installRunner(udid)
				}
			} else if oldDevice.OS == "android" && oldDevice.Status != device.Status {
				go dw.sync(false, []DeviceInfo{device})
			}
		}
		dw.OldDevices = newDevices
		time.Sleep(3 * time.Second)
	}
}

func (dw *DeviceWatcher) installRunner(udid string) {
	runner := fmt.Sprintf("%s/WebDriverAgentRunner-Runner.app", common.AppDirs.Assets)
	command := fmt.Sprintf("%s install --path=%s --udid %s", common.GoIOS, runner, udid)
	_, err := common.Execute(command)
	if err != nil {
		log.Println("error while installing runner", err.Error())
	}
}

func (dw *DeviceWatcher) launchTunnel() {
	common.Execute("pkill -SIGTERM remoted go-ios")
	_, err := common.Execute(fmt.Sprintf("%s tunnel start --pair-record-path=/tmp", common.GoIOS))
	if err != nil {
		log.Println("Unable to launch tunnel")
	} else {
		log.Println("Tunnel launched")
	}
}

func (dw *DeviceWatcher) setAppiumPort(udid string) {
	var port string
	storage.Store.Get("Appium_Port_"+udid, &port)
	if port == "" {
		storage.Store.Put("Appium_Port_"+udid, common.BaseAppiumPort)
		port, _ := strconv.Atoi(port)
		common.BaseAppiumPort = fmt.Sprintf("%d", (port + 1))
	}
}

func (dw *DeviceWatcher) sync(isSync bool, devices []DeviceInfo) {
	hostInfo := HostInfo{
		IsSyncHost:                isSync,
		HostIP:                    common.GetOutboundIP(),
		HostPort:                  4723,
		DiscoveryTunnelIdentifier: "1028068",
		HostType:                  common.OS(),
		HostUserID:                strconv.Itoa(common.UserInfo.UserID),
		DedicatedOrg:              strconv.Itoa(common.UserInfo.Organization.OrgID),
		Devices:                   devices,
	}
	for _, device := range devices {
		log.Println("Marking", device.UDID, device.Status, "...")
	}
	jsonInfo, err := json.Marshal(hostInfo)
	if err != nil {
		log.Println("Error marshaling data: ", err.Error())
		return
	}
	status, _, err := services.MakePostRequest(common.SyncEndpoint, jsonInfo)
	if err != nil {
		log.Println("Error while updating device info", err.Error())
		return
	}
	log.Println(status)
}

func (dw *DeviceWatcher) keepAlive() {
	for {
		time.Sleep(60 * time.Second)
		var devices []DeviceInfo
		for _, device := range dw.OldDevices {
			devices = append(devices, device)
		}
		dw.sync(true, devices)
	}
}

func (dw *DeviceWatcher) syncDiskImages(udid, version string) error {
	diskImagesPath := fmt.Sprintf("%s/%s", common.AppDirs.DiskImages, version)
	_, err := os.Stat(diskImagesPath)
	if err == nil {
		return nil
	}

	source := fmt.Sprintf("%s/diskimages/%s.zip", common.SanitisatioEndpoint, version)
	target := fmt.Sprintf("%s/%s.zip", common.AppDirs.DiskImages, version)
	common.Download(source, target)
	common.Unzip(target, common.AppDirs.DiskImages)

	_, err = common.Execute(fmt.Sprintf("%s image auto --basedir=%s/diskimages --udid %s", common.GoIOS, common.AppDirs.Assets, udid))
	return err
}
