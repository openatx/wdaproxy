package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/facebookgo/freeport"
	"github.com/gorilla/mux"
	accesslog "github.com/mash/go-accesslog"
	flag "github.com/ogier/pflag"
	"github.com/openatx/wdaproxy/web"
	"github.com/qiniu/log"
	_ "github.com/shurcooL/vfsgen"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

var (
	version        = "develop"
	lisPort        = 8100
	pWda           string
	udid           string
	yosemiteServer string
	yosemiteGroup  string
	debug          bool

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

// LocalIP returns the non loopback local IP of the host
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func main() {
	showVer := flag.BoolP("version", "v", false, "Print version")
	flag.IntVarP(&lisPort, "port", "p", 8100, "Proxy listen port")
	flag.StringVarP(&udid, "udid", "u", "", "device udid")
	flag.StringVarP(&pWda, "wda", "W", "", "WebDriverAgent project directory [optional]")
	flag.BoolVarP(&debug, "debug", "d", false, "Open debug mode")

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

	if *showVer {
		println(version)
		return
	}

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(lisPort))
	if err != nil {
		log.Fatal(err)
	}

	if yosemiteServer != "" {
		mockIOSProvider()
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
		errC <- http.Serve(lis, accesslog.NewLoggingHandler(rt, HTTPLogger{}))
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
				log.Fatal("[WDA] exit", err)
			}

			if isPrefix {
				lineStr = lineStr + string(line)
				continue
			} else {
				lineStr = string(line)
			}
			lineStr := strings.TrimSpace(string(line))
			if debug {
				fmt.Printf("[WDA] %s\n", lineStr)
			}
			if strings.Contains(lineStr, "Successfully wrote Manifest cache to") {
				log.Println("[WDA] test ipa successfully generated")
			}
			if strings.HasPrefix(lineStr, "Test Case '-[UITestingUITests testRunner]' started") {
				log.Println("[WDA] successfully started")
			}
			lineStr = "" // reset str
		}
	}()

	log.Printf("Open webbrower with http://%s:%d", LocalIP(), lisPort)
	log.Fatal(<-errC)
}
