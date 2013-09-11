package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Message struct {
	channel string
	text    string
	author  string
}

func clientChannelName(serverID rune, serverChannel string) string {
	name := []rune(serverChannel)
	return string(name[0]) + string(serverID) + string(name[1:])
}

func (message Message) Command(entry *Entry, cc *ClientConnection) string {
	return fmt.Sprintf(":%s PRIVMSG %s :%s (%d)", message.author, clientChannelName(entry.server, message.channel), message.text, entry.sequenceNumber)
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

func (topic TopicSet) Command(entry *Entry, cc *ClientConnection) string {
	return ""
}

type Inspecter interface {
	Command(entry *Entry, cc *ClientConnection) string
}

type Entry struct {
	sequenceNumber int
	time           time.Time
	server         rune
	payload        Inspecter
}

func (entry *Entry) Render() string {
	return fmt.Sprintf("entry %05d server %c: %#v", entry.sequenceNumber, entry.server, entry.payload)

}

type ChannelMembers struct {
	channel string
	members []string
}

func (cm *ChannelMembers) Command(entry *Entry, cc *ClientConnection) string {
	return fmt.Sprintf(":%s %03d %s @ %s :%s", cc.address, RPL_NAMREPLY, cc.nick, clientChannelName(entry.server, cm.channel), strings.Join(cm.members, " "))
}

type EndOfNames struct {
	channel string
}

func (eon *EndOfNames) Command(entry *Entry, cc *ClientConnection) string {
	return fmt.Sprintf(":%s %03d %s %s :End of /NAMES list.", cc.address, RPL_ENDOFNAMES, cc.nick, clientChannelName(entry.server, eon.channel))
}

type ChannelMode struct {
	channel string
	mode    string
}

func (cm *ChannelMode) Command(entry *Entry, cc *ClientConnection) string {
	return fmt.Sprintf(":%s %03d %s %s %s", cc.address, RPL_CHANNELMODEIS, cc.nick, clientChannelName(entry.server, cm.channel), cm.mode)
}

type CreationTime struct {
	channel string
	time    string
}

func (ct *CreationTime) Command(entry *Entry, cc *ClientConnection) string {
	return fmt.Sprintf(":%s %03d %s %s %s", cc.address, RPL_CREATIONTIME, cc.nick, clientChannelName(entry.server, ct.channel), ct.time)
}

type Registrar struct {
	entries      []Entry
	notifiers    []chan Entry
	recorder     chan Entry
	servers      map[rune]*IRCServer
	serversMutex sync.Mutex
}

func CreateRegistrar() *Registrar {
	entries := make([]Entry, 0, 100)
	notifiers := make([]chan Entry, 0, 100)
	recorder := make(chan Entry, 100)
	servers := make(map[rune]*IRCServer)
	reg := &Registrar{entries: entries, notifiers: notifiers, recorder: recorder, servers: servers}
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
			fmt.Printf("%s recorded %d:%#v\n", prefix, entry.sequenceNumber, entry.payload)
		}
	}()
	reg.Subscribe(echonotify)
}
