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

type Entry struct {
	sequenceNumber int
	time           time.Time
	server         string
	message        *Message
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
			fmt.Printf("recorded %v\n", entry)
		}
	}()
	return reg
}

func (reg *Registrar) Add(server string, message *Message) {
	reg.recorder <- Entry{0, time.Now(), server, message}
}
