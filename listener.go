package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
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

type ClientConnection struct {
	conn      *net.TCPConn
	reader    *bufio.Reader
	writer    *bufio.Writer
	regnotify chan Entry
	registrar *Registrar
}

func (cc ClientConnection) Start() {
	cc.registrar.Subscribe(cc.regnotify)

	go func() {
		for {
			entry := <-cc.regnotify
			str := entry.payload.Command(&entry)
			n, err := cc.writer.WriteString(str)
			if err != nil {
				fmt.Printf("writestring via registrar %d error: %v\n", n, err)
			}
			fmt.Printf("writec via registrar: %s\n", str)
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

			msg := ParseMessage(str[0 : len(str)-2])
			fmt.Printf("got message %v\n", msg)
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
			(ClientConnection{conn, bufio.NewReader(conn), bufio.NewWriter(conn), make(chan Entry, 100), lisn.registrar}).Start()
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
