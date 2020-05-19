package signerCommon

type KeyPair struct{
	PrivateKey  interface{}
	PublicKey   interface{}
}

type ISigner interface {
	Type() string
	Sign(data []byte) (signature []byte)
	Verify(data []byte, signature []byte) (passed bool)
	RawKeyPair() (kp interface{})
	KeyPairData() (keyData []byte)

	PublicKeyString() (key string)
	PrivateKeyString() (key string)

	PublicKeyBytes() (key [] byte)
	PrivateKeyBytes() (key []byte)

	PublicKeyCompare(k interface{}) (equal bool)
}
