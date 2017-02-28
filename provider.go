package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openatx/wdaproxy/connector"
	"github.com/qiniu/log"
)

func init() {
	rt.HandleFunc("/devices/{udid}", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Not finished yet")
	})

	rt.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("index").Parse(assetsContent("/index.html")))
		t.Execute(w, nil)
	})

	rt.HandleFunc("/packages", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("index").Delims("[[", "]]").Parse(assetsContent("/packages.html")))
		t.Execute(w, nil)
	})
	v1Rounter(rt)
}

func v1Rounter(rt *mux.Router) {
	rt.HandleFunc("/api/v1/packages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		pkgs, err := ListPackages(udid)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"description": err.Error(),
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"value":   pkgs,
		})
	})

	rt.HandleFunc("/api/v1/packages/{bundleId}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		bundleId := mux.Vars(r)["bundleId"]
		output, err := UninstallPackage(udid, bundleId)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     err == nil,
			"description": output,
		})
	}).Methods("DELETE")
}

func mockIOSProvider() {
	c := connector.New(yosemiteServer, yosemiteGroup, lisPort)
	go c.KeepOnline()

	device, err := GetDeviceInfo(udid)
	if err != nil {
		log.Fatal(err)
	}

	c.WriteJSON(map[string]interface{}{
		"type": "addDevice",
		"data": device,
	})
	rt.HandleFunc("/api/devices/{udid}/remoteConnect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.Method == "POST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":          true,
				"description":      "notice this is mock data",
				"remoteConnectUrl": fmt.Sprintf("http://%s:%d/", c.RemoteIp, lisPort),
			})
		}
		if r.Method == "DELETE" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     true,
				"description": "Device remote disconnected successfully",
			})
		}
	}).Methods("POST", "DELETE")
}
