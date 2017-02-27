package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
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
