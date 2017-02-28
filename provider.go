package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/openatx/wdaproxy/connector"
	"github.com/qiniu/log"
)

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

	rt.HandleFunc("/devices/{udid}", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Not finished yet")
	})
}
