package main

import (
	"bytes"
	"os/exec"

	plist "github.com/DHowett/go-plist"
)

type DeviceInfo struct {
	MacAddress      string `plist:"EthernetAddress" json:"-"`      // eg: a0:11:28:31:42:21
	ProductName     string `plist:"ProductName" json:"-"`          // eg: iPhone OS
	HardwareModel   string `plist:"HardwareModel" json:"-"`        // eg: N56AP
	ProductType     string `plist:"ProductType" json:"model"`      // eg: iPhone7,1
	Version         string `plist:"ProductVersion" json:"version"` // eg: 10.2
	CPUArchitecture string `plist:"CPUArchitecture" json:"abi"`    // eg: arm64
	Udid            string `plist:"UniqueDeviceID" json:"serial"`
	Manufacturer    string `json:"manufacturer"`
}

func GetDeviceInfo(udid string) (v DeviceInfo, err error) {
	c := exec.Command("ideviceinfo", "--udid", udid, "--xml")
	output, err := c.Output()
	if err != nil {
		return
	}
	v.Manufacturer = "Apple"
	err = plist.NewDecoder(bytes.NewReader(output)).Decode(&v)
	return
}
