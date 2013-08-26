package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
)

type Config struct {
	Hostname string
	Nick     string
	Login    string
	Ident    string
	Port     int
	Servers  []ServerConfig
}

type ServerConfig struct {
	Host     string
	Port     int
	Channels []string
}

var conf Config

func readConfig() {
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
			fmt.Printf("Unable to determine hostname: %v\n", err)
			os.Exit(3)
		}
	}
	if conf.Port == 0 {
		conf.Port = 1234
	}
	if conf.Login == "" {
		u, err := user.Current()
		if err != nil {
			fmt.Printf("Unable to determinte user name: %v\n", err)
			os.Exit(4)
		}
		conf.Login = u.Username
	}
	for i, elem := range conf.Servers {
		fmt.Printf("Server %d: %s:%d\n", i+1, elem.Host, elem.Port)
		if elem.Host == "" {
			fmt.Printf("No host specified on server %d\n", i+1)
			os.Exit(4)
		}
	}
}
