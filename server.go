package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type IRCServer struct {
	registrar *Registrar

	connected bool
	sock      io.Closer
	read      chan *IRCMessage
	write     chan string

	addr     *net.TCPAddr
	channels map[string]*IRCChannel

	ServerConfig
	serverId  rune
	givenNick string
}

type IRCChannel struct {
	name         string
	active       bool
	creationTime uint64
	mode         string
	members      map[string]string // nick to full name
}

func (channel *IRCChannel) Copy() IRCChannel {
	outmembers := make(map[string]string)
	for k, v := range channel.members {
		outmembers[k] = v
	}
	return IRCChannel{channel.name, channel.active, channel.creationTime, channel.mode, outmembers}
}

func CreateServer(registrar *Registrar, sc ServerConfig) (*IRCServer, error) {
	read := make(chan *IRCMessage, 1000)
	write := make(chan string, 1000)
	channels := make(map[string]*IRCChannel)
	if sc.Port == 0 {
		sc.Port = 6667
	}
	addr, err := net.ResolveTCPAddr("tcp4", sc.Host+":"+strconv.Itoa(sc.Port))
	if err != nil {
		return nil, err
	}
	serverId := ([]rune(sc.Name))[0]
	srv := &IRCServer{registrar, false, nil, read, write, addr, channels, sc, serverId, ""}
	registrar.serversMutex.Lock()
	defer registrar.serversMutex.Unlock()
	_, already := registrar.servers[serverId]
	if already {
		return nil, errors.New("ERROR: already have server starting with " + string(sc.Name[0]) + " when adding " + sc.Name)
	}
	registrar.servers[serverId] = srv
	var reader *bufio.Reader
	var writer *bufio.Writer
	if srv.Ssl {
		config := &tls.Config{InsecureSkipVerify: true}
		conn, err := tls.Dial("tcp", srv.Host+":"+strconv.Itoa(srv.Port), config)
		if err != nil {
			return srv, err
		}
		reader = bufio.NewReader(conn)
		writer = bufio.NewWriter(conn)
		srv.sock = conn
	} else {
		sock, err := net.DialTCP("tcp4", nil, srv.addr)
		if err != nil {
			return srv, err
		}
		reader = bufio.NewReader(sock)
		writer = bufio.NewWriter(sock)
		srv.sock = sock
	}
	srv.connected = true

	go func() {
		for srv.connected {
			str, err := reader.ReadString('\n')
			if err != nil {
				continue
			}

			msg := ParseMessage(str[0 : len(str)-2]) // Cut off the \r\n and parse
			fmt.Printf("reads(%s): %s\n", srv.Name, msg.raw)
			srv.read <- msg
		}
	}()
	go func() {
		for srv.connected {
			str := <-srv.write

			_, err := writer.WriteString(str + "\r\n")
			fmt.Printf("writes(%s): %s\n", srv.Name, str)
			if err != nil {
				fmt.Printf("server write error %v\n", err)
				continue
			}
			writer.Flush()
		}
	}()

	if srv.Password != "" {
		srv.write <- "PASS " + srv.Password
	}
	srv.write <- "NICK " + srv.Nick
	srv.write <- "USER " + srv.Login + " 0 * :XBNC"

	for {
		msg := <-srv.read
		if msg == nil {
			continue
		}

		if msg.command == "PING" {
			srv.write <- "PONG :" + msg.message
		} else if msg.replycode >= 1 && msg.replycode <= 5 {
			if msg.replycode == 1 {
				srv.givenNick = msg.param[0]
				fmt.Printf("set givenNick to %s\n", srv.givenNick)
				srv.record(&Join{channel: "#hello"})
			}
			//srv.client.write <- ":-!xbnc@xbnc PRIVMSG " + srv.client.hostToChannel(srv.Host, "") + " :" + msg.message
			srv.record(&Message{"#hello", msg.message, "server"})
			// Successful connect
			break
		} else if msg.replycode == 433 {
			fmt.Printf("Nick already in use: %s\n", srv.Nick)
			srv.Nick = srv.Nick + "_"
			srv.write <- "NICK " + srv.Nick
		} else if msg.replycode >= 400 && msg.replycode != 439 {
			srv.Close()
			return srv, errors.New("Could not log into IRC server: " + msg.raw)
		}
	}

	go srv.handler()
	return srv, nil
}

func (srv *IRCServer) record(payload Inspecter) {
	srv.registrar.Add(srv.serverId, payload)
}

func (srv *IRCServer) GetChannel(channel string) *IRCChannel {
	tmp, exists := srv.channels[channel]
	if exists {
		tmp.active = true
	} else {
		tmp = &IRCChannel{channel, true, 0, "", make(map[string]string)}
		srv.channels[channel] = tmp
	}
	return tmp
}

