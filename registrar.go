package main

import (
	"fmt"
	"time"
)

type Registrar struct {
	messages []Message
	recorder chan Message
}

type Message struct {
	sequenceNumber int
	channel        string
	text           string
	time           time.Time
}

func CreateRegistrar() *Registrar {
	messages := make([]Message, 0, 100)
	recorder := make(chan Message, 100)
	reg := &Registrar{messages, recorder}
	go func() {
		for {
			mesrec := <-reg.recorder
			mesrec.sequenceNumber = len(reg.messages)
			reg.messages = append(reg.messages, mesrec)
			fmt.Printf("recorded %v\n", mesrec)
		}

	}()
	return reg
}

func (reg *Registrar) Add(message, channel string) {
	mesrec := Message{0, message, channel, time.Now()}
	reg.recorder <- mesrec
}
