package types

import "github.com/greymass/go-eosio/pkg/chain"

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
