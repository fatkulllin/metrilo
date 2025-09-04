package keysmanager

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func LoadPublicKey(filename string) (*rsa.PublicKey, error) {
	// Используем ReadFile потому что ключи весят мало
	file, err := os.ReadFile(filename)

	if err != nil {
		return nil, fmt.Errorf("не удалось открыть public key: %w", err)
	}

	block, _ := pem.Decode(file)

	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("файл %s не содержит PUBLIC KEY", filename)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга PUBLIC KEY: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("ключ в %s не RSA public key", filename)
	}
	return rsaPub, nil

}

func LoadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	// Используем ReadFile потому что ключи весят мало
	file, err := os.ReadFile(filename)

	if err != nil {
		return nil, fmt.Errorf("не удалось открыть public key: %w", err)
	}

	block, _ := pem.Decode(file)
	if block == nil {
		return nil, fmt.Errorf("файл %s не содержит RSA PRIVATE KEY", filename)
	}

	switch block.Type {
	case "RSA PRIVATE KEY": // PKCS#1
		return x509.ParsePKCS1PrivateKey(block.Bytes)

	case "PRIVATE KEY": // PKCS#8
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("ошибка парсинга PKCS#8: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("ключ в %s не является RSA", filename)
		}
		return rsaKey, nil

	default:
		return nil, fmt.Errorf("неизвестный тип ключа: %s", block.Type)
	}

}
