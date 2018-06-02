package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

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
	if req.URL.Path == "/status" {
		jsonResp := &statusResp{}
		err = json.NewDecoder(resp.Body).Decode(jsonResp)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		jsonResp.Value["device"] = map[string]interface{}{
			"udid": udid,
			"name": udidNames[udid],
		}
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
		httpProxy.ServeHTTP(rw, r)
	}
}

type fakeProxy struct {
}

func (p *fakeProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("FAKE", r.RequestURI)
	log.Println("P", r.URL.Path)
	io.WriteString(w, "Fake")
}

func NewAppiumProxyHandlerFunc(targetURL *url.URL) http.HandlerFunc {
	httpProxy := httputil.NewSingleHostReverseProxy(targetURL)
	rt := mux.NewRouter()
	rt.HandleFunc("/wd/hub/sessions", func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"status":    0,
			"value":     []string{},
			"sessionId": nil,
		}, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.Write(data)
	})
	rt.HandleFunc("/wd/hub/session/{sessionId}/window/current/size", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, "/current/size", "/size", -1)
		r.URL.Path = r.URL.Path[len("/wd/hub"):]
		httpProxy.ServeHTTP(w, r)
	})
	rt.Handle("/wd/hub/{subpath:.*}", http.StripPrefix("/wd/hub", httpProxy))
	return rt.ServeHTTP
}
