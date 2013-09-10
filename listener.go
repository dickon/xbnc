package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
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

type ServerInfo struct {
	channels map[string]IRCChannel
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
	servers    map[rune]ServerInfo
}

func (cc ClientConnection) Start() {

	go func() {
		for {
			entry := <-cc.regnotify
			_, exists := cc.servers[entry.server]
			if !exists {
				cc.servers[entry.server] = ServerInfo{make(map[string]IRCChannel)}
			}
			si, _ := cc.servers[entry.server]
			fmt.Printf("handling %s\n", entry.Render())
			str := entry.payload.Command(&entry, &cc)
			switch t := entry.payload.(type) {
			case IRCChannel:
				si.channels[t.name] = t
			}
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
			case "MODE":
				name := []rune(msg.param[0])
				si, exists := cc.servers[name[1]]
				if !exists {
					fmt.Printf("unknown server %s\n", name[1])
				} else {
					// TODO: actually pass the MODE request to the server
					sname := string(name[0]) + string(name[2:])
					channel, cexists := si.channels[sname]
					fmt.Printf("server %s channel %s exists %v\n", name[1], sname, cexists)
					if cexists {
						members := make([]string, len(channel.members))
						i := 0
						for k, _ := range channel.members {
							members[i] = k
							i++
						}
						// TODO replace = with correct char, depending on join mode
						cc.Send(RPL_NAMREPLY, " = "+msg.param[0]+" :"+strings.Join(members, " "), "MODE response")
						cc.Send(RPL_ENDOFNAMES, msg.param[0]+" :End of /NAMES list.", "MODE response")
						cc.Send(RPL_CHANNELMODEIS, msg.param[0]+" "+channel.mode, "MODE response")
						cc.Send(RPL_CREATIONTIME, fmt.Sprintf("%s %d", msg.param[0], channel.creationTime), "MODE response")
					}
				}

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

	}()
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
			(ClientConnection{conn, bufio.NewReader(conn), bufio.NewWriter(conn), make(chan Entry, 100), lisn.registrar, "", "", segments[0], make(chan ClientOut, 100), false, make(map[rune]ServerInfo)}).Start()
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
