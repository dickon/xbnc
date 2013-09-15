package main

import (
	"fmt"
	"os"
)

func main() {
	reg := CreateRegistrar()
	readConfig()
	for _, serverConf := range conf.Servers {
		go func(sc *ServerConfig) {
			server, err := CreateServer(reg, *sc)
			if err != nil {
				fmt.Printf("failed to connect to %v: error %v", sc, err)
				os.Exit(5)
			}
			for _, channel := range sc.Channels {
				fmt.Printf("Joining channel %s of %s\n", channel, sc.Name)
				server.write <- "JOIN " + channel
			}
		}(serverConf)
	}
	lisn, err := CreateListener(reg, conf.Port)
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
