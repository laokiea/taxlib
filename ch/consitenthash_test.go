package ch

import (
	"testing"
)

func FuzzJoin(f *testing.F) {
	f.Add("182.168.17.4", "server1")
	f.Fuzz(func(t *testing.T, input string, input2 string) {
		c := ConsistentHash{}
		hfo := HashFuncOpt{DefaultHashFunc}
		c.New(&hfo)
		err := c.Join(input, input2)
		if err != nil {
			t.Fatalf("ParseQuery failed to decode a valid encoded query %s: %v", "test", err)
		}
	})
}
