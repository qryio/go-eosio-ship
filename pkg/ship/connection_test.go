// Copyright 2022 Thiago Souza <tcostasouza@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ship

import (
	"github.com/greymass/go-eosio/pkg/chain"
	"github.com/test-go/testify/assert"
	test "github.com/tsouza/go-eosio-ship/pkg/testing"
	"github.com/tsouza/go-eosio-ship/pkg/types"
	"testing"
)

func TestConnection_Open(t *testing.T) {
	sp, _ := test.StartEosio(t)
	defer test.StopEosio(t)
	conn := NewConnection("localhost", sp)
	err := conn.Open()
	if err != nil {
		t.Fatal(err)
	}
	for {
		select {
		case a, ok := <-conn.Ready:
			if ok {
				assert.NotNil(t, a)
				return
			}
		case err, ok := <-conn.Error:
			if ok {
				t.Fatal(err)
				return
			}
		}
	}
}

func TestConnection_SendStatusRequest(t *testing.T) {
	sp, _ := test.StartEosio(t)
	defer test.StopEosio(t)
	conn := NewConnection("localhost", sp)
	err := conn.Open()
	if err != nil {
		t.Fatal(err)
	}
	var rs *types.GetStatusResult
	for {
		select {
		case _, ok := <-conn.Ready:
			if ok {
				err = conn.SendStatusRequest()
				if err != nil {
					t.Fatal(err)
				}
			}
		case s, ok := <-conn.Status:
			if ok {
				rs = s
				break
			}
		case err, ok := <-conn.Error:
			if ok {
				t.Fatal(err)
				return
			}
		}
		if rs != nil {
			break
		}
	}
	assert.NotNil(t, rs)
	assert.Equal(t, chain.BlockNum(2), rs.TraceBeginBlock)
}

func TestConnection_SendBlocksRequest(t *testing.T) {
	sp, _ := test.StartEosio(t)
	defer test.StopEosio(t)
	conn := NewConnection("localhost", sp)
	err := conn.Open()
	if err != nil {
		t.Fatal(err)
	}
	var rb *types.GetBlocksResult
	for {
		select {
		case _, ok := <-conn.Ready:
			if ok {
				err = conn.SendBlocksRequest(&types.GetBlocksRequest{
					StartBlockNum:       2,
					EndBlockNum:         3,
					MaxMessagesInFlight: 1,
					HavePositions:       nil,
					IrreversibleOnly:    false,
					FetchBlock:          true,
					FetchTraces:         false,
					FetchDeltas:         false,
					FetchBlockHeader:    false,
				})
				if err != nil {
					t.Fatal(err)
				}
			}
		case b, ok := <-conn.Blocks:
			if ok {
				rb = b
				break
			}
		case err, ok := <-conn.Error:
			if ok {
				t.Fatal(err)
				return
			}
		}
		if rb != nil {
			break
		}
	}
	assert.NotNil(t, rb)
	assert.Equal(t, chain.BlockNum(2), rb.ThisBlock.BlockNum)
}

func TestConnection_SendAckBlocksRequest(t *testing.T) {
	sp, _ := test.StartEosio(t)
	defer test.StopEosio(t)
	conn := NewConnection("localhost", sp)
	err := conn.Open()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	var rb *types.GetBlocksResult
	for {
		select {
		case _, ok := <-conn.Ready:
			if ok {
				err = conn.SendBlocksRequest(&types.GetBlocksRequest{
					StartBlockNum:       2,
					EndBlockNum:         4,
					MaxMessagesInFlight: 1,
					HavePositions:       nil,
					IrreversibleOnly:    false,
					FetchBlock:          true,
					FetchTraces:         false,
					FetchDeltas:         false,
					FetchBlockHeader:    false,
				})
				if err != nil {
					t.Fatal(err)
				}
			}
		case b, ok := <-conn.Blocks:
			if ok {
				rb = b
				count++
				if count < 2 {
					err = conn.SendAckBlocksRequest(&types.GetBlocksAckRequest{
						NumMessages: 1,
					})
					if err != nil {
						t.Fatal(err)
					}
				} else {
					break
				}
			}
		case err, ok := <-conn.Error:
			if ok {
				t.Fatal(err)
				return
			}
		}
		if count == 2 {
			break
		}
	}
	assert.NotNil(t, rb)
	assert.Equal(t, chain.BlockNum(3), rb.ThisBlock.BlockNum)
}
