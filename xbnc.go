package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	reg := CreateRegistrar()
	reg.AddNotifier("primary")
	readConfig()
	client := CreateClient(reg, conf.Nick, conf.Login, conf.Ident)
	for _, serverConf := range conf.Servers {
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
	/* lisn, err := CreateListener(reg, client, conf.Port)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	} 

	err = lisn.Listen()
	if err != nil {
		fmt.Println(err)
		return
	}*/

	go func() {
		time.Sleep(5000 * time.Millisecond)
		reg.AddNotifier("later")
	}()
	<-make(chan int)
}
