package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strconv"

	"github.com/facebookgo/freeport"
)

var (
	version = "develop"
	lisPort = 8100
	udid    string
)

func NewReverseProxyHandlerFunc(targetURL *url.URL) http.HandlerFunc {
	httpProxy := httputil.NewSingleHostReverseProxy(targetURL)
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/" {
			io.WriteString(rw, `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>A static page</title>
  <link rel="stylesheet" href="/stylesheets/main.css">
</head>
<body>
  <h1>WDA Proxy</h1>
  <div>
  	<ul>
	  <li><a href="/inspector">Inspector</a></li>
	  <li><a href="/status">Status</a></li>
	</ul>
  </div>
</body>
</html>`)
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

	go func() {
		log.Printf("launch tcp-proxy, listen on %d", lisPort)
		targetURL, _ := url.Parse("http://127.0.0.1:" + strconv.Itoa(freePort))
		httpProxyFunc := NewReverseProxyHandlerFunc(targetURL)
		http.HandleFunc("/", httpProxyFunc)
		errC <- http.ListenAndServe(":"+strconv.Itoa(lisPort), nil)
	}()
	go func() {
		log.Printf("launch iproxy, device udid(%s)", udid)
		c := exec.Command("iproxy", strconv.Itoa(freePort), strconv.Itoa(lisPort), udid)
		errC <- c.Run()
	}()

	log.Fatal(<-errC)
}
