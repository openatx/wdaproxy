package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/openatx/wdaproxy/connector"
	"github.com/openatx/wdaproxy/web"
	"github.com/qiniu/log"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	koardPower = NewKoardPower("/dev/tty.usbmodem1471")
)

func init() {
	rt.HandleFunc("/devices/{udid}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", 302)
		// io.WriteString(w, "Not finished yet")
	})

	rt.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("index").Parse(assetsContent("/index.html")))
		t.Execute(w, nil)
	})

	rt.HandleFunc("/packages", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("pkgs").Delims("[[", "]]").Parse(assetsContent("/packages.html")))
		t.Execute(w, nil)
	})

	rt.HandleFunc("/remote", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("pkgs").Delims("[[", "]]").Parse(assetsContent("/remote-control.html")))
		t.Execute(w, nil)
	})

	rt.PathPrefix("/res/").Handler(http.StripPrefix("/res/", http.FileServer(web.Assets)))
	rt.PathPrefix("/recorddata/").Handler(http.StripPrefix("/recorddata/", http.FileServer(http.Dir("./recorddata"))))

	rt.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		favicon, _ := web.Assets.Open("images/favicon.ico")
		http.ServeContent(w, r, "favicon.ico", time.Now(), favicon)
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
	}).Methods("GET")

	rt.HandleFunc("/api/v1/packages/{bundleId}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		bundleId := mux.Vars(r)["bundleId"]
		output, err := UninstallPackage(udid, bundleId)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     err == nil,
			"description": output,
		})
	}).Methods("DELETE")

	rt.HandleFunc("/api/v1/packages", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(0)
		defer r.MultipartForm.RemoveAll()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		renderError := func(err error, description string) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"description": fmt.Sprintf("%s: %v", description, err),
			})
		}
		var reader io.Reader
		url := r.FormValue("url")
		if url != "" {
			resp, err := http.Get(url)
			if err != nil {
				renderError(err, "download from url")
				return
			}
			reader = resp.Body
			defer resp.Body.Close()
		} else {
			file, _, err := r.FormFile("file")
			if err != nil {
				renderError(err, "parse form 'file'")
				return
			}
			reader = file
			defer file.Close()
		}
		os.Mkdir("uploads", 0755)
		tmpfile, err := ioutil.TempFile("uploads", "tempipa-")
		if err != nil {
			renderError(err, "create tmpfile")
			return
		}
		defer os.Remove(tmpfile.Name())

		log.Println("[pkg] create tmpfile", tmpfile.Name())
		_, err = io.Copy(tmpfile, reader)
		if err != nil {
			renderError(err, "read upload file")
			return
		}
		if err := tmpfile.Close(); err != nil {
			renderError(err, "finish write tmpfile")
			return
		}
		log.Println("[pkg] install ipa")
		cmd := exec.Command("ideviceinstaller", "--udid", udid, "-i", tmpfile.Name())
		output, err := cmd.CombinedOutput()
		if err != nil {
			renderError(errors.New(string(output)), "install ipa")
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"description": "Successfully installed ipa",
			"value":       string(output),
		})
	}).Methods("POST")

	rt.HandleFunc("/ws/admin", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade to websocket error:", err)
			return
		}
		defer conn.Close()

		wchan := make(chan string)
		go func() {
			for m := range wchan {
				conn.WriteMessage(websocket.TextMessage, []byte(m))
			}
		}()

		// unit: mA
		powerHook := func(current float32) error {
			wchan <- fmt.Sprintf("%.2f", current)
			return nil
		}
		defer func() {
			koardPower.RemoveListener(&powerHook)
			close(wchan)
		}()

		for {
			mtype, p, err := conn.ReadMessage()
			log.Println(mtype, string(p), err)
			switch string(p) {
			case "current-on":
				koardPower.AddListener(&powerHook)
			case "current-off":
				koardPower.RemoveListener(&powerHook)
			}
			if string(p) == "hello" {
				wchan <- "world"
			}
		}
	})

	rt.HandleFunc("/api/v1/records", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.Method == "POST" {
			os.MkdirAll("./recorddata/20170603-test", 0755)
			err := launchXrecord("./recorddata/20170603-test/camera.mp4")
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     false,
					"description": "Launch xrecord failed: " + err.Error(),
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     true,
					"description": "Record started",
				})
			}
		} else {
			if xrecordCmd != nil && xrecordCmd.Process != nil {
				xrecordCmd.Process.Signal(os.Interrupt)
			}
			select {
			case <-GoFunc(xrecordCmd.Wait):
			case <-time.After(5 * time.Second):
				xrecordCmd.Process.Kill()
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     false,
					"description": "xrecord handle Ctrl-C longer than 5 second",
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     true,
				"description": "Record stopped",
			})
		}
	}).Methods("POST", "DELETE")
}

func mockIOSProvider() {
	c := connector.New(yosemiteServer, yosemiteGroup, lisPort)
	go c.KeepOnline()

	device, err := GetDeviceInfo(udid)
	if err != nil {
		log.Fatal(err)
	}

	c.AddDevice(device.Udid, device)
	// c.WriteJSON(map[string]interface{}{
	// 	"type": "addDevice",
	// 	"data": device,
	// })
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

var xrecordCmd *exec.Cmd

func launchXrecord(output string) error {
	rpath, err := exec.LookPath("xrecord")
	if err != nil {
		return err
	}
	xrecordCmd = exec.Command(rpath, "-i", "0x14100000046d082d", "-o", output, "-f")
	xrecordCmd.Stdout = os.Stdout
	xrecordCmd.Stderr = os.Stderr
	return xrecordCmd.Start()
}

func GoFunc(f func() error) chan error {
	errc := make(chan error)
	go func() {
		errc <- f()
	}()
	return errc
}
