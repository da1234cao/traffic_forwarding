package client

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"simple-proxy/config"
	"simple-proxy/proxy"
	"simple-proxy/utils"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const SOCKS5VERSION uint8 = 5

const (
	MethodNoAuth uint8 = iota
	MethodGSSAPI
	MethodUserPass
	MethodNoAcceptable uint8 = 0xFF
)

const (
	RequestConnect uint8 = iota + 1
	RequestBind
	RequestUDP
)

const (
	RequestAtypIPV4       uint8 = iota
	RequestAtypDomainname uint8 = 3
	RequestAtypIPV6       uint8 = 4
)

const (
	Succeeded uint8 = iota
	Failure
	Allowed
	NetUnreachable
	HostUnreachable
	ConnRefused
	TTLExpired
	CmdUnsupported
	AddrUnsupported
)

func Start() error {
	// 读取配置文件中的监听地址和端口
	log.Debug("socks5 server start")
	listenPort := config.Conf.LocalListen.ListenPort
	listenIp := config.Conf.LocalListen.ListenIp
	if listenPort <= 0 || listenPort > 65535 {
		log.Error("invalid listen port:", listenPort)
		return errors.New("invalid listen port")
	}

	//创建监听
	addr, _ := net.ResolveTCPAddr("tcp", listenIp+":"+strconv.Itoa(listenPort))
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Error("fail in listen port:", listenPort, err)
		return errors.New("fail in listen port")
	}

	// 建立连接
	for {
		conn, _ := listener.Accept()
		go socks5Handle(conn)
	}
}

func socks5Handle(conn net.Conn) {
	proxy := &proxy.Proxy{}
	proxy.Inbound.Reader = bufio.NewReader(conn)
	proxy.Inbound.Conn = conn

	err := handshake(proxy)
	if err != nil {
		log.Warn("fail in handshake", err)
		return
	}
	transport(proxy)
}

func handshake(proxy *proxy.Proxy) error {
	err := auth(proxy)
	if err != nil {
		log.Warn(err)
		return err
	}

	err = readRequest(proxy)
	if err != nil {
		log.Warn(err)
		return err
	}

	err = replay(proxy)
	if err != nil {
		log.Warn(err)
		return err
	}
	return err
}

func auth(proxy *proxy.Proxy) error {
	/*
		Read
		   +----+----------+----------+
		   |VER | NMETHODS | METHODS  |
		   +----+----------+----------+
		   | 1  |    1     | 1 to 255 |
		   +----+----------+----------+
	*/
	buf := utils.SPool.Get().([]byte)
	defer utils.SPool.Put(buf)

	n, err := io.ReadFull(proxy.Inbound.Reader, buf[:2])
	if n != 2 {
		return errors.New("fail to read socks5 request:" + err.Error())
	}

	ver, nmethods := uint8(buf[0]), int(buf[1])
	if ver != SOCKS5VERSION {
		return errors.New("only support socks5 version")
	}
	_, err = io.ReadFull(proxy.Inbound.Reader, buf[:nmethods])
	if err != nil {
		return errors.New("fail to read methods" + err.Error())
	}
	supportNoAuth := false
	for _, m := range buf[:nmethods] {
		switch m {
		case MethodNoAuth:
			supportNoAuth = true
		}
	}
	if !supportNoAuth {
		return errors.New("no only support no auth")
	}

	/*
		replay
			+----+--------+
			|VER | METHOD |
			+----+--------+
			| 1  |   1    |
			+----+--------+
	*/
	n, err = proxy.Inbound.Conn.Write([]byte{0x05, 0x00}) // 无需认证
	if n != 2 {
		return errors.New("fail to wirte socks method " + err.Error())
	}

	return nil
}

