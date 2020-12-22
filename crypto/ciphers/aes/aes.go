/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package aes

import (
    "github.com/SealSC/SealABC/crypto/ciphers/cipherCommon"
    "bytes"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "errors"
    "io"
)

type aesCipher struct {}

var encrypterBuilder = map[string] interface{} {
    cipherCommon.CBC: cipher.NewCBCEncrypter,
    cipherCommon.CFB: cipher.NewCFBEncrypter,
    cipherCommon.CTR: cipher.NewCTR,
    cipherCommon.OFB: cipher.NewOFB,
}

var decrypterBuilder = map[string] interface{} {
    cipherCommon.CBC: cipher.NewCBCDecrypter,
    cipherCommon.CFB: cipher.NewCFBDecrypter,
    cipherCommon.CTR: cipher.NewCTR,
    cipherCommon.OFB: cipher.NewOFB,
}

func (a aesCipher) Type() string {
    return "AES"
}

func (a aesCipher) Encrypt(plaintext []byte, key []byte, encMode interface{}) (encrypted cipherCommon.EncryptedData, err error) {
    defer func() {
        if e := recover(); e != nil {
            err = e.(error)
        }
    }()

    mode, ok := encMode.(string)
    if !ok {
        err = errors.New("invalid parameter")
        return
    }

    iv := make([]byte, aes.BlockSize)
    _, err = io.ReadFull(rand.Reader, iv)
    if err != nil {
        return
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return
    }


    encrypterMethod, exist := encrypterBuilder[mode]
    if !exist {
        err = errors.New("not supported mode: " + mode)
        return
    }

    if cipherCommon.CBC == mode {
        builder := encrypterMethod.(func(b cipher.Block, iv []byte) cipher.BlockMode)
        encrypter := builder(block, iv)
        blockSize := block.BlockSize()
        plaintext = PKCS5Padding(plaintext, blockSize)

        encrypted.CipherText = make([]byte, len(plaintext))
        encrypter.CryptBlocks(encrypted.CipherText, plaintext)
    } else {
        encrypted.CipherText = make([]byte, len(plaintext))
        builder := encrypterMethod.(func(cipher.Block, []byte) cipher.Stream)
        encrypter := builder(block, iv)
        encrypter.XORKeyStream(encrypted.CipherText, plaintext)
    }

    encrypted.ExternalData = iv
    return
}

func (a aesCipher) Decrypt(encrypted cipherCommon.EncryptedData, key []byte, encMode interface{}) (plaintext []byte, err error) {
    defer func() {
        if e := recover(); e != nil {
            err = e.(error)
        }
    }()

    mode, ok := encMode.(string)
    if !ok {
        err = errors.New("invalid parameter")
        return
    }

    iv := encrypted.ExternalData
    if !ok {
        err = errors.New("not valid iv")
        return
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return
    }

    decryptMethod, exist := decrypterBuilder[mode]
    if !exist {
        err = errors.New("not supported mode: " + mode)
        return
    }

    plaintext = make([]byte, len(encrypted.CipherText))
    if cipherCommon.CBC == mode {
        builder := decryptMethod.(func(b cipher.Block, iv []byte) cipher.BlockMode)
        decrypter := builder(block, iv)
        decrypter.CryptBlocks(plaintext, encrypted.CipherText)
        plaintext = PKCS5Trimming(plaintext)
    } else {
        builder := decryptMethod.(func(cipher.Block, []byte) cipher.Stream)
        decrypter := builder(block, iv)
        decrypter.XORKeyStream(plaintext, encrypted.CipherText)
    }
    return
}


func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
    padding := blockSize - len(ciphertext)%blockSize
    padtext := bytes.Repeat([]byte{byte(padding)}, padding)
    return append(ciphertext, padtext...)
}

func PKCS5Trimming(encrypt []byte) []byte {
    padding := encrypt[len(encrypt)-1]
    return encrypt[:len(encrypt)-int(padding)]
}

var Cipher aesCipher
