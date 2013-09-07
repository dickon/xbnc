package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

const (
	RPL_WELCOME  = 1
	RPL_YOURHOST = 2
	RPL_CREATED  = 3
	RPL_MYINFO   = 4
	RPL_BOUNCE   = 5
)

type IRCListener struct {
	addr *net.TCPAddr

	client    *IRCClient
	registrar *Registrar
}

func CreateListener(registrar *Registrar, client *IRCClient, port int) (*IRCListener, error) {
	addr, err := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return &IRCListener{addr, client, registrar}, nil
}

type ClientOut struct {
	message, why string
}

type ClientConnection struct {
	conn       *net.TCPConn
	reader     *bufio.Reader
	writer     *bufio.Writer
	regnotify  chan Entry
	registrar  *Registrar
	login      string
	nick       string
	address    string
	output     chan ClientOut
	registered bool // set to true when registration is complete
}

func (cc ClientConnection) Start() {

	go func() {
		for {
			entry := <-cc.regnotify
			str := entry.payload.Command(&entry)
			cc.output <- ClientOut{str, "via registrar"}
		}
	}()

	go func() {
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

	go func() {
		for {
			str, err := cc.reader.ReadString('\n')
			if err != nil {
				fmt.Printf("readc error: %v\n", err)
				break
			}
			fmt.Printf("readc [%s]\n", str)

			msg := ParseMessage(str[0 : len(str)-2])
			fmt.Printf("got client command %s\n", msg.command)

			switch msg.command {
			case "NICK":
				cc.nick = msg.param[0]
			case "USER":
				cc.login = msg.param[0]
			}
			if !cc.registered && cc.nick != "" && cc.login != "" {
				cc.output <- ClientOut{fmt.Sprintf("%03d %s :Welcome to XBNC %s!%s@%s", RPL_WELCOME, cc.nick, cc.nick, cc.login, cc.address), "logged in"}
				cc.output <- ClientOut{fmt.Sprintf("%03d %s :Your host is %s", RPL_YOURHOST, cc.nick, conf.Hostname), "logged in"}
				cc.output <- ClientOut{fmt.Sprintf("%03d %s :This server was created today", RPL_CREATED, cc.nick), "logged in"} // TODO: give proper date
				cc.output <- ClientOut{fmt.Sprintf("%03d %s :%s XBNC2.0 iowghraAsORTVSxNCWqBzvdHtGpI lvhopsmntikrRcaqOALQbSeIKVfMCuzNTGjZ", RPL_MYINFO, cc.nick, conf.Hostname), "logged in"}
				cc.output <- ClientOut{fmt.Sprintf("%03d %s ::CHANTYPES=# NETWORK=XBNC PREFIX=(qaohv)~&@%+ CASEMAPPING=ascii :are supported by this serVer", RPL_BOUNCE, cc.nick), "logged in"}
				cc.registrar.Subscribe(cc.regnotify)
			}
		}

	}()
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
			fmt.Printf("Accepted incoming connection on %s:%s\n", remaddr.Network(), remaddr.String())
			(ClientConnection{conn, bufio.NewReader(conn), bufio.NewWriter(conn), make(chan Entry, 100), lisn.registrar, "", "", remaddr.String(), make(chan ClientOut, 100), false}).Start()
		}
	}()
	return nil
}

/*func (lisn *IRCListener) Close() {
  lisn.listening = false
  if lisn.client != nil {
    lisn.client.Close()
    lisn.client = nil
  }
  if lisn.listener != nil {
    lisn.listener.Close()
    lisn.listener = nil
  }
}*/
