package main

import (
	. "github.com/robertkrimen/terst"
	"testing"
)

func TestTokenizeString(t *testing.T) {
	Terst(t)
	tokens1, result1 := tokenizeString("hello")

	Is(result1, "")
	Is(len(tokens1), 1)
	Is(tokens1[0], "hello")

	tokens2, result2 := tokenizeString("hello:world")
	Is(result2, "world")
	Is(tokens2, []string{"hello"})
}

func TestParseMessage(t *testing.T) {
	Terst(t)

	m := ParseMessage("PRIVMSG Vultan :Gordon's alive")
	Compare(m.time, ">=", 1e9)
	Is(m.command, "PRIVMSG")

	m = ParseMessage(":irc.example.com 001 MyNickname :Welcome to the Internet Relay Network MyNickname!~MyUsername@client.example.com")
	Compare(m.time, ">=", 1e9)
	Is(m.command, "REPLY")

}
