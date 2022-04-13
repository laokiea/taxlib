package taxintern

import (
	"runtime"
	"sync"
	"unsafe"
)

type Value struct {
	val interface{}
	// 字段表示指向Value的多个*Value指针不会因为gc回收导致use-after-free的问题
	// 当这个字段为true的时候，我们需要重新调用runtime.SetFinalizer来跳过一轮gc
	resurrect bool
}

type key struct {
	val interface{}
}

var (
	mu sync.Mutex
	// gm sync.Map
	gm = map[key]uintptr{}
	c  any
)

func Get(val interface{}) *Value {
	mu.Lock()
	defer mu.Unlock()
	var v *Value
	if i, ok := gm[key{val: val}]; ok {
		v = (*Value)(unsafe.Pointer(i))
	}
	if v != nil {
		v.resurrect = true
		return v
	}
	k := key{val: val}
	v = &Value{val: val}
	runtime.SetFinalizer(v, finalizer)
	gm[k] = uintptr(unsafe.Pointer(v))
	return v
}

func finalizer(v *Value) {
	mu.Lock()
	defer mu.Unlock()
	if v.resurrect {
		v.resurrect = false
		// skip GC collect
		runtime.SetFinalizer(v, finalizer)
		return
	}
	delete(gm, key{val: v.val})
}
