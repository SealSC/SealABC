/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the Licegogo nse at
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

package main

import (
	"encoding/hex"
	"fmt"
	"github.com/SealSC/SealABC/cli"
	"github.com/SealSC/SealABC/common/utility"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/crypto/hashes/sha3"
	"github.com/SealSC/SealABC/crypto/signers/ecdsa/secp256k1"
	"github.com/SealSC/SealABC/engine"
	"github.com/SealSC/SealABC/engine/engineStartup"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/network/topology/p2p/fullyConnect"
	//"github.com/SealSC/SealABC/service/application/basicAssets"
	"github.com/SealSC/SealABC/service/application/memo"
	"github.com/SealSC/SealABC/service/application/smartAssets"
	"github.com/SealSC/SealABC/service/application/smartAssets/smartAssetsLedger"
	"github.com/SealSC/SealABC/service/application/traceableStorage"
	//"github.com/SealSC/SealABC/service/application/universalIdentification"
	"github.com/SealSC/SealABC/service/system"
	"github.com/SealSC/SealABC/service/system/blockchain"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"github.com/SealSC/SealABC/storage/db"
	"github.com/SealSC/SealABC/storage/db/dbDrivers/levelDB"
	"github.com/SealSC/SealABC/storage/db/dbDrivers/simpleMysql"
	"github.com/SealSC/SealABC/storage/db/dbInterface"
	"github.com/sirupsen/logrus"
	"net/http"
	"runtime"
	"time"

	_ "net/http/pprof"

	appCommonCfg "github.com/SealSC/SealABC/metadata/applicationCommonConfig"
	"github.com/SealSC/SealEVM"
)

//func getPassword() string {
//    fmt.Print("enter password: ")
//    pwdBytes, err := terminal.ReadPassword(syscall.Stdin)
//    if err != nil {
//        fmt.Println("get password failed: ", err.Error())
//    }
//    pwdStr := string(pwdBytes)
//
//    fmt.Println(environment.Block{})
//    fmt.Println(evmInt256.New(0))
//    return strings.TrimSpace(pwdStr)
//}

func testHeap(prvBytes []byte) (ret []byte) {
	ret = append(ret, prvBytes...)
	blk := block.Entity{}
	blk.Header.Version = "just for test!!!!!!"
	fakePrev := make([]byte, len(prvBytes), len(prvBytes))
	copy(fakePrev, prvBytes)
	blk.Header.PrevBlock = fakePrev

	//tools := crypto.Tools{SignerGenerator: ed25519.SignerGenerator, HashCalculator: sha3.Sha256}
	tools := crypto.Tools{SignerGenerator: secp256k1.SignerGenerator, HashCalculator: sha3.Sha256}
	_ = blk.Sign(tools, prvBytes)

	blkBytes, _ := structSerializer.ToMFBytes(blk)
	_, _ = blk.Seal.Verify(blkBytes, tools.HashCalculator)

	return
}

