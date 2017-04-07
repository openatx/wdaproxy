package muuid

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/satori/go.uuid"
)

var (
	ErrUuidNotFound = errors.New("uuid not found")
)

func UUID() string {
	return UUIDFromOS(runtime.GOOS)
}

func UUIDFromOS(osName string) string {
	var uuid string
	var err error

	switch osName {
	case "darwin":
		uuid, err = osxUUID()
	case "linux":
		uuid, err = linuxUUID()
	case "windows":
		uuid, err = winUUID()
	}
	if err != nil || uuid == "" {
		uuid = defaultUuid()
	}
	return uuid
}

func osxUUID() (string, error) {
	c := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := c.Output()
	if err != nil {
		return "", err
	}
	pattern := regexp.MustCompile(`IOPlatformUUID" = "(.*?)"`)
	ss := pattern.FindStringSubmatch(string(output))
	if len(ss) == 0 {
		return "", ErrUuidNotFound
	}
	return ss[1], nil
}

func linuxUUID() (string, error) {
	data, err := ioutil.ReadFile("/var/lib/dbus/machine-id")
	if err != nil {
		return "", ErrUuidNotFound
	}
	id := strings.TrimSpace(string(data))
	if len(id) > 20 {
		return id[0:8] + "-" + id[8:12] + "-" + id[12:16] + "-" + id[16:20] + "-" + id[20:], nil
	}
	return "", ErrUuidNotFound
}

func winUUID() (string, error) {
	c := exec.Command("wmic", "CsProduct", "Get", "UUID")
	output, err := c.Output()
	if err != nil {
		return "", err
	}
	pattern := regexp.MustCompile(`([\w\d]+)-([\w\d]+)-([\w\d]+)-([\w\d]+)-([\w\d]+)`)
	s := pattern.FindString(string(output))
	return s, nil
}

func getMachineUidPath() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".muid")
}

func defaultUuid() string {
	filePath := getMachineUidPath()
	id := ""
	data, err := ioutil.ReadFile(filePath)
	if err != nil || strings.TrimSpace(string(data)) == "" {
		id = fmt.Sprintf("%s", uuid.NewV4())
		ioutil.WriteFile(filePath, []byte(id), 0644)
	} else {
		return strings.TrimSpace(string(data))
	}
	return id
}

func RemoveTempUidFile() error {
	return os.Remove(getMachineUidPath())
}
