package ch

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
)

// register
// unregister
// get
// hash
// hashRing
// hostMap
// hostToVHostMap

var (
	DefaultHashFunc = func(key string) uint32 {
		h := fnv.New32a()
		_, _ = h.Write([]byte(key))
		return h.Sum32()
	}
)

const ReplicaNum = 2

type Host struct {
	Name string
	Key  string
}

type ConsistentHash struct {
	sync.Mutex
	HashFunc          func(key string) uint32
	HostToVHostMap    map[uint32]string
	HostToVHostMapRev map[string][]uint32
	HostMap           map[string]*Host
	HashRing          []uint32
	ReplicaNum        int
}

type Options interface {
	Apply(c *ConsistentHash)
}

type HashFuncOpt struct {
	HashFunc func(key string) uint32
}

func (hfo *HashFuncOpt) Apply(c *ConsistentHash) {
	c.HashFunc = hfo.HashFunc
}

func (c *ConsistentHash) New(opts ...Options) {
	c.ReplicaNum = ReplicaNum
	c.HostToVHostMap = make(map[uint32]string)
	c.HostMap = make(map[string]*Host)
	c.HostToVHostMapRev = make(map[string][]uint32)
	for _, o := range opts {
		o.Apply(c)
	}
}

func (c *ConsistentHash) Join(host, name string) error {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.HostMap[host]; ok {
		// already exists
		return errors.New("already exists")
	}
	if name == "" {
		name = host
	}
	c.HostMap[host] = &Host{name, host}
	for i := 0; i < ReplicaNum; i++ {
		i := c.HashFunc(fmt.Sprintf("%s#%d", host, i))
		c.HostToVHostMap[i] = host
		c.HostToVHostMapRev[host] = append(c.HostToVHostMapRev[host], i)
		c.HashRing = append(c.HashRing, i)
	}
	sort.Slice(c.HashRing, func(i, j int) bool {
		return c.HashRing[i] < c.HashRing[j]
	})
	sort.Slice(c.HostToVHostMapRev[host], func(i, j int) bool {
		return c.HostToVHostMapRev[host][i] < c.HostToVHostMapRev[host][j]
	})
	return nil
}

func (c *ConsistentHash) Quit(host string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.HostMap[host]; !ok {
		// not exists
		return
	}
	// Map VHostMap HashRing
	delete(c.HostMap, host)
	cur := sort.Search(len(c.HashRing), func(i int) bool {
		return c.HashRing[i] >= c.HostToVHostMapRev[host][0]
	})
	for _, v := range c.HostToVHostMapRev[host] {
		delete(c.HostToVHostMap, v)
		l := len(c.HashRing)
		for cur < l {
			if c.HashRing[cur] == v {
				if cur == 0 {
					c.HashRing = c.HashRing[1:]
				} else {
					c.HashRing = append(c.HashRing[0:cur-1], c.HashRing[cur+1:]...)
				}
				break
			}
		}
	}
	delete(c.HostToVHostMapRev, host)
}

func (c *ConsistentHash) Get(key string) (string, error) {
	if len(c.HashRing) == 0 {
		return "", errors.New("hash ring is empty")
	}
	l := len(c.HashRing)
	idx := c.HashFunc(key)
	i := sort.Search(l, func(i int) bool {
		return c.HashRing[i] >= idx
	})
	if i == l {
		return c.HostToVHostMap[c.HashRing[0]], nil
	}
	if l != len(c.HashRing) {
		return c.Get(key)
	}
	return c.HostToVHostMap[c.HashRing[i]], nil
}

//func main() {
//	c := new(ConsistentHash)
//	c.New(&HashFuncOpt{HashFunc: DefaultHashFunc})
//	c.Join("10.9.97.189", "server1")
//	c.Join("10.9.24.3", "server2")
//	c.Quit("10.9.97.189")
//	c.Join("10.9.97.189", "server1")
//	fmt.Println(c.HashRing)
//	fmt.Println(c.Get("hello"))
//	fmt.Println(c.Get("halo"))
//}
