package main

import (
	"fmt"
	"time"
)

type Message struct {
	channel string
	text    string
	author  string
}

func (message Message) Render() string {
	return fmt.Sprintf("%s on %s said '%s'", message.author, message.channel, message.text)
}

func (message Message) Command(entry *Entry) string {
	return fmt.Sprintf("PRIVMSG #%c%s :%s(%d)", entry.server, message.channel, message.text, entry.sequenceNumber)
}

type OtherJoin struct {
	channel string
	author  string
}

func (join OtherJoin) Render() string {
	return join.channel + " joined by " + join.author
}

func (join OtherJoin) Command(entry *Entry) string {
	return fmt.Sprintf(":%s 353 %c%s %s", conf.Hostname, entry.server, join.channel, join.author)
}

type MyJoin struct {
	channel string
}

func (join MyJoin) Render() string {
	return "joined " + join.channel
}

func (join MyJoin) Command(entry *Entry) string {
	return fmt.Sprintf("JOIN :#%c%s", entry.server, join.channel[1:])
}

type TopicSet struct {
	channel string
	text    string
	author  string
}

func (topic TopicSet) Render() string {
	return topic.channel + " topic set to " + topic.text + " by " + topic.author
}

func (topic TopicSet) Command(entry *Entry) string {
	return ""
}

type Inspecter interface {
	Render() string
	Command(entry *Entry) string
}

type Entry struct {
	sequenceNumber int
	time           time.Time
	server         rune
	payload        Inspecter
}

type Registrar struct {
	entries   []Entry
	notifiers []chan Entry
	recorder  chan Entry
}

func CreateRegistrar() *Registrar {
	entries := make([]Entry, 0, 100)
	notifiers := make([]chan Entry, 0, 100)
	recorder := make(chan Entry, 100)
	reg := &Registrar{entries, notifiers, recorder}
	go func() {
		for {
			entry := <-reg.recorder
			entry.sequenceNumber = len(reg.entries)
			reg.entries = append(reg.entries, entry)
			for _, notifier := range reg.notifiers {
				notifier <- entry
			}
		}
	}()

	return reg
}

func (reg *Registrar) Add(server rune, payload Inspecter) {
	reg.recorder <- Entry{0, time.Now(), server, payload}
}

func (reg *Registrar) Subscribe(notifier chan Entry) {
	reg.notifiers = append(reg.notifiers, notifier)
	for _, entry := range reg.entries {
		notifier <- entry
	}
}

func (reg *Registrar) AddNotifier(prefix string) {
	echonotify := make(chan Entry, 100)
	go func() {
		for {
			entry := <-echonotify
			fmt.Printf("%s recorded %d:%s\n", prefix, entry.sequenceNumber, entry.payload.Render())
		}
	}()
	reg.Subscribe(echonotify)
}
