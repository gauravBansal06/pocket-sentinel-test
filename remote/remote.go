package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var (
	env              = "prod"
	tunnelBinaryPath = "./LT"
	tunnelProcess    *os.Process
	tunnelInfo       TunnelInfo
)

func SetTunnelArgs(tunnelPath, envPass string) {
	if envPass != "" {
		env = envPass
	}
	if tunnelPath != "" {
		tunnelBinaryPath = tunnelPath
	}
}

type TunnelInfo struct {
	Status string     `json:"status"`
	Data   TunnelData `json:"data"`
}

type TunnelData struct {
	ID             int32  `json:"id"`
	LocalProxyPort string `json:"localProxyPort"`
	User           string `json:"user"`
	TunnelName     string `json:"tunnelName"`
	Environment    string `json:"environment"`
	Version        string `json:"version"`
}

func GetTunnelId() (string, error) {
	if tunnelInfo.Status != "FAILED" && tunnelInfo.Data.ID > 0 {
		return fmt.Sprintf("%v", tunnelInfo.Data.ID), nil
	}

	log.Println("fetching tunnel info")
	resp, err := http.Get("http://127.0.0.1:8000/api/v1.0/info")
	if err != nil {
		log.Println("failed to get tunnel information", err)
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("failed to read tunnel information", err)
		return "", err
	}

	tunnelInfo = TunnelInfo{}
	err = json.Unmarshal(body, &tunnelInfo)
	if err != nil {
		log.Println("failed to parse tunnel information", err)
		return "", err
	}
	if tunnelInfo.Status == "FAILED" {
		return "", errors.New("tunnel info not found")
	}
	log.Printf("tunnel information: %v\n", tunnelInfo)
	return fmt.Sprintf("%v", tunnelInfo.Data.ID), nil
}

func LaunchTunnel(user, key string) {
	infoAPIPort := "8000"

	_, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", infoAPIPort), time.Second)
	if err == nil {
		err = fmt.Errorf("port (%s) is busy ... can't start binary tunnel, make sure that port 9090, 8000 and 4723 is free", infoAPIPort)
		log.Printf("%v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command(tunnelBinaryPath, "--user", user, "--key", key, "--infoAPIPort", infoAPIPort)
	if env == "stage" {
		cmd = exec.Command(tunnelBinaryPath, "--user", user, "--key", key, "--infoAPIPort", infoAPIPort, "--env", "stage")
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("failed to start tunnel: %v\n", err)
		os.Exit(1)
	}

	pid := cmd.Process.Pid
	log.Printf("tunnel started, pid %v\n", pid)

	tunnelProcess = cmd.Process
}

func KillTunnel() {
	log.Println("killing tunnel")
	if tunnelProcess == nil {
		log.Println("tunnel already not started")
		return
	}
	err := tunnelProcess.Kill()
	if err != nil {
		log.Println("error killing tunnel process: ", err)
	}
}
