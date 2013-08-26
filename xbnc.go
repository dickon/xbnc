package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var conf Config

type Config struct {
	Hostname string
	Nick     string
	Login    string
	Ident    string
	Port     int
}

func main() {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("Config file error: %v\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(file, &conf)
	if err != nil {
		fmt.Printf("Config file error: %v\n", err)
		os.Exit(2)
	}
	if conf.Hostname == "" {
		if conf.Hostname, err = os.Hostname(); err != nil {
			fmt.Printf("Unable to determinte hostname: %v\n", err)
			os.Exit(3)
		}
	}
	if conf.Port == 0 {
		conf.Port = 1234
	}

	client := CreateClient(conf.Nick, conf.Login, conf.Ident)

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
