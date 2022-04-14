//go:build linux

package epoll

import (
	"errors"
	"net"
	"reflect"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

type Epoll struct {
	fd    int
	conns map[int]net.Conn
	sync.RWMutex
}

var (
	ErrorSocketIsNil = errors.New("socket is nil")
)

func NewEpoll() (*Epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &Epoll{
		fd,
		make(map[int]net.Conn),
		sync.RWMutex{},
	}, nil
}

func (e *Epoll) Add(conn net.Conn) error {
	fd, err := getSocketFD(conn)
	if err != nil {
		return err
	}
	err = unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	e.Lock()
	defer e.Unlock()
	e.conns[fd] = conn
	return nil
}

func (e *Epoll) Wait() ([]net.Conn, error) {
	events := make([]unix.EpollEvent, 100)
retry:
	n, err := unix.EpollWait(e.fd, events, 100)
	if err != nil {
		if err == unix.EINTR {
			goto retry
		}
		return nil, err
	}
	e.RLock()
	defer e.RUnlock()
	var connections []net.Conn
	for i := 0; i < n; i++ {
		conn := e.conns[int(events[i].Fd)]
		connections = append(connections, conn)
	}
	return connections, nil
}

func getSocketFD(conn net.Conn) (int, error) {
	v := reflect.ValueOf(conn)
	if v.IsNil() {
		return -1, ErrorSocketIsNil
	}
	fd := reflect.Indirect(v).FieldByName("conn").FieldByName("fd")
	pfd := reflect.Indirect(reflect.ValueOf(fd)).FieldByName("pfd")
	return int(pfd.FieldByName("Sysfd").Int()), nil
}