func (srv *IRCServer) processServerMessage(msg *IRCMessage) {
	switch msg.command {
	case PING:
		srv.write <- "PONG :" + msg.message
	case REPLY:
		srv.handleReplyCode(msg)
	case JOIN:
		channel := srv.GetChannel(msg.param[0])
		if msg.source == srv.givenNick {
			srv.record(&Join{channel: msg.message})
			srv.write <- "MODE " + msg.message
		} else {
			// another user joined a channel
			fmt.Printf("message source [%s] != server nick [%s]\n", msg.source, srv.givenNick)
			//srv.client.write <- ":" + msg.fullsource + " JOIN :" + srv.client.hostToChannel(srv.Host, msg.message)
			channel.members[msg.message] = msg.fullsource
		}
	case PART:
		if msg.source == srv.Nick {
			_, exists := srv.channels[msg.param[0]]
			if exists {
				delete(srv.channels, msg.param[0])
			}
			//srv.client.partChannel(srv.client.hostToChannel(srv.Host, msg.param[0]))
		} else {
			//srv.client.write <- ":" + msg.fullsource + " PART " + srv.client.hostToChannel(srv.Host, msg.param[0]) + " :" + msg.message
		}
	case KICK:
		if msg.param[1] == srv.Nick {
			channel, exists := srv.channels[msg.param[0]]
			if exists && channel.active {
				channel.active = false
				go func(srv *IRCServer, name string) {
					time.Sleep(3 * time.Second)
					if srv.connected {
						channel, exists := srv.channels[name]
						if exists && !channel.active {
							srv.write <- "JOIN " + channel.name
						}
					}
				}(srv, channel.name)
			}
			//srv.client.kickChannel(srv.client.hostToChannel(srv.Host, msg.param[0]), msg.message)
		} else {
			//srv.client.write <- ":" + msg.fullsource + " KICK " + srv.client.hostToChannel(srv.Host, msg.param[0]) + " " + msg.param[1] + " :" + msg.message
		}
	case QUIT:
		//srv.client.write <- ":" + msg.fullsource + " QUIT :" + msg.message
	case PRIVMSG:
		name := msg.param[0]
		if name == srv.givenNick {
			name = msg.source
		}
		//channel := srv.client.hostToChannel(srv.Host, name)
		//srv.client.joinChannel(channel, false)
		//srv.client.write <- ":" + msg.fullsource + " PRIVMSG " + channel + " :" + msg.message
		srv.record(&Message{name, msg.message, msg.fullsource})
	case NOTICE:
		//srv.client.write <- msg.raw
		if len(srv.Ident) > 0 && msg.source == "NickServ" && strings.HasPrefix(msg.message, "This nickname is registered and protected") {
			srv.write <- "NICKSERV IDENTIFY " + srv.Ident
		}
	case MODE:
		if msg.paramlen == 4 {
			//srv.client.write <- ":" + msg.fullsource + " MODE " + srv.client.hostToChannel(srv.Host, msg.param[0]) + " " + msg.param[1] + " " + msg.param[2] + " " + msg.param[3]
		} else if msg.paramlen == 3 {
			//srv.client.write <- ":" + msg.fullsource + " MODE " + srv.client.hostToChannel(srv.Host, msg.param[0]) + " " + msg.param[1] + " " + msg.param[2]
		} else if msg.paramlen == 2 {
			//srv.client.write <- ":" + msg.fullsource + " MODE " + srv.client.hostToChannel(srv.Host, msg.param[0]) + " " + msg.param[1]
		} else if msg.paramlen == 1 {
			//srv.client.write <- ":" + msg.fullsource + " MODE " + msg.param[0] + " :" + msg.message
		} else {
			//srv.client.write <- ":-!xbnc@xbnc PRIVMSG " + srv.client.hostToChannel(srv.Host, "") + " :" + msg.raw
			srv.record(&Message{srv.Host, msg.raw, "mode"})
		}
	case NICK:
		//srv.client.write <- msg.raw
	case TOPIC:
		//srv.client.write <- ":" + msg.fullsource + " TOPIC " + srv.client.hostToChannel(srv.Host, msg.param[0]) + " :" + msg.message
		srv.record(&TopicSet{msg.param[0], msg.message, msg.fullsource})
	case CTCP_VERSION:
		//srv.client.write <- ":" + msg.source + "!xbnc@xbnc PRIVMSG " + srv.client.hostToChannel(srv.Host, "") + " :Received CTCP VERSION: " + msg.raw
		srv.write <- "NOTICE " + msg.source + " :\x01XBNC 1.0: Created By xthexder\x01"
		srv.record(&Message{"ctcp", msg.raw, "ctcp"})
	default:
		srv.record(&Message{srv.Host, msg.raw, "server"})
		//srv.client.write <- ":-!xbnc@xbnc PRIVMSG " + srv.client.hostToChannel(srv.Host, "") + " :" + msg.raw
	}

}
func (srv *IRCServer) handler() {
	for srv.connected {
		msg := <-srv.read
		if msg == nil {
			continue
		}
		srv.processServerMessage(msg)

	}
}

