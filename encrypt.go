package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

const (
	key = "u2oh6Vu^HWe4_AES"
	iv  = "u2oh6Vu^HWe4_AES"
	pattern = "%sd`~7^/>N4!Q#){''"
)

func PKCS7Pad(src []byte, blockSize int) ([]byte, error) {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...), nil
}

func AES_Encrypt(data string) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err.Error())
	}

	plaintext := []byte(data)
	paddedPlaintext, err := PKCS7Pad(plaintext, block.BlockSize())
	if err != nil {
		panic(err.Error())
	}

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

func resort(submitInfo map[string]string) []string {
	keys := make([]string, 0, len(submitInfo))
	for k := range submitInfo {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func enc(submitInfo map[string]string) string {
	keys := resort(submitInfo)
	var needed []string
	for _, key := range keys {
		value := submitInfo[key]
		needed = append(needed, fmt.Sprintf("[%s=%s]", key, value))
	}
	needed = append(needed, fmt.Sprintf("[%s]", pattern))

	seq := strings.Join(needed, "")
	hash := md5.Sum([]byte(seq))
	return hex.EncodeToString(hash[:])
}