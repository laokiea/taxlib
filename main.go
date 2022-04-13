package main

import lock "./lock"

func main()  {
	var l lock.Lock
	l.Lock()
	println(l.TryLock())
	l.Unlock()
	println(l.TryLock())
	l.Lock()
}