package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
)

func main() {
	log.Println("program start......")
	errC := make(chan error)

	go func() {
		log.Println("launch tcp-proxy")
		targetURL, _ := url.Parse("http://127.0.0.1:8200")
		httpProxy := httputil.NewSingleHostReverseProxy(targetURL)
		http.Handle("/", httpProxy)
		errC <- http.ListenAndServe(":8100", nil)
	}()
	go func() {
		log.Println("launch iproxy")
		c := exec.Command("iproxy", "8200", "8100")
		errC <- c.Run()
	}()

	log.Fatal(<-errC)
}
