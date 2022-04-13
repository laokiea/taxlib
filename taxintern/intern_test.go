package taxintern

import (
	"testing"
)

func TestGet(t *testing.T) {
	var a, b = "hello", "hello"
	_a, _b := Get(a), Get(b)
	if _a != _b {
		t.Error("different pointer")
	}
}
