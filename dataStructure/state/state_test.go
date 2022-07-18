package state

import (
	"github.com/SealSC/SealABC/common"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/crypto/hashes"
	"github.com/SealSC/SealABC/crypto/hashes/sha3"
	"github.com/SealSC/SealABC/crypto/signers/ed25519"
	ledger "github.com/SealSC/SealABC/service/application/smartAssets/smartAssetsLedger"
	"github.com/SealSC/SealABC/storage/db"
	"github.com/SealSC/SealABC/storage/db/dbDrivers/levelDB"
	"github.com/SealSC/SealABC/storage/db/dbInterface"
	"math/big"
	"testing"
)

func Test(t *testing.T) {
	hashes.Load()
	driver, _ := db.NewKVDatabaseDriver(dbInterface.LevelDB, levelDB.Config{
		DBFilePath: "/Users/bianning/Project/Seal/TestDemo/et/temp/nodes/n1-1/chain",
	})

	cryptoTools := crypto.Tools{
		HashCalculator:  sha3.Sha256,
		SignerGenerator: ed25519.SignerGenerator,
	}

	state, _ := New(common.Hash{}, NewDatabase(driver), cryptoTools, &ledger.AccountTool{})

	stateobjaddr0 := common.BytesToAddress([]byte("so0"))
	stateobjaddr1 := common.BytesToAddress([]byte("so1"))
	var storageaddr common.Hash

	data0 := common.BytesToHash([]byte{17})
	data1 := common.BytesToHash([]byte{18})

	state.SetState(stateobjaddr0, storageaddr, data0)
	state.SetState(stateobjaddr1, storageaddr, data1)

	// db, trie are already non-empty values
	so0 := state.getStateObject(stateobjaddr0)
	so0.SetBalance(big.NewInt(42))
	so0.SetNonce(43)
	//so0.SetCode(crypto.Keccak256Hash([]byte{'c', 'a', 'f', 'e'}), []byte{'c', 'a', 'f', 'e'})
	so0.suicided = false
	so0.deleted = false
	state.setStateObject(so0)

	root, _ := state.CommitTo(driver, false)

	t.Log(root)

	t.Log(driver.Stat())
}
