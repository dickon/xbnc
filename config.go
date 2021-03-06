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
	Port     int
	Servers  []*ServerConfig
	Nick     string // client facing
	Login    string // client facing
	Ident    string // client facing
}

type ServerConfig struct {
	Name     string
	Host     string
	Port     int
	Ssl      bool
	Password string
	Nick     string
	Login    string
	Ident    string
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
	if conf.Login == "" {
		u, err := user.Current()
		if err != nil {
			fmt.Printf("Unable to determinte user name: %v\n", err)
			os.Exit(4)
		}
		conf.Login = u.Username
	}
	if conf.Port == 0 {
		conf.Port = 1234
	}
	seen := make(map[string]bool)
	for i, elem := range conf.Servers {
		if elem.Name == "" {
			elem.Name = elem.Host
		}
		if elem.Login == "" {
			u, err := user.Current()
			if err != nil {
				fmt.Printf("Unable to determinte user name: %v\n", err)
				os.Exit(4)
			}
			elem.Login = u.Username
		}
		_, already := seen[elem.Name]
		if already {
			fmt.Printf("Server name %s reused\n", elem.Name)
			os.Exit(5)
		}
		seen[elem.Name] = true
		fmt.Printf("Server %d: %v\n", i+1, elem)
		if elem.Host == "" {
			fmt.Printf("No host specified on server %d\n", i+1)
			os.Exit(4)
		}
	}
}
