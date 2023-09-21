package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
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

// aes加解密from: https://cloud.tencent.com/developer/article/1420428

// @brief:填充明文
func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

// @brief:去除填充数据
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// @brief:AES加密
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesEncryptStr(origData, key string) (string, error) {
	crypted, err := AesEncrypt([]byte(origData), []byte(key))
	return string(crypted), err
}

// @brief:AES解密
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func AesDecryptStr(origData, key string) (string, error) {
	decData, err := AesDecrypt([]byte(origData), []byte(key))
	return string(decData), err
}
