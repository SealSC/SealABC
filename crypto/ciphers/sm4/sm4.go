package sm4

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"github.com/SealSC/SealABC/crypto/ciphers/cipherCommon"
	"github.com/d5c5ceb0/sm_crypto_golang/sm4"
	"io"
)

type sm4Cipher struct{}

var encrypterBuilder = map[string]interface{}{
	cipherCommon.CBC: CBCEncrypt,
	cipherCommon.ECB: ECBEncrypt,
}

var decrypterBuilder = map[string]interface{}{
	cipherCommon.CBC: CBCDecrypt,
	cipherCommon.ECB: ECBDecrypt,
}

func (c sm4Cipher) Type() string {
	return "sm4"
}

func (c sm4Cipher) Encrypt(plainText []byte, key []byte, encMode interface{}) (result cipherCommon.EncryptedData, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	if len(key) != sm4.BlockSize {
		err = errors.New("invalid key")
		return
	}

	mode, ok := encMode.(string)
	if !ok {
		err = errors.New("invalid parameter")
		return
	}

	iv := make([]byte, sm4.BlockSize)
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return
	}

	encryptMethod, exist := encrypterBuilder[mode]
	if !exist {
		err = errors.New("not supported mode: " + mode)
		return
	}

	if cipherCommon.CBC == mode {
		method := encryptMethod.(func([]byte, []byte, []byte) ([]byte, error))
		result.CipherText, err = method(key, iv, PKCS5Padding(plainText, sm4.BlockSize))
	} else if cipherCommon.ECB == mode {
		method := encryptMethod.(func([]byte, []byte) ([]byte, error))
		result.CipherText, err = method(key, PKCS5Padding(plainText, sm4.BlockSize))
	}

	result.ExternalData = iv
	return
}

func (c sm4Cipher) Decrypt(cipherText cipherCommon.EncryptedData, key []byte, encMode interface{}) (plaintext []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	if len(key) != sm4.BlockSize {
		err = errors.New("invalid key")
		return
	}

	mode, ok := encMode.(string)
	if !ok {
		err = errors.New("invalid parameter")
		return
	}

	iv := cipherText.ExternalData
	if !ok {
		err = errors.New("not valid iv")
		return
	}

	decryptMethod, exist := decrypterBuilder[mode]
	if !exist {
		err = errors.New("not supported mode: " + mode)
		return
	}

	if cipherCommon.CBC == mode {
		decrypt := decryptMethod.(func([]byte, []byte, []byte) ([]byte, error))
		plaintext, err = decrypt(key, iv, cipherText.CipherText)
		plaintext = PKCS5UnPadding(plaintext)
	} else if cipherCommon.ECB == mode {
		decrypt := decryptMethod.(func([]byte, []byte) ([]byte, error))
		plaintext, err = decrypt(key, cipherText.CipherText)
		plaintext = PKCS5UnPadding(plaintext)
	}
	return
}

func CBCEncrypt(key, iv, plainText []byte) (cipherText []byte, err error) {
	plainTextLen := len(plainText)
	if plainTextLen%sm4.BlockSize != 0 {
		return nil, errors.New("input not full blocks")
	}

	c, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	encrypter := cipher.NewCBCEncrypter(c, iv)
	cipherText = make([]byte, plainTextLen)
	encrypter.CryptBlocks(cipherText, plainText)
	return cipherText, nil
}

func ECBEncrypt(key, plainText []byte) (cipherText []byte, err error) {
	plainTextLen := len(plainText)
	if plainTextLen%sm4.BlockSize != 0 {
		return nil, errors.New("input not full blocks")
	}

	c, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipherText = make([]byte, plainTextLen)
	for i := 0; i < plainTextLen; i += sm4.BlockSize {
		c.Encrypt(cipherText[i:i+sm4.BlockSize], plainText[i:i+sm4.BlockSize])
	}
	return cipherText, nil
}

func CBCDecrypt(key, iv, cipherText []byte) (plainText []byte, err error) {
	cipherTextLen := len(cipherText)
	if cipherTextLen%sm4.BlockSize != 0 {
		return nil, errors.New("input not full blocks")
	}

	c, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCBCDecrypter(c, iv)
	plainText = make([]byte, len(cipherText))
	decrypter.CryptBlocks(plainText, cipherText)
	return plainText, nil
}

func ECBDecrypt(key, cipherText []byte) (plainText []byte, err error) {
	cipherTextLen := len(cipherText)
	if cipherTextLen%sm4.BlockSize != 0 {
		return nil, errors.New("input not full blocks")
	}

	c, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plainText = make([]byte, cipherTextLen)
	for i := 0; i < cipherTextLen; i += sm4.BlockSize {
		c.Decrypt(plainText[i:i+sm4.BlockSize], cipherText[i:i+sm4.BlockSize])
	}
	return plainText, nil
}

func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unPadding := int(src[length-1])
	return src[:(length - unPadding)]
}

var Sm4 sm4Cipher
