package client

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"simple-proxy/config"
	"strconv"
)

func NewTlsConn(sni string) (net.Conn, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.Conf.NextHop.SkipVerify,
		ServerName:         sni,
	}
	return tls.Dial("tcp", config.Conf.NextHop.ServerIp+":"+strconv.Itoa(config.Conf.NextHop.ServerPort), tlsConfig)
}

func SendRequest(conn net.Conn, data []byte) {
	// 构造一个请求
	buf := bytes.NewBuffer(nil)
	buf.WriteString("POST /no_thing")
	buf.WriteString(" HTTP/1.1\r\n")
	buf.WriteString("Content-Length: " + strconv.Itoa(len(data)) + "\r\n")
	buf.WriteString("\r\n")
	buf.Write(data)

	// 发送请求
	buf.WriteTo(conn)

	// 读取回复
	ioBuf := bufio.NewReader(conn)
	res, err := http.ReadResponse(ioBuf, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()
	bodyByte, _ := io.ReadAll(res.Body)
	log.Println(string(bodyByte))
}
