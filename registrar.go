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
	server         string
	text           string
	author         string
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

func (reg *Registrar) Add(message, channel, server, author string) {
	mesrec := Message{0, message, channel, server, author, time.Now()}
	reg.recorder <- mesrec
}
