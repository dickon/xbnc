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

func (message Message) Command(entry *Entry, cc *ClientConnection) string {
	channel := []rune(message.channel)
	return fmt.Sprintf(":%s PRIVMSG %c%c%s :%s (%d)", message.author, channel[0], entry.server, string(channel[1:]), message.text, entry.sequenceNumber)
}

func (channel IRCChannel) Render() string {
	return "joined " + channel.name
}

func (channel IRCChannel) Command(entry *Entry, cc *ClientConnection) string {
	name := []rune(channel.name)
	return fmt.Sprintf(":%s!%s@%s JOIN :%c%c%s", cc.nick, cc.login, cc.address, name[0], entry.server, string(name[1:]))
}

type TopicSet struct {
	channel string
	text    string
	author  string
}

func (topic TopicSet) Render() string {
	return topic.channel + " topic set to " + topic.text + " by " + topic.author
}

func (topic TopicSet) Command(entry *Entry, cc *ClientConnection) string {
	return ""
}

type Inspecter interface {
	Render() string
	Command(entry *Entry, cc *ClientConnection) string
}

type Entry struct {
	sequenceNumber int
	time           time.Time
	server         rune
	payload        Inspecter
}

func (entry *Entry) Render() string {
	return fmt.Sprintf("entry %05d server %c: %s", entry.sequenceNumber, entry.server, entry.payload.Render())
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
			fmt.Printf("recorded %s\n", entry.Render())
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
