package encoder

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"

	"github.com/fatkulllin/metrilo/internal/logger"
	"go.uber.org/zap"
)

// BuildPacket формирует бинарный пакет: [RSA key][nonce][ciphertext]
func BuildPacket(publicKey *rsa.PublicKey, gzipped []byte, label []byte) ([]byte, error) {
	// 1. Случайный AES-ключ (32B)
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("ошибка генерации AES ключа: %w", err)
	}

	// 2. Случайный nonce (12B для AES-GCM)
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("ошибка генерации nonce: %w", err)
	}

	// 3. AES-GCM шифрование
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка AES блока: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка AES-GCM: %w", err)
	}
	ciphertext := aesgcm.Seal(nil, nonce, gzipped, nil)

	// 4. RSA-шифрование AES-ключа
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, label)
	if err != nil {
		return nil, fmt.Errorf("ошибка RSA-шифрования AES ключа: %w", err)
	}

	// 5. Собираем пакет: [RSA key][nonce][ciphertext]
	var buf bytes.Buffer
	buf.Write(encryptedKey) // 256B для RSA-2048
	buf.Write(nonce)        // 12B
	buf.Write(ciphertext)   // переменной длины

	return buf.Bytes(), nil
}

// ParsePacket разбирает бинарный пакет [RSA key][nonce][ciphertext]
func ParsePacket(privateKey *rsa.PrivateKey, packet []byte, label []byte) ([]byte, error) {
	// Минимальная длина: 256 (RSA key) + 12 (nonce)
	if len(packet) < 256+12 {
		return nil, fmt.Errorf("пакет слишком короткий: %d байт", len(packet))
	}

	// 1. Делим пакет на куски
	encryptedKey := packet[:256]  // фиксированный размер для RSA-2048
	nonce := packet[256 : 256+12] // фиксированный размер
	ciphertext := packet[256+12:] // остаток

	// 2. Расшифровка AES-ключа RSA-приватным ключом
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, label)
	if err != nil {
		return nil, fmt.Errorf("ошибка RSA-дешифрования AES ключа: %w", err)
	}

	// 3. AES-GCM дешифровка ciphertext
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка AES блока: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка AES-GCM: %w", err)
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка AES-дешифрования данных: %w", err)
	}

	return plaintext, nil // это gzip-сжатый JSON
}

func DecodeMiddleware(privateKey *rsa.PrivateKey, label []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

			logger.Log.Debug("DecodeMiddleware triggered", zap.String("path", req.URL.Path))

			packet, err := io.ReadAll(req.Body)
			if err != nil {
				logger.Log.Error("failed to read body", zap.Error(err))
				http.Error(res, "failed to read body", http.StatusBadRequest)
				return
			}
			plaintext, err := ParsePacket(privateKey, packet, label)
			if err != nil {
				logger.Log.Error("failed to parse packet", zap.Error(err))
				http.Error(res, "failed to parse packet", http.StatusBadRequest)
				return
			}
			fmt.Println(string(plaintext))

			// Подменяем тело запроса на расшифрованное (gzip-сжатый JSON)
			req.Body = io.NopCloser(bytes.NewReader(plaintext))

			next.ServeHTTP(res, req)
		})
	}
}
