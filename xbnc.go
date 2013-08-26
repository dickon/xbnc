package main

import (
	"fmt"
	"os"
)

func main() {
	reg := CreateRegistrar()
	readConfig()
	client := CreateClient(reg, conf.Nick, conf.Login, conf.Ident)
	for _, serverConf := range conf.Servers {
		// TODO: why not just pass serverConf to addServer?
		server := client.addServer(*serverConf)
		if server == nil {
			fmt.Printf("failed to connect to %v", serverConf)
			os.Exit(5)
		}
		for _, channel := range serverConf.Channels {
			fmt.Printf("Joining channel %s of %s\n", channel, server.serverConfig.Name)
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
