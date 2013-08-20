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
