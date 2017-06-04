package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type KoradPowerHookFunc *func(float32) error

type KoradPower struct {
	listeners map[KoradPowerHookFunc]bool
	mu        sync.Mutex
	ticker    *time.Ticker
	portName  string
}

func NewKoardPower(portName string) *KoradPower {
	return &KoradPower{
		portName:  portName,
		listeners: make(map[KoradPowerHookFunc]bool),
	}
}

func (k *KoradPower) AddListener(pf KoradPowerHookFunc) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listeners[pf] = true
	if len(k.listeners) == 1 {
		k.start()
	}
}

func (k *KoradPower) RemoveListener(pf KoradPowerHookFunc) {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.listeners, pf)
	if len(k.listeners) == 0 {
		k.stop()
	}
}

func (k *KoradPower) broadcast(v float32) {
	k.mu.Lock()
	defer k.mu.Unlock()

	for pf := range k.listeners {
		if err := (*pf)(v); err != nil {
			delete(k.listeners, pf)
		}
	}
}

func (k *KoradPower) start() {
	s, err := serial.Open(serial.OpenOptions{
		PortName:        k.portName, // only in mac
		BaudRate:        9600,
		StopBits:        1,
		DataBits:        8,
		MinimumReadSize: 4,
	})
	if err != nil {
		log.Println("com communicate error:", err)
		return
	}
	k.ticker = time.NewTicker(time.Second)

	go func() {
		defer s.Close()
		buf := make([]byte, 10)
		for range k.ticker.C {
			s.Write([]byte("IOUT1?")) // current
			time.Sleep(100 * time.Millisecond)
			n, err := s.Read(buf)
			if err != nil {
				continue
			}
			var current float32
			fmt.Sscanf(string(buf[0:n]), "%f", &current)

			//fakeCurrent := 10 + rand.Float32()*10
			k.broadcast(current)
		}
	}()
}

func (k *KoradPower) stop() {
	if k.ticker != nil {
		k.ticker.Stop()
	}
}
