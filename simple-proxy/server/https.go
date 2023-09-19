package server

import (
	"bytes"
	"crypto/tls"
	"errors"
	"net"
	"simple-proxy/config"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func TLSDataHandle(conn net.Conn, sni string) {
	for {
		// 构建一个response
		bodyByte := []byte("hello world")
		buf := bytes.NewBuffer(nil)
		buf.WriteString("HTTP/1.1 200 OK\r\n")
		buf.WriteString("Content-Length: " + strconv.Itoa(len(bodyByte)) + "\r\n")
		buf.WriteString("\r\n")
		buf.Write(bodyByte)

		// 发送response
		buf.WriteTo(conn)
	}
}

func TLSStart() error {
	listenPort := config.Conf.LocalListen.ListenPort
	listenIp := config.Conf.LocalListen.ListenIp
	if listenPort <= 0 || listenPort > 65535 {
		log.Println("invalid listen port:", listenPort)
		return errors.New("invalid listen port")
	}

	cert, err := tls.LoadX509KeyPair(config.Conf.Certificate, config.Conf.PrivateKey)
	if err != nil {
		log.Println("fail to laod x509 key pair", err)
	}

	conf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	listener, _ := tls.Listen("tcp", listenIp+":"+strconv.Itoa(listenPort), conf)
	defer listener.Close()
	for {
		conn, _ := listener.Accept()
		tlsConn := conn.(*tls.Conn)

		// 为了拿到sni,显示的握手(通常是不需要的,在第一次read/write的时候自动调用握手)
		if err := tlsConn.Handshake(); err != nil {
			log.Println("fail to handshake with client", err)
			continue
		}

		sni := tlsConn.ConnectionState().ServerName
		// log.Debug("sni:", sni)
		go TLSDataHandle(conn, sni)
	}
}
