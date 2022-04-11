package sm2

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"github.com/SealSC/SealABC/crypto/signers/signerCommon"
	"github.com/btcsuite/btcd/btcec"
	"github.com/d5c5ceb0/sm_crypto_golang/sm2"
	"math/big"
)

const algorithmName = "sm2"

type keyPair struct {
	PrivateKey *sm2.PrivateKey
	PublicKey  *sm2.PublicKey
}

type sm2Signature struct {
	R, S *big.Int
}

func (k keyPair) Type() string {
	return algorithmName
}

func (k keyPair) Sign(data []byte) (signature []byte, err error) {
	if k.PrivateKey == nil {
		return nil, errors.New("no private key")
	}

	sig, err := k.PrivateKey.Sign(rand.Reader, data, nil)
	if err != nil {
		return
	}

	return sig, nil
}

func (k keyPair) Verify(data []byte, signature []byte) (passed bool, err error) {
	sig := sm2Signature{}
	_, err = asn1.Unmarshal(signature, &sig)
	if err != nil {
		return false, err
	}
	passed = sm2.Verify(k.PublicKey, data, sig.R, sig.S)
	return
}

func (k keyPair) RawKeyPair() (kp interface{}) {
	return nil
}

func (k keyPair) KeyPairData() (keyData []byte) {
	return nil
}

func (k keyPair) PublicKeyString() (key string) {
	keyBytes := k.PublicKeyBytes()
	return hex.EncodeToString(keyBytes)
}

func (k keyPair) PrivateKeyString() (key string) {
	keyBytes := k.PrivateKeyBytes()
	return hex.EncodeToString(keyBytes)
}

func (k keyPair) PublicKeyBytes() (key []byte) {
	if k.PublicKey == nil {
		return
	}
	pub := k.toBtcecPublickey()
	return pub.SerializeCompressed()
}

func (k keyPair) PrivateKeyBytes() (key []byte) {
	if k.PrivateKey == nil {
		return
	}
	pk := btcec.PrivateKey{}
	pk.PublicKey = ecdsa.PublicKey(k.toBtcecPublickey())
	pk.D = k.PrivateKey.D
	return pk.Serialize()
}

func (k keyPair) PublicKeyCompare(pub interface{}) (equal bool) {
	pubBytes, ok := pub.([]byte)
	if !ok {
		return false
	}

	return bytes.Equal(k.PublicKeyBytes(), pubBytes)
}

func (k keyPair) toBtcecPublickey() btcec.PublicKey {
	pub := btcec.PublicKey{}
	pub.Curve = k.PublicKey.Curve
	pub.X = k.PublicKey.X
	pub.Y = k.PublicKey.Y
	return pub
}

type keyGenerator struct{}

func (k *keyGenerator) NewSigner(_ interface{}) (signer signerCommon.ISigner, err error) {
	c := sm2.P256_sm2()
	privateKey, err := sm2.GenerateKey(c.Params(), rand.Reader)
	signer = &keyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}
	return
}

func (k *keyGenerator) FromKeyPairData(_ []byte) (signer signerCommon.ISigner, err error) {
	err = errors.New("only support gen from key bytes")
	return
}

func (k *keyGenerator) FromRawKeyPair(_ interface{}) (signer signerCommon.ISigner, err error) {
	err = errors.New("only support gen from key bytes")
	return
}

func (k *keyGenerator) FromRawPrivateKey(key interface{}) (signer signerCommon.ISigner, err error) {
	keyBytes, ok := key.([]byte)
	if !ok {
		err = errors.New("only support bytes type key")
		return
	}
	if len(keyBytes) != btcec.PrivKeyBytesLen {
		err = errors.New("invalid key size")
		return
	}

	c := sm2.P256_sm2()
	privateKey, publicKey := btcec.PrivKeyFromBytes(c.Params(), keyBytes)

	pub := sm2.PublicKey{
		Curve: publicKey.Curve,
		X:     publicKey.X,
		Y:     publicKey.Y,
	}
	pK := sm2.PrivateKey{
		PublicKey: pub,
		D:         privateKey.D,
	}
	signer = &keyPair{
		PrivateKey: &pK,
		PublicKey:  &pub,
	}
	return
}

func (k *keyGenerator) FromRawPublicKey(key interface{}) (signer signerCommon.ISigner, err error) {
	keyBytes, ok := key.([]byte)
	if !ok {
		err = errors.New("only support bytes type key")
		return
	}
	c := sm2.P256_sm2()
	var x, y *big.Int
	switch len(keyBytes) {
	case btcec.PubKeyBytesLenCompressed:
		x, y = elliptic.UnmarshalCompressed(c.Params(), keyBytes)
		break
	case btcec.PubKeyBytesLenUncompressed:
		x, y = elliptic.Unmarshal(c.Params(), keyBytes)
		break
	default:
		err = errors.New("invalid key size")
		return
	}

	if x == nil || y == nil {
		err = errors.New("only support bytes type key")
		return
	}

	pub := &sm2.PublicKey{
		Curve: c,
		X:     x,
		Y:     y,
	}

	signer = &keyPair{
		PublicKey: pub,
	}

	return
}

func (keyGenerator) Type() string {
	return algorithmName
}

var SignerGenerator = &keyGenerator{}
