package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type IRCListener struct {
	addr      *net.TCPAddr
	registrar *Registrar
}

func CreateListener(registrar *Registrar, port int) (*IRCListener, error) {
	addr, err := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return &IRCListener{addr, registrar}, nil
}

type ClientOut struct {
	message, why string
}

type ServerInfo struct {
	channels map[string]IRCChannel
}

type ClientConnection struct {
	conn       *net.TCPConn
	reader     *bufio.Reader
	writer     *bufio.Writer
	regnotify  chan Notification
	registrar  *Registrar
	login      string
	nick       string
	address    string
	output     chan ClientOut
	registered bool // set to true when registration is complete
}

func getServer(cchannel string, reg *Registrar) (string, *IRCServer) {
	name := []rune(cchannel)
	server, exists := reg.servers[name[1]]
	sname := string(name[0]) + string(name[2:])
	if !exists {
		fmt.Printf("unknown server %s\n", name[1])
		return sname, nil
	}
	return sname, server
}

func (cc ClientConnection) Start() {

	go func() { // send registrar messages down to the output queue
		for {
			notification := <-cc.regnotify
			fmt.Printf("handling %b %s\n", notification.fresh, notification.Render())
			switch t := notification.payload.(type) {
			case *Message:
				server, exists := cc.registrar.servers[notification.server]
				// TODO: check that this came from this client
				if exists && server.givenNick == t.author {
					fmt.Printf("ignoring message from me\n")
					continue
				}
			}
			str := notification.payload.Command(notification.server, &cc)
			cc.output <- ClientOut{str, "via registrar"}
		}
	}()

	go func() { // dispatch the output queue
		for {
			cmesg := <-cc.output
			n, err := cc.writer.WriteString(cmesg.message + "\r\n")
			if err != nil {
				fmt.Printf("writestring %s %d error: %v\n", cmesg.why, n, err)
			}
			fmt.Printf("writec %s: %s\n", cmesg.why, cmesg.message)
			cc.writer.Flush()
		}
	}()

	go func() { // client message dispatch
		for {
			str, err := cc.reader.ReadString('\n')
			if err != nil {
				fmt.Printf("readc error: %v\n", err)
				break
			}
			fmt.Printf("readc %#v\n", str)
			cc.handleClientMessage(str)
		}

	}()
}

func (cc *ClientConnection) handleClientMessage(str string) {

	msg := ParseMessage(str)
	fmt.Printf("got client command %#v message %#v\n", msg.command, msg.message)

	switch msg.command {
	case PING:
		cc.output <- ClientOut{":" + conf.Hostname + " " + PONG + conf.Hostname + " :" + msg.param[0], "client ping"}
	case NICK:
		cc.nick = msg.param[0]
	case USER:
		cc.login = msg.param[0]
	case MODE:
		sname, server := getServer(msg.param[0], cc.registrar)
		if server != nil {
			server.write <- MODE + " " + sname
		}
	case PRIVMSG:
		sname, server := getServer(msg.param[0], cc.registrar)
		if server != nil {
			server.write <- PRIVMSG + " " + sname + " :" + msg.message
			server.record(&Message{sname, msg.message, server.givenNick})
		}
	default:
		fmt.Printf("Unhandled client command %s", msg.command)
	}
	if !cc.registered && cc.nick != "" && cc.login != "" {
		cc.Send(RPL_WELCOME, ":Welcome to XBNC "+cc.nick+"!"+cc.login+"@"+cc.address, "logged in")
		cc.Send(RPL_YOURHOST, ":Your host is "+conf.Hostname, "logged in")
		cc.Send(RPL_CREATED, ":This server was created today", "logged in") // TODO correct date
		cc.Send(RPL_MYINFO, ":"+conf.Hostname+" XBNC2.0 iowghraAsORTVSxNCWqBzvdHtGpI lvhopsmntikrRcaqOALQbSeIKVfMCuzNTGjZ", "logged in")
		cc.Send(RPL_BOUNCE, ":CHANTYPES=# NETWORK=XBNC PREFIX=(qaohv)~&@%+ CASEMAPPING=ascii :are supported by this serVer", "logged in")
		cc.registered = true
		cc.registrar.Subscribe(cc.regnotify)
	}

}

func (cc *ClientConnection) Send(code int, payload string, why string) {
	cc.output <- ClientOut{fmt.Sprintf("%03d %s %s", code, cc.nick, payload), "logged in"}
}

func (lisn *IRCListener) Listen() error {

	listener, err := net.ListenTCP("tcp4", lisn.addr)
	if err != nil {
		return err
	}

	fmt.Printf("Listening for TCP on %d\n", lisn.addr.Port)

	go func() {
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Printf("Listen error: %v\n", err)
				continue
			}
			remaddr := conn.RemoteAddr()
			segments := strings.Split(remaddr.String(), ":")
			fmt.Printf("Accepted incoming connection on %s\n", segments[0])
			(ClientConnection{conn, bufio.NewReader(conn), bufio.NewWriter(conn), make(chan Notification, 100), lisn.registrar, "", "", segments[0], make(chan ClientOut, 100), false}).Start()
		}
	}()
	return nil
}
