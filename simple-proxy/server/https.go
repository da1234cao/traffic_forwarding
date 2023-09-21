package server

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"simple-proxy/config"
	"simple-proxy/utils"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func TLSDataHandle(conn net.Conn, sni string) {
	// 与sni指向的机器三次握手,目前还不需要tls握手
	sni_conn, err := net.Dial("tcp", sni)
	if err != nil {
		log.Warn("server fail to connect ", sni)
		conn.Close()
		return
	}
	go io.Copy(conn, sni_conn)
	go io.Copy(sni_conn, conn)
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

		// 解密sni
		dec_sni := sni
		if config.Conf.Esni && len(config.Conf.EsniKey) == 16 {
			var err error
			dec_sni_byte, err := base64.StdEncoding.DecodeString(sni)
			if err != nil {
				log.Warn("fail in base decode sni " + dec_sni + err.Error())
				continue
			}
			dec_sni, err = utils.AesDecryptStr(string(dec_sni_byte), config.Conf.EsniKey)
			if err != nil {
				log.Warn("fail in aes dencrypt sni " + sni + err.Error())
				continue
			}
		}
		log.Debug("rev sni:" + dec_sni)

		go TLSDataHandle(conn, dec_sni)
	}
}
