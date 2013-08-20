package main
import (
	"testing"
	. "github.com/robertkrimen/terst"
)

func TestTokenizeString(t *testing.T) {
	Terst(t)
	foo, bar := tokenizeString("hello")
	
	Is(bar, "")
	Is(len(foo), 1)
	Is(foo[0], "hello")
}
