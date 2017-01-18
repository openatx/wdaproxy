package main

import (
	"flag"
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
		httpProxy := httputil.NewSingleHostReverseProxy(targetURL)
		http.Handle("/", httpProxy)
		errC <- http.ListenAndServe(":"+strconv.Itoa(lisPort), nil)
	}()
	go func() {
		log.Printf("launch iproxy, device udid(%s)", udid)
		c := exec.Command("iproxy", strconv.Itoa(freePort), strconv.Itoa(lisPort), udid)
		errC <- c.Run()
	}()

	log.Fatal(<-errC)
}
