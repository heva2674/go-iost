package iserver

import (
	"fmt"
	"time"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/genesis"
	"github.com/iost-official/go-iost/consensus/snapshot"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/verifier"
)

func checkGenesis(bv global.BaseVariable) error {
	blockChain := bv.BlockChain()
	stateDB := bv.StateDB()
	conf := bv.Config()
	if !conf.Snapshot.Enable && blockChain.Length() == int64(0) { //blockchaindb is empty
		// TODO: remove the module of starting iserver from yaml.

		ilog.Infof("Genesis is not exist.")
		hash := stateDB.CurrentTag()
		if hash != "" {
			return fmt.Errorf("blockchaindb is empty, but statedb is not")
		}

		blk, err := genesis.GenGenesisByFile(stateDB, conf.Genesis)
		if err != nil {
			return fmt.Errorf("new GenGenesis failed, stop the program. err: %v", err)
		}
		err = blockChain.Push(blk)
		if err != nil {
			return fmt.Errorf("push block in blockChain failed, stop the program. err: %v", err)
		}

		err = stateDB.Flush(string(blk.HeadHash()))
		if err != nil {
			return fmt.Errorf("flush block into stateDB failed, stop the program. err: %v", err)
		}
		ilog.Infof("Created Genesis.")

		// TODO check genesis hash between config and db
		ilog.Infof("GenesisHash: %v", common.Base58Encode(blk.HeadHash()))
	}
	return nil
}

func recoverDB(bv global.BaseVariable) error {
	blockChain := bv.BlockChain()
	stateDB := bv.StateDB()
	conf := bv.Config()

	if conf.Snapshot.Enable {
		blk, err := snapshot.Load(stateDB)
		if err != nil {
			return fmt.Errorf("load block from snapshot failed. err: %v", err)
		}
		blockChain.SetLength(blk.Head.Number + 1)
	} else {
		startNumebr := int64(0)
		hash := stateDB.CurrentTag()
		ilog.Infoln("current Tag:", common.Base58Encode([]byte(hash)))
		var parent *block.Block
		if hash != "" {
			blk, err := blockChain.GetBlockByHash([]byte(hash))
			if err != nil {
				return fmt.Errorf("statedb doesn't coincides with blockchaindb. err: %v", err)
			}
			startNumebr = blk.Head.Number + 1
			parent = blk
		}
		for i := startNumebr; i < blockChain.Length(); i++ {
			blk, err := blockChain.GetBlockByNumber(i)
			if err != nil {
				return fmt.Errorf("get block by number failed, stop the pogram. err: %v", err)
			}
			v := verifier.Verifier{}
			err = v.Verify(blk, parent, stateDB, &verifier.Config{
				Mode:        0,
				Timeout:     common.SlotLength / 3 * time.Second,
				TxTimeLimit: time.Millisecond * 100,
			})
			if err != nil {
				return fmt.Errorf("verify block with VM failed, stop the pogram. err: %v", err)
			}
			parent = blk
			err = snapshot.Save(stateDB, blk)
			if err != nil {
				return err
			}
			stateDB.Commit(string(blk.HeadHash()))
			err = stateDB.Flush(string(blk.HeadHash()))
			if err != nil {
				return fmt.Errorf("flush stateDB failed, stop the pogram. err: %v", err)
			}
		}
	}
	return nil
}
