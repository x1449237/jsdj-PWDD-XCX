package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// AESEncrypt 使用 AES-256-GCM 加密明文，返回 base64 编码的密文
// key 必须为 32 字节(AES-256)
func AESEncrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("AES-256 密钥长度必须为 32 字节")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// nonce 拼接在密文前
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt 解密 AES-256-GCM 加密的 base64 密文
func AESDecrypt(encoded string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("AES-256 密钥长度必须为 32 字节")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("密文长度不足")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// PadKey 将任意长度密钥填充/截断为 32 字节(AES-256)
func PadKey(key string) []byte {
	k := []byte(key)
	if len(k) >= 32 {
		return k[:32]
	}
	pad := make([]byte, 32)
	copy(pad, k)
	return pad
}
