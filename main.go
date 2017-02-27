//go:generate go run web/assets_generate.go

package main

import (
	"bufio"
	"encoding/json"
	"path/filepath"
	// "flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/facebookgo/freeport"
	"github.com/gorilla/mux"
	accesslog "github.com/mash/go-accesslog"
	flag "github.com/ogier/pflag"
	"github.com/openatx/wdaproxy/connector"
	"github.com/openatx/wdaproxy/web"
)

var (
	version        = "develop"
	lisPort        = 8100
	pWda           string
	udid           string
	yosemiteServer string
	yosemiteGroup  string

	rt = mux.NewRouter()
)

type statusResp struct {
	Value     map[string]interface{} `json:"value,omitempty"`
	SessionId string                 `json:"sessionId,omitempty"`
	Status    int                    `json:"status"`
}

func getUdid() string {
	if udid != "" {
		return udid
	}
	output, err := exec.Command("idevice_id", "-l").Output()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(output))
}

type packageInfo struct {
	BundleId string `json:"bundleId"`
	Name     string `json:"name"`
	Version  string `json:"version"`
}

func assetsContent(name string) string {
	fd, err := web.Assets.Open(name)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		panic(err)
	}
	return string(data)
}

type Device struct {
	Udid         string `json:"serial"`
	Manufacturer string `json:"manufacturer"`
}

func mockIOSProvider() {
	c := connector.New(yosemiteServer, yosemiteGroup, lisPort)
	go c.KeepOnline()

	device, err := GetDeviceInfo(getUdid())
	if err != nil {
		log.Fatal(err)
	}

	c.WriteJSON(map[string]interface{}{
		"type": "addDevice",
		"data": device,
	})

	rt.HandleFunc("/api/devices/{udid}/remoteConnectUrl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":          true,
				"description":      "notice this is mock data",
				"remoteConnectUrl": fmt.Sprintf("http://%s:%d/", c.RemoteIp, lisPort),
			})
		}
		if r.Method == "DELETE" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     true,
				"description": "Device remote disconnected successfully",
			})
		}
	})

	rt.HandleFunc("/devices/{udid}", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Not finished yet")
	})
}

func main() {
	showVer := flag.BoolP("version", "v", false, "Print version")
	flag.IntVarP(&lisPort, "port", "p", 8100, "Proxy listen port")
	flag.StringVarP(&udid, "udid", "u", "", "device udid")
	flag.StringVarP(&pWda, "wda", "W", "", "WebDriverAgent project directory [optional]")

	flag.StringVarP(&yosemiteServer, "yosemite-server", "S",
		os.Getenv("YOSEMITE_SERVER"),
		"server center(not open source yet")
	flag.StringVarP(&yosemiteGroup, "yosemite-group", "G",
		"everyone",
		"server center group")
	flag.Parse()
	if udid == "" {
		udid = getUdid()
	}

	mockIOSProvider()

	if *showVer {
		println(version)
		return
	}

	errC := make(chan error)
	freePort, err := freeport.Get()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("freeport %d", freePort)

	go func() {
		log.Printf("launch tcp-proxy, listen on %d", lisPort)
		targetURL, _ := url.Parse("http://127.0.0.1:" + strconv.Itoa(freePort))
		rt.HandleFunc("/{path:.*}", NewReverseProxyHandlerFunc(targetURL))
		errC <- http.ListenAndServe(":"+strconv.Itoa(lisPort), accesslog.NewLoggingHandler(rt, HTTPLogger{}))
	}()
	go func() {
		log.Printf("launch iproxy (udid: %s)", strconv.Quote(udid))
		c := exec.Command("iproxy", strconv.Itoa(freePort), "8100")
		if udid != "" {
			c.Args = append(c.Args, udid)
		}
		errC <- c.Run()
	}()
	go func() {
		if pWda == "" {
			return
		}
		log.Printf("launch WebDriverAgent(dir=%s)", pWda)
		c := exec.Command("xcodebuild",
			"-verbose",
			"-project", "WebDriverAgent.xcodeproj",
			"-scheme", "WebDriverAgentRunner",
			"-destination", "id="+udid, "test")
		c.Dir, _ = filepath.Abs(pWda)
		// Test Suite 'All tests' started at 2017-02-27 15:55:35.263
		// Test Suite 'WebDriverAgentRunner.xctest' started at 2017-02-27 15:55:35.266
		// Test Suite 'UITestingUITests' started at 2017-02-27 15:55:35.267
		// Test Case '-[UITestingUITests testRunner]' started.
		// t =     0.00s     Start Test at 2017-02-27 15:55:35.270
		// t =     0.01s     Set Up
		pipeReader, writer := io.Pipe()
		c.Stdout = writer
		c.Stderr = writer
		c.Stdin = os.Stdin

		bufrd := bufio.NewReader(pipeReader)
		if err = c.Start(); err != nil {
			log.Fatal(err)
		}
		lineStr := ""
		for {
			line, isPrefix, err := bufrd.ReadLine()
			if err != nil {
				log.Fatal(err)
			}

			if isPrefix {
				lineStr = lineStr + string(line)
				continue
			} else {
				lineStr = string(line)
			}
			lineStr := strings.TrimSpace(string(line))
			// log.Println("WWW:", lineStr)
			if strings.Contains(lineStr, "Successfully wrote Manifest cache to") {
				log.Println("[WDA] test ipa successfully generated")
			}
			if strings.HasPrefix(lineStr, "Test Case '-[UITestingUITests testRunner]' started") {
				log.Println("[WDA] successfully started")
			}
			lineStr = "" // reset str
		}
	}()

	log.Fatal(<-errC)
}
