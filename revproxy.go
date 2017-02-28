package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
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
		httpProxy.ServeHTTP(rw, r)
	}
}
