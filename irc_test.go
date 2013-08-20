package main
import "testing"

Func TestTokenizeString(t *testing.T) {
	foo, bar := tokenizeString("hello")
	if bar != "" {
		T.Errorf("tokenizeString foo")
	}
	if len(foo) != 1 || foo[0] != "hello" {
		t.Errorf("bad out");
	}
}
