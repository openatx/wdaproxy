package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strconv"
)

var (
	version = "develop"
	lisPort = 8100
)

func main() {
	showVer := flag.Bool("v", false, "Print version")
	flag.IntVar(&lisPort, "p", 8100, "Proxy listen port")
	flag.Parse()

	if *showVer {
		println(version)
		return
	}

	log.Println("program start......")
	errC := make(chan error)

	go func() {
		log.Println("launch tcp-proxy")
		targetURL, _ := url.Parse("http://127.0.0.1:8200")
		httpProxy := httputil.NewSingleHostReverseProxy(targetURL)
		http.Handle("/", httpProxy)
		errC <- http.ListenAndServe(":"+strconv.Itoa(lisPort), nil)
	}()
	go func() {
		log.Println("launch iproxy")
		c := exec.Command("iproxy", "8200", strconv.Itoa(lisPort))
		errC <- c.Run()
	}()

	log.Fatal(<-errC)
}
