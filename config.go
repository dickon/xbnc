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
			fmt.Printf("Unable to determinte hostname: %v\n", err)
			os.Exit(3)
		}
	}
	if conf.Port == 0 {
		conf.Port = 1234
	}

}
