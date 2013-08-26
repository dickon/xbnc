package main

import (
	"fmt"
	"os"
)

func main() {
	readConfig()
	client := CreateClient(conf.Nick, conf.Login, conf.Ident)
	for _, serverConf := range conf.Servers {
		server := client.addServer(serverConf.Host, serverConf.Port, serverConf.Password, serverConf.Ssl)
		for _, channel := range serverConf.Channels {
			server.write <- "JOIN " + channel
		}
	}
	lisn, err := CreateListener(client, conf.Port)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	/*reader := bufio.NewReader(os.Stdin)
	  for {
	    str, err := reader.ReadString('\n')
	    if err != nil {
	      fmt.Printf("stdin: %s\n", err)
	    }
	    srv.write <- "PRIVMSG #bnctest :" + str[0:len(str)-1]
	  }*/

	err = lisn.Listen()
	if err != nil {
		fmt.Println(err)
		return
	}
	<-make(chan int)
}
