package sm3

import (
	"github.com/SealSC/SealABC/crypto/hashes/commonHash"
	sm "github.com/d5c5ceb0/sm_crypto_golang/sm3"
)

const (
	sm3 = "sm3"
)

func Load() {
	commonHash.RegisterHashMethod(sm3, sm.New)
}

var Sm3 = &commonHash.CommonHash{HashType: sm3}
