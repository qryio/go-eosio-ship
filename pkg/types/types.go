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

package types

import "github.com/greymass/go-eosio/pkg/chain"

const MaxBlockNum = 0xffffffff

type BlockPosition struct {
	BlockNum chain.BlockNum
	BlockId	 chain.Checksum256
}

type GetBlocksRequest struct {
	StartBlockNum       chain.BlockNum
	EndBlockNum         chain.BlockNum
	MaxMessagesInFlight uint32
	HavePositions       []BlockPosition
	IrreversibleOnly    bool
	FetchBlock          bool
	FetchTraces         bool
	FetchDeltas         bool
	FetchBlockHeader	bool
}

type GetBlocksResult struct {
	Head			 BlockPosition
	LastIrreversible BlockPosition
	ThisBlock		 *BlockPosition
	PrevBlock		 *BlockPosition
	Block			 chain.Bytes
	BlockHeader		 chain.Bytes
	Traces			 chain.Bytes
	Deltas			 chain.Bytes
}

type GetBlocksAckRequest struct {
	NumMessages	uint32
}

type GetStatusResult struct {
	Head				 BlockPosition
	LastIrreversible	 BlockPosition
	TraceBeginBlock		 chain.BlockNum
	TraceEndBlock		 chain.BlockNum
	ChainStateBeginBlock chain.BlockNum
	ChainStateEndBlock	 chain.BlockNum
	ChainId				 chain.Checksum256
}