func (srv *IRCServer) handleReplyCode(msg *IRCMessage) {
	//replycode := fmt.Sprintf("%03d", msg.replycode)
	if msg.replycode >= 1 && msg.replycode <= 3 {
		//srv.client.write <- ":-!xbnc@xbnc PRIVMSG " + srv.client.hostToChannel(srv.Host, "") + " :" + msg.message
		srv.record(&Message{"#hello", msg.message, srv.Name})
	} else if (msg.replycode >= 4 && msg.replycode <= 5) || (msg.replycode >= 251 && msg.replycode <= 255) { // Server info
		tmpi := strings.Index(msg.raw, msg.param[0])
		if tmpi >= 0 && len(msg.param[0]) > 0 {
			tmpmsg := strings.TrimSpace(msg.raw[tmpi+len(msg.param[0]):])
			if strings.HasPrefix(tmpmsg, ":") {
				tmpmsg = tmpmsg[1:]
			}
			//srv.client.write <- ":-!xbnc@xbnc PRIVMSG " + srv.client.hostToChannel(srv.Host, "") + " :" + tmpmsg
			srv.record(&Message{"#hello", tmpmsg, srv.Name})
		} else {
			//srv.client.write <- ":-!xbnc@xbnc PRIVMSG " + srv.client.hostToChannel(srv.Host, "") + " :" + msg.raw
		}
	} else if (msg.replycode >= 265 && msg.replycode <= 266) || msg.replycode == 375 || msg.replycode == 372 || msg.replycode == 376 { // Server info and MOTD
		//srv.client.write <- ":-!xbnc@xbnc PRIVMSG " + srv.client.hostToChannel(srv.Host, "") + " :" + msg.message
		srv.record(&Message{"#hello", msg.message, srv.Name})
	} else if msg.replycode == RPL_TOPIC { // Channel topic
		//srv.client.write <- ":" + conf.Hostname + " " + replycode + " " + msg.param[0] + " " + srv.client.hostToChannel(srv.Host, msg.param[1]) + " :" + msg.message
		srv.record(&TopicSet{msg.param[1], msg.message, msg.param[0]})
	} else if msg.replycode == 333 { // Channel topic setter
		//srv.client.write <- ":" + conf.Hostname + " " + replycode + " " + msg.param[0] + " " + srv.client.hostToChannel(srv.Host, msg.param[1]) + " " + msg.param[2] + " " + msg.param[3]
	} else if msg.replycode == RPL_NAMREPLY { // Channel members
		//srv.client.write <- ":" + conf.Hostname + " " + replycode + " " + msg.param[0] + " " + msg.param[1] + " " + srv.client.hostToChannel(srv.Host, msg.param[2]) + " :" + msg.message
		channel := srv.GetChannel(msg.param[2])
		for _, name := range strings.Fields(msg.message) {
			channel.members[name] = ""
		}
		srv.record(&ChannelMembers{channel: msg.param[2], members: strings.Fields(msg.message)})
	} else if msg.replycode == RPL_ENDOFNAMES {
		//srv.client.write <- ":" + conf.Hostname + " " + replycode + " " + msg.param[0] + " " + srv.client.hostToChannel(srv.Host, msg.param[1]) + " :" + msg.message
		srv.record(&EndOfNames{channel: msg.param[1]})
	} else if msg.replycode == RPL_ENDOFWHO {
		//srv.client.write <- ":" + conf.Hostname + " " + replycode + " " + msg.param[0] + " " + srv.client.hostToChannel(srv.Host, msg.param[1]) + " :" + msg.message
	} else if msg.replycode == RPL_CHANNELMODEIS { // Channel mode
		//srv.client.write <- ":" + conf.Hostname + " " + replycode + " " + msg.param[0] + " " + srv.client.hostToChannel(srv.Host, msg.param[1]) + " " + msg.param[2]
		channel := srv.GetChannel(msg.param[1])
		channel.mode = msg.param[2]
		srv.record(&ChannelMode{channel: msg.param[1], mode: msg.param[2]})
	} else if msg.replycode == RPL_CREATIONTIME { // Channel mode
		//srv.client.write <- ":" + conf.Hostname + " " + replycode + " " + msg.param[0] + " " + srv.client.hostToChannel(srv.Host, msg.param[1]) + " " + msg.param[2]
		channel := srv.GetChannel(msg.param[1])
		ct, err := strconv.ParseUint(msg.param[2], 10, 64)
		if err != nil {
			fmt.Printf("unable to parse creation time [%s]: %s\n", msg, msg.param[2], ct)
		} else {
			channel.creationTime = ct
		}
		srv.record(&CreationTime{channel: msg.param[1], time: msg.param[2]})
	} else if msg.replycode == 352 { // Channel who reply
		//srv.client.write <- ":" + conf.Hostname + " " + replycode + " " + msg.param[0] + " " + srv.client.hostToChannel(srv.Host, msg.param[1]) + " " + msg.param[2] + " " + msg.param[3] + " " + conf.Hostname + " " + msg.param[5] + " " + msg.param[6] + " :" + msg.message
	} else {
		//srv.client.write <- ":-!xbnc@xbnc PRIVMSG " + srv.client.hostToChannel(srv.Host, "") + " :" + msg.raw
	}
}

func (srv *IRCServer) Close() {
	srv.connected = false
	close(srv.read)
	close(srv.write)
	srv.channels = make(map[string]*IRCChannel)
	if srv.sock != nil {
		srv.sock.Close()
		srv.sock = nil
	}
}
