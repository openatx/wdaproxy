package main

import (
	"bytes"
	"os/exec"

	plist "github.com/DHowett/go-plist"
)

type Package struct {
	Name             string   `plist:"CFBundleDisplayName" json:"name"`
	BundleId         string   `plist:"CFBundleIdentifier" json:"bundleId"`
	Version          string   `plist:"CFBundleVersion" json:"version"`
	MinOSVersion     string   `plist:"MinimumOSVersion" json:"miniOSVersion"`
	SupportedDevices []string `plist:"UISupportedDevices" json:"UISupportedDevices"`
}

func ListPackages(udid string) (pkgs []Package, err error) {
	pkgs = make([]Package, 0)
	c := exec.Command("ideviceinstaller", "--udid", udid, "-l", "-o", "xml")
	data, err := c.Output()
	if err != nil {
		return nil, err
	}
	err = plist.NewDecoder(bytes.NewReader(data)).Decode(&pkgs)
	return
}

func UninstallPackage(udid, bundleId string) (output string, err error) {
	c := exec.Command("ideviceinstaller", "--udid", udid, "--uninstall", bundleId)
	data, err := c.CombinedOutput()
	return string(data), err
}