func readRequest(proxy *proxy.Proxy) error {
	/*
		Read
		   +----+-----+-------+------+----------+----------+
		   |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
		   +----+-----+-------+------+----------+----------+
		   | 1  |  1  | X'00' |  1   | Variable |    2     |
		   +----+-----+-------+------+----------+----------+
	*/
	buf := utils.SPool.Get().([]byte)
	defer utils.SPool.Put(buf)
	n, err := io.ReadFull(proxy.Inbound.Reader, buf[:4])
	if n != 4 {
		return errors.New("fail to read request " + err.Error())
	}
	ver, cmd, _, atyp := uint8(buf[0]), uint8(buf[1]), uint8(buf[2]), uint8(buf[3])
	if ver != SOCKS5VERSION {
		return errors.New("only support socks5 version")
	}
	if cmd != RequestConnect {
		return errors.New("only support connect requests")
	}
	var addr string
	switch atyp {
	case RequestAtypIPV4:
		_, err = io.ReadFull(proxy.Inbound.Reader, buf[:4])
		if err != nil {
			return errors.New("fail in read requests ipv4 " + err.Error())
		}
		addr = string(buf[:4])
	case RequestAtypDomainname:
		_, err = io.ReadFull(proxy.Inbound.Reader, buf[:1])
		if err != nil {
			return errors.New("fail in read requests domain len" + err.Error())
		}
		domainLen := int(buf[0])
		_, err = io.ReadFull(proxy.Inbound.Reader, buf[:domainLen])
		if err != nil {
			return errors.New("fail in read requests domain " + err.Error())
		}
		addr = string(buf[:domainLen])
	case RequestAtypIPV6:
		_, err = io.ReadFull(proxy.Inbound.Reader, buf[:16])
		if err != nil {
			return errors.New("fail in read requests ipv4 " + err.Error())
		}
		addr = string(buf[:16])
	}
	_, err = io.ReadFull(proxy.Inbound.Reader, buf[:2])
	if err != nil {
		return errors.New("fail in read requests port " + err.Error())
	}
	port := binary.BigEndian.Uint16(buf[:2])
	proxy.Request.Atyp = atyp
	proxy.Request.Addr = fmt.Sprintf("%s:%d", addr, port)
	log.Debug("request is:", proxy.Request.Addr)
	return nil
}

func errReplay(proxy *proxy.Proxy) error {
	log.Warn("fail to connect server: ", config.Conf.NextHop.ServerIp+":"+strconv.Itoa(config.Conf.NextHop.ServerPort))
	_, rerr := proxy.Inbound.Conn.Write([]byte{SOCKS5VERSION, HostUnreachable, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if rerr != nil {
		return errors.New("fail in replay " + rerr.Error())
	}
	return nil
}

func replay(proxy *proxy.Proxy) error {
	/*
		write
		   +----+-----+-------+------+----------+----------+
		   |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		   +----+-----+-------+------+----------+----------+
		   | 1  |  1  | X'00' |  1   | Variable |    2     |
		   +----+-----+-------+------+----------+----------+
	*/

	// 对称加密sni,然后base64编码
	enc_sni := proxy.Request.Addr
	var err error
	if config.Conf.Esni && len(config.Conf.EsniKey) == 16 {
		enc_sni, err = utils.AesEncryptStr(proxy.Request.Addr, config.Conf.EsniKey)
		if err != nil {
			errReplay(proxy)
			return errors.New("fail in encrypt sni " + proxy.Request.Addr + err.Error())
		}
		enc_sni = base64.StdEncoding.EncodeToString([]byte(enc_sni))
	}

	// 与服务端建立连接
	conn, err := NewTlsConn(string(enc_sni))
	if err != nil {
		errReplay(proxy)
		return errors.New("fail in connect addr " + proxy.Request.Addr + err.Error())
	}
	_, err = proxy.Inbound.Conn.Write([]byte{SOCKS5VERSION, Succeeded, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return errors.New("fail in replay " + err.Error())
	}
	proxy.OutBound.Reader = bufio.NewReader(conn)
	proxy.OutBound.Conn = conn
	return nil
}

func transport(proxy *proxy.Proxy) {
	go io.Copy(proxy.OutBound.Conn, proxy.Inbound.Reader) // outbound <- inbound
	go io.Copy(proxy.Inbound.Conn, proxy.OutBound.Reader) // inbound <- outbound
}
