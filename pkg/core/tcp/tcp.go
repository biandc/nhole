package tcp

import (
	"fmt"
	"net"
	"time"
)

func NewTcpListener(ip string, port int) (listener net.Listener, err error) {
	listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	return
}

func NewConner(ip string, port int) (conn net.Conn, err error) {
	conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 5*time.Second)
	return
}
