# go-eosio-ship [![GoDoc](https://godoc.org/github.com/qryio/go-eosio-ship?status.svg)](https://godoc.org/github.com/qryio/go-eosio-ship) [![Test](https://github.com/qryio/go-eosio-ship/actions/workflows/test.yml/badge.svg)](https://github.com/qryio/go-eosio-ship/actions/workflows/test.yml)

`go-eosio-ship` is a golang library built on top of [go-eosio](https://github.com/greymass/go-eosio) for connecting to and consuming from [EOSIO State History Plugin (SHiP)](https://developers.eos.io/manuals/eos/v2.1/nodeos/plugins/state_history_plugin/index).

## Usage

```go
package main

import (
	"bytes"
	"fmt"
	"github.com/greymass/go-eosio/pkg/chain"
	"github.com/qryio/go-eosio-ship/pkg/ship"
	"github.com/qryio/go-eosio-ship/pkg/types"
)

func main() {
	var shipAbi *chain.Abi
	conn := ship.NewConnection("nodeos-address", 8080)
	err := conn.Open()
	if err != nil {
		panic(err)
	}
	for {
		select {
		case a, ok := <-conn.Ready:
			if ok {
				shipAbi = a
				// request current state
				err = conn.SendStatusRequest()
				if err != nil {
					panic(err)
				}
			}
		case status, ok := <-conn.Status:
			if ok {
				// request blocks
				err = conn.SendBlocksRequest(&types.GetBlocksRequest{
					StartBlockNum:       status.ChainStateEndBlock,
					EndBlockNum:         types.MaxBlockNum,
					MaxMessagesInFlight: 1,
					HavePositions:       nil,
					IrreversibleOnly:    false,
					FetchBlock:          true,
					FetchTraces:         true,
					FetchDeltas:         true,
					FetchBlockHeader:    false,
				})
				if err != nil {
					panic(err)
				}
			}
		case blocks, ok := <-conn.Blocks:
			if ok {
				// do something with blocks request result, like decoding deltas
				deltas, err := shipAbi.Decode(bytes.NewReader(blocks.Deltas), "table_delta[]")
				if err != nil {
					panic(err)
				}
				fmt.Printf("%v\n", deltas)
				
				// request more
				err = conn.SendAckBlocksRequest(&types.GetBlocksAckRequest{
					NumMessages: 1,
				})
				if err != nil {
					panic(err)
				}
			}
		case err, ok := <- conn.Error:
			if ok {
				// oops!
				panic(err)
			}
		}
	}
}
```
## License

Code and documentation released under [Apache License 2.0](LICENSE)