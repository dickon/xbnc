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
		go func(sc *ServerConfig) {
			server := client.addServer(*sc)
			if server == nil {
				fmt.Printf("failed to connect to %v", sc)
				os.Exit(5)
			}
			for _, channel := range sc.Channels {
				fmt.Printf("Joining channel %s of %s\n", channel, sc.Name)
				server.write <- "JOIN " + channel
			}
		}(serverConf)
	}
	lisn, err := CreateListener(reg, client, conf.Port)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	err = lisn.Listen()
	if err != nil {
		fmt.Println(err)
		return
	}
	<-make(chan int)
}
