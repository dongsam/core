package main

import (
	"encoding/binary"
	"fmt"
	costypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tm-db"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"github.com/terra-project/core/app"
	"log"
	"os"
	"path"
	"strings"
)

var cdc = amino.NewCodec()

// go build -tags "netgo" -ldflags '-X github.com/terra-project/core/version.Version=0.2.6-17-g45f1a6c -X github.com/terra-project/core/version.Commit=45f1a6ca98083f3da513fe901394e9aa2939d8a9 -X "github.com/terra-project/core/version.BuildTags=netgo"' -o build/terratxextract ./cmd/terratxextract
func main() {
	run()

}

func int642Bytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func isTagKey(key []byte) bool {
	return strings.Count(string(key), "/") == 3
}


func run() {
	var cnt, errCnt int64
	var dir, file string
	if len(os.Args) < 2 {
		fmt.Println("input target tx_index dir (without `.db`),  default: <HOME>/.terrad/data/tx_index")
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Errorf("error get home dir %v", err)
		}
		dir = path.Join(home, ".terrad/data")
		//dir = "~/.terrad/data"
		file = "tx_index"
	} else {
		arg := os.Args[1]
		dir, file = path.Split(arg)
	}
	fmt.Println("target_dir:", dir, "target_tx_index", file)

	// event bus
	eventBus := types.NewEventBus()
	err := eventBus.Start()
	if err != nil{
		fmt.Println(err)
	}
	defer eventBus.Stop()

	cdc := app.MakeCodec()
	txDecoder := auth.DefaultTxDecoder(cdc)

	f, err := os.OpenFile("tx_output.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	store, err := db.NewGoLevelDB(file, dir)
	if err != nil {
		fmt.Println(err)
	}
	defer store.Close()

	iter := store.Iterator(int642Bytes(0), nil)
	fmt.Println(store.Stats())
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {

		if !isTagKey(iter.Key()){
			rawBytes := iter.Value()
			if rawBytes == nil {
				continue
				errCnt++
			}

			txResult := new(types.TxResult)

			err := cdc.UnmarshalBinaryBare(rawBytes, &txResult)
			if err != nil {
				fmt.Errorf("error reading TxResult: %v", err)
				errCnt++
			}

			tx, err := txDecoder(txResult.Tx)
			if err != nil {
				fmt.Errorf("error decoding TxResult: %v", err)
				errCnt++
			} else{
				height := txResult.Height
				index := txResult.Index

				apiResults := &ctypes.ResultTx{
					Hash:     txResult.Tx.Hash(),
					Height:   height,
					Index:    index,
					TxResult: txResult.Result,
					Tx:       txResult.Tx,
					Proof:    types.TxProof{},
				}
				txRes := costypes.NewResponseResultTx(apiResults, tx, "")
				raw, err := cdc.MarshalJSON(txRes)
				if err == nil {
					jsonStr := string(raw)
					//fmt.Println(jsonStr)
					if _, err := f.WriteString(jsonStr+"\n"); err != nil {
						log.Println(err)
						errCnt++
					} else {
						cnt++
					}
				} else{
					fmt.Errorf("json decoding txRes: %v", err)
					errCnt++
				}
			}
		}
		if cnt % 100 == 1 || errCnt % 100 == 1 {
			fmt.Println("success_cnt", cnt, "err_cnt", errCnt)
		}
	}
}
