package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strconv"

	"encoding/json"

	"strings"

	"github.com/facebookgo/freeport"
)

var (
	version = "develop"
	lisPort = 8100
	udid    string

	indexContent = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>A static page</title>
  <style>
  body {
	  padding: 50px 80px;
	  margin: 0px;
  }
  pre {
	  font-family: "Courier New";
	  padding: 10px;
	  border: 1px solid gray;
  }
  </style>
</head>
<body>
  <h1>WDA Proxy</h1>
  <div>
  	<ul>
	  <li><a href="/inspector">Inspector</a></li>
	  <li><a href="/status">Status</a></li>
	</ul>
	<pre id="status"></pre>
  </div>
</body>
<script src="//cdn.jsdelivr.net/jquery/3.1.1/jquery.min.js"></script>
<script>
var pre = document.getElementById("status");
$.ajax({
	url: "/status",
	dataType: "json",
	success: function(ret){
		pre.innerHTML = JSON.stringify(ret, null, "    ");
	},
	error: function(xhr, status, err){
		pre.innerHTML = xhr.status + " " + err;
		pre.style.color = "red";
	}
})
</script>
</html>`
)

type statusResp struct {
	Value     map[string]interface{} `json:"value,omitempty"`
	SessionId string                 `json:"sessionId,omitempty"`
	Status    int                    `json:"status"`
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

func NewReverseProxyHandlerFunc(targetURL *url.URL) http.HandlerFunc {
	httpProxy := httputil.NewSingleHostReverseProxy(targetURL)
	httpProxy.Transport = &transport{http.DefaultTransport}
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/" {
			io.WriteString(rw, indexContent)
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
