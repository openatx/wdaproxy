package connector

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/codeskyblue/muuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	qlog "github.com/qiniu/log"
)

var log = qlog.New(os.Stdout, "", qlog.Llevel|qlog.Lshortfile|qlog.LstdFlags)

const (
	ActionInit          = "init"
	ActionDeviceAdd     = "addDevice"
	ActionDeviceRemove  = "removeDevice"
	ActionDeviceRelease = "releaseDevice"
)

type Connector struct {
	ws         *websocket.Conn
	host       string
	listenPort int
	msgC       chan interface{}

	Id      string `json:"id"`
	Name    string `json:"name"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Group   string `json:"group"`
	Address string `json:"address"`

	RemoteIp string `json:"-"`
	devices  map[string]interface{}
}

func New(host string, group string, listenPort int) *Connector {
	c := &Connector{
		host:       host,
		msgC:       make(chan interface{}),
		Group:      group,
		listenPort: listenPort,
		Id:         muuid.UUID() + ":" + strconv.Itoa(listenPort),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		devices:    make(map[string]interface{}),
	}
	c.Name, _ = os.Hostname()
	return c
}

func (w *Connector) KeepOnline() {
	if w.host == "" {
		return
	}
	for {
		w.keepOnline()
		log.Println("Retry connect to center after 3.0s")
		time.Sleep(3 * time.Second)
	}
}

func (w *Connector) keepOnline() error {
	u := url.URL{
		Scheme: "ws",
		Host:   w.host,
		Path:   "/websocket",
	}
	log.Printf("connecting to %s", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "dial websocket")
	}
	w.ws = ws
	defer ws.Close()

	// Step 1: get remote ip
	var t = struct {
		RemoteIp string `json:"remoteIp"`
	}{}
	if err := ws.ReadJSON(&t); err != nil {
		return errors.Wrap(err, "read remoteIp")
	}
	w.RemoteIp = t.RemoteIp
	w.Address = fmt.Sprintf("http://%s:%d", t.RemoteIp, w.listenPort)

	// Step 2: send provider info
	err = ws.WriteJSON(map[string]interface{}{
		"type": ActionInit,
		"data": w,
	})
	if err != nil {
		return errors.Wrap(err, "send init data")
	}

	done := make(chan bool, 1)
	go w.keepPing(done)
	defer func() {
		done <- true
	}()

	// resend registed device data
	for _, device := range w.devices {
		w.Do(ActionDeviceAdd, device)
	}

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			return errors.Wrap(err, "read message")
		}
		log.Printf("recv: %s", message)
	}
	return nil
}

func (w *Connector) keepPing(done chan bool) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := w.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Warnf("send ping: %v", err)
				return
			}
		case msg := <-w.msgC:
			if err := w.ws.WriteJSON(msg); err != nil {
				log.Warnf("send data: %v", err)
				return
			}
		case <-done:
			return
		}
	}
}

func (w *Connector) WriteJSON(v interface{}) {
	w.msgC <- v
}

func (w *Connector) Do(action string, data interface{}) {
	w.msgC <- map[string]interface{}{
		"type": action,
		"data": data,
	}
}

func (w *Connector) AddDevice(id string, device interface{}) {
	w.Do(ActionDeviceAdd, device)
	w.devices[id] = device
}

// func (w *Connector) ReleaseDevice(serial string, oneOffToken string) {
// 	w.msgC <- map[string]interface{}{
// 		"type": "releaseDevice",
// 		"data": map[string]string{
// 			"serial":      serial,
// 			"oneOffToken": oneOffToken,
// 		},
// 	}
// }
