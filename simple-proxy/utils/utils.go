package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"sync"
	"time"
)

var SPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 512)
	},
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func Gencertificate(private_path string, certificate_path string) error {
	// ref: https://foreverz.cn/go-cert

	// 生成私钥
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// x509证书内容
	var csr = &x509.Certificate{
		Version:      3,
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Province:           []string{"Shanghai"},
			Locality:           []string{"Shanghai"},
			Organization:       []string{"httpsDemo"},
			OrganizationalUnit: []string{"httpsDemo"},
			CommonName:         "da1234cao.top",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  false,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// 证书签名
	certDer, err := x509.CreateCertificate(rand.Reader, csr, csr, priv.Public(), priv)
	if err != nil {
		return err
	}

	// 二进制证书解析
	interCert, err := x509.ParseCertificate(certDer)
	if err != nil {
		return err
	}

	// 证书写入文件
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: interCert.Raw,
	})
	if err = ioutil.WriteFile(certificate_path, pemData, 0644); err != nil {
		panic(err)
	}

	// 私钥写入文件
	keyData := pem.EncodeToMemory(&pem.Block{
		Type:  "BEGIN RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	if err = ioutil.WriteFile(private_path, keyData, 0644); err != nil {
		panic(err)
	}

	return nil
}
