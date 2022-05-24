package main

import (
	"errors"
	"time"

	breaker "./breaker"
)

func main() {
	// var l lock.Lock
	// l.Lock()
	// println(l.TryLock())
	// l.Unlock()
	// println(l.TryLock())
	// l.Lock()
	var i int
	breaker := breaker.NewBreaker()
	for {
		_, _ = breaker.Do(func() (interface{}, error) {
			if i == 0 {
				return "ok", nil
			} else if i > 11 {
				return "ok", nil
			} else {
				return nil, errors.New("e")
			}
		})
		println(i, ":", breaker.State())
		i++
		time.Sleep(1e9)
	}
}
