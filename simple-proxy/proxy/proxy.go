package proxy

import (
	"bufio"
	"net"
)

type Proxy struct {
	Inbound struct {
		Reader *bufio.Reader
		Conn   net.Conn
	}
	Request struct {
		Atyp uint8
		Addr string
	}
	OutBound struct {
		Reader *bufio.Reader
		Conn   net.Conn
	}
}
