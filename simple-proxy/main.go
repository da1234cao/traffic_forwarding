package main

import (
	"flag"
	"simple-proxy/client"
	"simple-proxy/config"
	"simple-proxy/server"
	"simple-proxy/utils"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	confPath string
)

func init() {
	// 读取命令行参数
	flag.StringVar(&confPath, "config", "", "config path")

	// 初始化日志
	logger := &lumberjack.Logger{
		Filename: "logrus.log",
		// 单位是 MB
		MaxSize: 10,
		// 最大过期日志保留的个数
		MaxBackups: 3,
		// 保留过期文件的最大时间间隔,单位是天
		MaxAge: 28, //days
		// 是否需要压缩滚动日志, 使用的 gzip 压缩
		Compress: true, // disabled by default
	}
	log.SetOutput(logger)

	log.SetFormatter(&nested.Formatter{
		NoColors:      true,
		HideKeys:      true,
		ShowFullLevel: true,
	})
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Debug("a simple proxy")
	flag.Parse()
	config.LoadConfig(confPath)
	if config.Conf.Type == "client" {
		client.Start()
	} else if config.Conf.Type == "server" {
		// 检查配置中指定的公钥和私钥是否存在。不存在，则自动生成一对到配置中的位置
		if !utils.PathExists(config.Conf.PrivateKey) || !utils.PathExists(config.Conf.Certificate) {
			utils.Gencertificate(config.Conf.PrivateKey, config.Conf.Certificate)
		}
		server.TLSStart()
	}
}
