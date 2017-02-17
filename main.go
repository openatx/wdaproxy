//go:generate go run web/assets_generate.go

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	"html/template"

	"github.com/facebookgo/freeport"
	"github.com/openatx/wdaproxy/web"
)

var (
	version = "develop"
	lisPort = 8100
	udid    string
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

type transport struct {
	http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// rewrite url
	if strings.HasPrefix(req.RequestURI, "/origin/") {
		req.URL.Path = req.RequestURI[len("/origin"):]
		return t.RoundTripper.RoundTrip(req)
	}

	// request
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// rewrite body
	if req.RequestURI == "/status" {
		jsonResp := &statusResp{}
		err = json.NewDecoder(resp.Body).Decode(jsonResp)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		jsonResp.Value["udid"] = udid
		data, _ := json.Marshal(jsonResp)
		// update body and fix length
		resp.Body = ioutil.NopCloser(bytes.NewReader(data))
		resp.ContentLength = int64(len(data))
		resp.Header.Set("Content-Length", strconv.Itoa(len(data)))
		return resp, nil
	}
	return resp, nil
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

func NewReverseProxyHandlerFunc(targetURL *url.URL) http.HandlerFunc {
	httpProxy := httputil.NewSingleHostReverseProxy(targetURL)
	httpProxy.Transport = &transport{http.DefaultTransport}
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/" {
			t := template.Must(template.New("index").Parse(assetsContent("/index.html")))
			t.Execute(rw, nil)
			// io.WriteString(rw, indexContent)
			return
		}
		if r.RequestURI == "/packages" {
			rw.Header().Set("Content-Type", "application/json; charset=utf-8")
			c := exec.Command("ideviceinstaller", "-l", "--udid", getUdid())
			out, err := c.Output()
			if err != nil {
				json.NewEncoder(rw).Encode(map[string]interface{}{
					"status": 1,
					"value":  err.Error(),
				})
				return
			}
			bufrd := bufio.NewReader(bytes.NewReader(out))
			bufrd.ReadLine() // ignore first line
			packages := make([]packageInfo, 0)
			for {
				bline, _, er := bufrd.ReadLine()
				if er != nil {
					break
				}
				fields := strings.Split(string(bline), ", ")
				if len(fields) != 3 {
					continue
				}
				version, _ := strconv.Unquote(fields[1])
				name, _ := strconv.Unquote(fields[2])
				packages = append(packages, packageInfo{fields[0], name, version})
			}

			json.NewEncoder(rw).Encode(map[string]interface{}{
				"status": 0,
				"value":  packages,
			})
			return
		}
		httpProxy.ServeHTTP(rw, r)
	}
}

func main() {
	showVer := flag.Bool("v", false, "Print version")
	flag.IntVar(&lisPort, "p", 8100, "Proxy listen port")
	flag.StringVar(&udid, "u", "", "device udid")
	flag.Parse()

	if *showVer {
		println(version)
		return
	}

	log.Println("program start......")
	errC := make(chan error)
	freePort, err := freeport.Get()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("get freeport %d", freePort)

	go func() {
		log.Printf("launch tcp-proxy, listen on %d", lisPort)
		targetURL, _ := url.Parse("http://127.0.0.1:" + strconv.Itoa(freePort))
		httpProxyFunc := NewReverseProxyHandlerFunc(targetURL)
		http.HandleFunc("/", httpProxyFunc)
		errC <- http.ListenAndServe(":"+strconv.Itoa(lisPort), nil)
	}()
	go func() {
		log.Printf("launch iproxy, device udid: %s", strconv.Quote(udid))
		c := exec.Command("iproxy", strconv.Itoa(freePort), "8100")
		if udid != "" {
			c.Args = append(c.Args, udid)
		}
		errC <- c.Run()
	}()

	log.Fatal(<-errC)
}
