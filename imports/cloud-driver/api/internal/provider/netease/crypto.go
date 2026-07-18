package netease

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
)

const (
	weapiPresetKey = "0CoJUm6Qyw8W8jud"
	weapiIV        = "0102030405060708"
	weapiKeyLength = 16
)

const weapiPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDgtQn2JZ34ZC28NWYpAUd98iZ37BUrX/aKzmFbt7clFSs6sXqHauqKWqdtLkF2KexO40H1YTX8z2lSgBBOAxLsvaklV8k4cBFK9snQXE9/DDaFt6Rr7iVZMldczhC0JNgTz+SHXT6CBHuX3e9SdB1Ua44oncaTWz7OBGLbCiK45wIDAQAB
-----END PUBLIC KEY-----`

type encryptedWEAPI struct {
	Params    string
	EncSecKey string
}

func parseWEAPIPublicKey() (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(weapiPublicKeyPEM))
	if block == nil {
		return nil, errors.New("decode public key")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("parse public key")
	}
	publicKey, ok := parsed.(*rsa.PublicKey)
	if !ok || publicKey.N == nil || publicKey.E <= 0 {
		return nil, errors.New("invalid public key")
	}
	return publicKey, nil
}

func encryptWEAPI(plaintext []byte, publicKey *rsa.PublicKey, random io.Reader) (encryptedWEAPI, error) {
	secretKey, err := randomBase62(random, weapiKeyLength)
	if err != nil {
		return encryptedWEAPI{}, err
	}
	firstPass, err := encryptAESCBC(plaintext, []byte(weapiPresetKey), []byte(weapiIV))
	if err != nil {
		return encryptedWEAPI{}, err
	}
	secondPass, err := encryptAESCBC([]byte(firstPass), secretKey, []byte(weapiIV))
	if err != nil {
		return encryptedWEAPI{}, err
	}
	encSecKey, err := rawRSAEncrypt(reverseBytes(secretKey), publicKey)
	if err != nil {
		return encryptedWEAPI{}, err
	}
	return encryptedWEAPI{Params: secondPass, EncSecKey: encSecKey}, nil
}

func randomBase62(random io.Reader, length int) ([]byte, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	upperBound := big.NewInt(int64(len(alphabet)))
	for index := range result {
		value, err := cryptorand.Int(random, upperBound)
		if err != nil {
			return nil, errors.New("generate secret key")
		}
		result[index] = alphabet[value.Int64()]
	}
	return result, nil
}

func encryptAESCBC(plaintext, key, initializationVector []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.New("create AES cipher")
	}
	padded := addPKCS7Padding(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, initializationVector).CryptBlocks(ciphertext, padded)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func addPKCS7Padding(plaintext []byte, blockSize int) []byte {
	paddingLength := blockSize - len(plaintext)%blockSize
	padded := make([]byte, len(plaintext)+paddingLength)
	copy(padded, plaintext)
	for index := len(plaintext); index < len(padded); index++ {
		padded[index] = byte(paddingLength)
	}
	return padded
}

func rawRSAEncrypt(message []byte, publicKey *rsa.PublicKey) (string, error) {
	messageInteger := new(big.Int).SetBytes(message)
	if messageInteger.Cmp(publicKey.N) >= 0 {
		return "", errors.New("RSA message is too large")
	}
	exponent := big.NewInt(int64(publicKey.E))
	ciphertext := new(big.Int).Exp(messageInteger, exponent, publicKey.N)
	encoded := make([]byte, publicKey.Size())
	ciphertext.FillBytes(encoded)
	return hex.EncodeToString(encoded), nil
}

func reverseBytes(value []byte) []byte {
	reversed := make([]byte, len(value))
	for index := range value {
		reversed[len(value)-1-index] = value[index]
	}
	return reversed
}
