package instrument

import (
	"byod/common"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mholt/archiver/v3"
	plist "howett.net/plist"
)

var (
	optool      string
	tempDir     string
	dylibFolder string
)

type InfoPlistData struct {
	CFBundleExecutable string `plist:"CFBundleExecutable"`
	CFBundleIdentifier string `plist:"CFBundleIdentifier"`
}

func init() {
	optool = fmt.Sprintf("%s/optool", common.AppDirs.Assets)
	dylibFolder = fmt.Sprintf("%s/DYLIBS", common.AppDirs.Assets)
}

func InstrumentIPA(ipaPath string) (string, error) {
	if _, err := os.Stat(ipaPath); os.IsNotExist(err) {
		fmt.Printf("File %s not found or not readable\n", ipaPath)
		return "", errors.New("File " + ipaPath + " not found or not readable\n")
	}

	fmt.Println("[+] Starting instrumentation...")
	tempDir = ipaPath + ".cache"
	defer os.RemoveAll(tempDir)

	_, err := UnpackIPA(ipaPath)
	if err != nil {
		return "", err
	}
	appDir, err := getAppDirectory()
	if err != nil {
		return "", err
	}
	copyLibraryAndLoad(appDir)
	return repackIPA(ipaPath)
}

func UnpackIPA(ipa string) (string, error) {
	os.RemoveAll(tempDir)
	os.Mkdir(tempDir, 0755)

	zip := archiver.NewZip()
	err := zip.Unarchive(ipa, tempDir)
	if err != nil {
		fmt.Println("UnpackIPA: Couldn't unzip the IPA file:", err)
		return "", err
	}
	fmt.Println("[+] Unpacking the .ipa file DONE...")
	return "success", nil
}

func getAppDirectory() (string, error) {
	var appDir string
	payloadPath := filepath.Join(tempDir, "Payload")
	entries, err := os.ReadDir(payloadPath)
	if err != nil {
		fmt.Println("GetAppDirectory: Error reading Payload directory:", err)
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() && filepath.Ext(entry.Name()) == ".app" {
			appDir = filepath.Join(payloadPath, entry.Name())
			fmt.Println("[+] Found app directory as ", appDir)
			return appDir, nil
		}
	}
	if appDir == "" {
		fmt.Println("GetAppDirectory: No .app directory found in Payload")
		return "", errors.New("No .app directory found in Payload")
	}
	return "", nil
}

func copyLibraryAndLoad(appDir string) {
	dylibPath := filepath.Join(appDir, "Dylibs")
	os.Mkdir(dylibPath, 0755)
	copyDir(dylibFolder, dylibPath)
	appInfo, err := getAppInfo(appDir)
	if err != nil {
		fmt.Println("CopyLibraryAndLoad: Error getting app info:", err)
		return
	}
	appBinary := filepath.Join(appDir, appInfo.CFBundleExecutable)
	loadFrameworks(dylibPath, appBinary)
}

func repackIPA(originalIPA string) (string, error) {
	fmt.Println("[+] Repacking the .ipa")
	outputIPA := filepath.Base(originalIPA)
	outputIPA = outputIPA[:len(outputIPA)-len(filepath.Ext(outputIPA))] + "-patched.zip"
	outputPath := filepath.Dir(originalIPA)
	err := os.MkdirAll(outputPath, 0755)
	if err != nil {
		return "", fmt.Errorf("RepackIPA: failed to create output directory: %v", err)
	}
	fullOutputPath := filepath.Join(outputPath, outputIPA)
	os.Remove(fullOutputPath)
	err = archiver.Archive([]string{filepath.Join(tempDir, "Payload")}, fullOutputPath)
	if err != nil {
		fmt.Println(err.Error())
		return "", fmt.Errorf("RepackIPA: failed to compress the app into an .ipa file: %v", err)
	}
	ext := filepath.Ext(fullOutputPath)
	finalOutput := fullOutputPath[:len(fullOutputPath)-len(ext)] + ".ipa"
	os.Rename(fullOutputPath, finalOutput)
	fmt.Println("[+] Wrote", finalOutput)
	return finalOutput, nil
}

func loadFrameworks(dir string, appBinary string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && filepath.Ext(path) == ".framework" {
			frameworkName := filepath.Base(path)
			binaryName := frameworkName[:len(frameworkName)-len(filepath.Ext(frameworkName))]
			binaryPath := filepath.Join(path, binaryName)
			if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
				fmt.Printf("LoadFrameworks: Binary not found for %s\n", frameworkName)
			} else {
				cmd := exec.Command(optool, "install", "-c", "load", "-p", "@executable_path/Dylibs/"+frameworkName+"/"+binaryName, "-t", appBinary)
				if err := cmd.Run(); err != nil {
					fmt.Println("LoadFrameworks: Failed to inject ", binaryName, "into", appBinary, ":", err)
				}
			}
		}
		return nil
	})
}

func getAppInfo(appDir string) (InfoPlistData, error) {
	var data InfoPlistData
	infoPlistPath := filepath.Join(appDir, "Info.plist")
	file, err := os.Open(infoPlistPath)
	if err != nil {
		return data, err
	}
	defer file.Close()
	decoder := plist.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func copyDir(srcDir, destDir string) error {
	// Create the destination directory
	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		return err
	}

	// Read entries from the source directory
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())
		if entry.IsDir() {
			err := copyDir(srcPath, destPath)
			if err != nil {
				return err
			}
		} else {
			err := copyFile(srcPath, destPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}
