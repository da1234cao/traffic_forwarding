# 目标

- [x] 一个proxy：将传入的sock5流量，使用https转发出去。
- [ ] 一个vpn
- [ ] 一个UI界面

# 使用

## simple-proxy

代码分别运行在服务端和客户端。通过其进行流量转发，可以顺利访问youtube,google等国外网站。

```shell
git clone git@github.com:da1234cao/traffic_forwarding.git
cd traffic_forwarding/simple-proxy
go mod tidy
```

客户端使用下面配置。运行`go run . --config=client.json`
* 类型为客户端,日志等级为错误
* 监听本地的10000端口
* 转发到服务端的10001端口,不验证服务端的证书
* 允许加密sni。使用EsniKey进行aes加密。EsniKey长度必须为16个字节。

```json
{
    "Type": "client",
    "LogLevel": "error",
    "LocalListen": {
        "ListenIp": "0.0.0.0",
        "ListenPort": 10000
    },
    "NextHop": {
        "skipVerify": true,
        "ServerIp": "YOUR SERVER ADDRESS",
        "ServerPort": 10001
    },
    "Esni": true,
    "EsniKey": "12345678abcdefgh"
}
```

服务端使用下面配置。运行`go run . --config=server.json`

* 类型为服务端
* 监听本地的10001端口
* 使用的证书和私钥路径。当指定路径的公钥和私钥不存在，自动自签名生成一份。
* 允许加密sni。使用EsniKey进行aes加密。EsniKey长度必须为16个字节。

```json
{
    "Type": "server",
    "LogLevel": "error",
    "LocalListen": {
        "ListenIp": "0.0.0.0",
        "ListenPort": 10001
    },
    "PrivateKey": "./key.pem",
    "Certificate": "./cert.pem",
    "Esni": true,
    "EsniKey": "12345678abcdefgh"
}
```

流量转发流程：

1. 浏览器安装SwitchyOmega插件，使用sock5代理协议。将要访问的地址，发送给客户端。
2. 客户端和服务端进行tls握手。浏览器要访问的地址，通过sni，从客户端传递到服务端。
3. 服务端与浏览器要访问的目标地址，三次握手建立连接。(服务器不需要和目标地址进行tls握手。tls握手是浏览器和目标地址之间的事情。tls在tcp之上。tcp代理不要过上层事情)
4. 上面打通tcp链路后，即可进行数据传输。浏览器<-->客户端<-->服务端<-->目标地址。

# 文档

我看过C++实现的正向代理。所以，很容易使用go实现一个正向代理。写这个项目的时候，我go也是从头学的。按照下面顺序，我们可以一步步使用go实现一个正向代理。

1. [go入门实践一-go语言的hello-world入门](https://blog.csdn.net/sinat_38816924/article/details/131901629)
2. [go入门实践二-tcp服务端](https://da1234cao.blog.csdn.net/article/details/132115292)
3. [go入门实践三-go日志库-Logrus入门教程](https://da1234cao.blog.csdn.net/article/details/132198049)
4. [go入门实践四-go实现一个简单的tcp-socks5代理服务](https://da1234cao.blog.csdn.net/article/details/132262289)

# 参考

[Subsocks: 用 Go 实现一个 Socks5 安全代理](https://luyuhuang.tech/2020/12/02/subsocks.html)

# 最后

这个项目的其他部分暂时不写了。
* vpn功能实现可参考clash和gosh，内部调用wireguard提供vpn功能。
* UI界面显示实时流量情况。需要实现一个功能，其可以跨平台的监听某一个端口的流量情况。