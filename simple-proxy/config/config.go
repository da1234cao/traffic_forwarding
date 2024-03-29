package config

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

type localListen struct {
	ListenIp   string `json:"ListenIp"`
	ListenPort int    `json:"ListenPort"`
}

type nextHop struct {
	SkipVerify bool   `json:"SkipVerify"`
	ServerIp   string `json:"ServerIp"`
	ServerPort int    `json:"ServerPort"`
}

type configure struct {
	Type        string      `json:"Type"`
	LogLevel    string      `json:"LogLevel"`
	LocalListen localListen `json:"LocalListen"`
	NextHop     nextHop     `json:"NextHop"`
	PrivateKey  string      `json:"PrivateKey"`
	Certificate string      `json:"Certificate"`
	Esni        bool        `json:"Esni"`
	EsniKey     string      `json:"EsniKey"`
}

var Conf configure

func LoadConfig(confPath string) {
	file, err := os.OpenFile(confPath, os.O_RDWR, 0755)
	if err != nil {
		// log.WithError(err).Error("fail to open config file")
		log.Error("fail to open config file:", err)
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&Conf)
	if err != nil {
		log.WithError(err).Error("fail to decode config file")
	}
	log.Debug(Conf)
}
