package main

import (
	"fmt"
	"time"
)

type Registrar struct {
	entries  []Entry
	recorder chan Entry
}

type Message struct {
	channel string
	text    string
	author  string
}

func (message Message) Render() string {
	return message.channel + ":" + message.author + ":" + message.text
}

type Inspecter interface {
	Render() string
}

type Entry struct {
	sequenceNumber int
	time           time.Time
	server         string
	payload        Inspecter
}

func CreateRegistrar() *Registrar {
	entries := make([]Entry, 0, 100)
	recorder := make(chan Entry, 100)
	reg := &Registrar{entries, recorder}
	go func() {
		for {
			entry := <-reg.recorder
			entry.sequenceNumber = len(reg.entries)
			reg.entries = append(reg.entries, entry)
			fmt.Printf("recorded %d:%s\n", entry.sequenceNumber, entry.payload.Render())
		}
	}()
	return reg
}

func (reg *Registrar) Add(server string, message *Message) {
	reg.recorder <- Entry{0, time.Now(), server, *message}
}
