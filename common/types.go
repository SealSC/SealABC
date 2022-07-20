package common

import "encoding/hex"

const (
	HashLength    = 32
	AddressLength = 20
)

type Hash [HashLength]byte
type Address [AddressLength]byte

func (h Hash) Bytes() []byte { return h[:] }

func HexToHash(s string) Hash { return BytesToHash(FromHex(s)) }

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
		if len(s)%2 == 1 {
			s = "0" + s
		}
		h, _ := hex.DecodeString(s)
		return h

	}
	return nil
}

func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

func (a *Address) Bytes() []byte  { return a[:] }
func (a *Address) String() string { return hex.EncodeToString(a.Bytes()) }
