package main
import (
	"testing"
	. "github.com/robertkrimen/terst"
)

func TestTokenizeString(t *testing.T) {
	Terst(t)
	tokens1, result1 := tokenizeString("hello")
	
	Is(result1, "")
	Is(len(tokens1), 1)
	Is(tokens1[0], "hello")

	
}
