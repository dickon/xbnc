package main
import "testing"

func TestTokenizeString(t *testing.T) {
	foo, bar := tokenizeString("hello")
	if bar != "" {
		t.Errorf("tokenizeString foo")
	}
	if len(foo) != 1 || foo[0] != "hello" {
		t.Errorf("bad out");
	}
}