func main() {
	cli.Run()

	SealEVM.Load()

	//runtime.GOMAXPROCS(1)
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	//runtime.SetCPUProfileRate(1)

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	utility.Load()
	crypto.Load()
	_ = SealABCConfig.Load()

	log.SetUpLogger(log.Config{
		Level:   logrus.Level(SealABCConfig.LogLevel),
		LogFile: SealABCConfig.LogFile,
	})

	//load sql db
	sqlStorage, _ := db.NewSimpleSQLDatabaseDriver(dbInterface.MySQL, simpleMysql.Config{
		User:          SealABCConfig.MySQLUser,
		Password:      SealABCConfig.MySQLPwd,
		DBName:        SealABCConfig.MySQLDBName,
		Charset:       simpleMysql.Charsets.UTF8.String(),
		MaxConnection: 0,
	})

	//common crypto tools
	cryptoTools := crypto.Tools{
		HashCalculator: sha3.Sha256,
		//SignerGenerator: ed25519.SignerGenerator,
		SignerGenerator: secp256k1.SignerGenerator,
	}

	//load self keys
	sk, _ := hex.DecodeString(SealABCConfig.SelfKey)
	//selfSigner, _ := ed25519.SignerGenerator.FromRawPrivateKey(sk)
	selfSigner, _ := secp256k1.SignerGenerator.FromRawPrivateKey(sk)

	//build consensus config
	bhtConfig := hotStuff.Config{
		MemberOnlineCheckInterval: time.Millisecond * 1000,
		ConsensusTimeout:          time.Millisecond * 10000,
		//SingerGenerator:           ed25519.SignerGenerator,
		SingerGenerator:   secp256k1.SignerGenerator,
		HashCalc:          sha3.Sha256,
		SelfSigner:        selfSigner,
		ConsensusInterval: time.Millisecond * 3000,
	}

	//build consensus member list
	for _, member := range SealABCConfig.Members {
		newMember := hotStuff.Member{}
		mk, _ := hex.DecodeString(member)
		//newMember.Signer, _ = ed25519.SignerGenerator.FromRawPublicKey(mk)
		newMember.Signer, _ = secp256k1.SignerGenerator.FromRawPublicKey(mk)
		bhtConfig.Members = append(bhtConfig.Members, newMember)
	}

	engineCfg := engineStartup.Config{}

	////set self signer
	//engineCfg.SelfSigner = selfSigner
	//
	////set crypto tools
	//engineCfg.CryptoTools = cryptoTools

	//config network
	engineCfg.ConsensusNetwork.ID = selfSigner.PublicKeyString()
	engineCfg.ConsensusNetwork.ServiceAddress = SealABCConfig.ConsensusServiceAddress
	engineCfg.ConsensusNetwork.ServiceProtocol = "tcp"
	engineCfg.ConsensusNetwork.P2PSeeds = SealABCConfig.ConsensusMember
	engineCfg.ConsensusNetwork.Topology = fullyConnect.NewTopology()

	//config consensus
	engineCfg.ConsensusDisabled = SealABCConfig.ConsensusDisabled
	engineCfg.Consensus = bhtConfig
	//engineCfg.ConsensusType = consensus.BasicHotStuff

	//config db
	//engineCfg.StorageConfig = levelDB.Config{
	//    DBFilePath: SealABCConfig.ChainDB,
	//}

	//load basic assets application
	//utxoCfg := basicAssets.Config {
	//    appCommonCfg.Config{
	//        KVDBName: dbInterface.LevelDB,
	//        KVDBConfig: levelDB.Config{
	//            DBFilePath: SealABCConfig.LedgerDB,
	//        },
	//        EnableSQLDB: SealABCConfig.EnableSQLStorage,
	//        SQLStorage:  sqlStorage,
	//    },
	//}
	//basicAssets.Load()
	//utxoService, _ := basicAssets.NewBasicAssetsApplication(utxoCfg)

	//load memo application
	memoCfg := memo.Config{
		Config: appCommonCfg.Config{
			KVDBName: dbInterface.LevelDB,
			KVDBConfig: levelDB.Config{
				DBFilePath: SealABCConfig.MemoDB,
			},
		},
	}

	memo.Load()
	memoCfg.SQLStorage = sqlStorage
	memoCfg.EnableSQLDB = SealABCConfig.EnableSQLStorage
	memoService, _ := memo.NewMemoApplication(memoCfg, cryptoTools)

	saCfg := &smartAssets.Config{
		Config: appCommonCfg.Config{
			KVDBName: dbInterface.LevelDB,
			KVDBConfig: levelDB.Config{
				DBFilePath: SealABCConfig.SmartAssetsDB,
			},

			EnableSQLDB: SealABCConfig.EnableSQLStorage,
			SQLStorage:  sqlStorage,
			CryptoTools: cryptoTools,
		},

		BaseAssets: smartAssetsLedger.BaseAssetsData{
			Name:        "Seal ABC Test Token",
			Symbol:      "SAT",
			Supply:      "3000000000000",
			Precision:   4,
			Increasable: false,
			Owner:       "c6baa18dbca0d1173cf403961119ad238df16bdaaf0a3d9038794fd0ab5ee9cf",
		},
	}

	smartAssets.Load()
	smartAssetsApplication, err := smartAssets.NewSmartAssetsApplication(saCfg)
	if err != nil {
		fmt.Println("start smart assets application failed: ", err.Error())
		return
	}

	tsCfg := traceableStorage.Config{
		Config: appCommonCfg.Config{
			KVDBName: dbInterface.LevelDB,
			KVDBConfig: levelDB.Config{
				DBFilePath: SealABCConfig.TraceableStorageDB,
			},
			EnableSQLDB: false,
			SQLStorage:  nil,
			CryptoTools: crypto.Tools{},
		},
	}
	traceableStorage.Load()
	traceableStorageApplication, err := traceableStorage.NewTraceableStorageApplication(&tsCfg)
	if err != nil {
		fmt.Println("start traceable storage application failed: ", err.Error())
		return
	}

	//uidCfg := universalIdentification.Config  {
	//   Config:      appCommonCfg.Config{
	//       KVDBName: dbInterface.LevelDB,
	//       KVDBConfig: levelDB.Config{
	//           DBFilePath: SealABCConfig.TraceableStorageDB,
	//       },
	//       EnableSQLDB: false,
	//       SQLStorage:  nil,
	//       CryptoTools: crypto.Tools{},
	//   },
	//}
	//universalIdentification.Load()
	//uidApp, err := universalIdentification.NewUniversalIdentificationApplication(uidCfg)
	//if err != nil {
	//   fmt.Println("start traceable storage application failed: ", err.Error())
	//   return
	//}

	//set application to chain service
	systemService := system.Config{}

	//config system service
	systemService.Chain = blockchain.Config{}

	systemService.Chain.Blockchain.Signer = selfSigner
	systemService.Chain.Blockchain.CryptoTools = cryptoTools
	systemService.Chain.Blockchain.NewWhenGenesis = true

	systemService.Chain.ExternalExecutors = []chainStructure.IBlockchainExternalApplication{
		//utxoService,
		memoService,
		smartAssetsApplication,
		traceableStorageApplication,
		//uidApp,
	}

	//start load chain
	systemService.Chain.Api.HttpJSON = SealABCConfig.BlockchainApiConfig

	//config blockchain system service network
	systemService.Chain.Network.ID = selfSigner.PublicKeyString()
	systemService.Chain.Network.ServiceAddress = SealABCConfig.BlockchainServiceAddress
	systemService.Chain.Network.ServiceProtocol = "tcp"
	systemService.Chain.Network.P2PSeeds = SealABCConfig.BlockchainServiceSeeds
	systemService.Chain.Network.Topology = fullyConnect.NewTopology()

	engineCfg.Log.LogFile = SealABCConfig.LogFile
	engineCfg.Log.Level = logrus.Level(SealABCConfig.LogLevel)

	engineCfg.Api.HttpJSON = SealABCConfig.EngineApiConfig

	systemService.Chain.EnableSQLDB = SealABCConfig.EnableSQLStorage
	systemService.Chain.SQLStorage = sqlStorage
	systemService.Chain.Blockchain.StorageDriver, _ = db.NewKVDatabaseDriver(dbInterface.LevelDB, levelDB.Config{
		DBFilePath: SealABCConfig.ChainDB,
	})
	engine.Startup(engineCfg)

	system.NewBlockchainService(systemService.Chain)

	for {
		time.Sleep(time.Second)
	}
}
