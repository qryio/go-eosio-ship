package ship

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/greymass/go-eosio/pkg/chain"
	"github.com/qryio/go-eosio-ship/pkg/types"
	"github.com/valyala/fasthttp"
)

type Connection struct {
	Abi		*chain.Abi

	addr	string
	port	int
	wsConn	*websocket.Conn

	Error  	chan error
	Ready  chan *chain.Abi
	Blocks chan *types.GetBlocksResult
	Status chan *types.GetStatusResult
}

func NewConnection(address string, port int) *Connection {
	return &Connection{
		addr: address, port: port,
	}
}

func (c *Connection) Open() error {
	if c.wsConn != nil {
		return nil
	}
	wsConn, resp, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%v:%v", c.addr, c.port), nil)
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode != fasthttp.StatusSwitchingProtocols {
		return fmt.Errorf("unhandled ship http response %v", resp.StatusCode)
	}
	c.wsConn = wsConn
	c.Error  = make(chan error)
	c.Ready  = make(chan *chain.Abi)
	c.Blocks = make(chan *types.GetBlocksResult)
	c.Status = make(chan *types.GetStatusResult)
	go c.reader()
	return nil
}

func (c *Connection) Close() {
	if c.wsConn == nil {
		return
	}
	wsConn := c.wsConn
	c.wsConn = nil
	_ = wsConn.Close()
	close(c.Blocks)
	close(c.Status)
	close(c.Error)
}

func (c *Connection) reader() {
	for c.wsConn != nil {
		_, msg, err := c.wsConn.ReadMessage()
		if err != nil {
			c.Error <- err
			return
		}
		if c.Abi == nil {
			err := json.Unmarshal(msg, &c.Abi)
			if err != nil {
				c.Error <- err
				return
			}
			c.Ready <- c.Abi
			close(c.Ready)
		} else {
			msgType, body, err := c.ReadResult(msg)
			if err != nil {
				c.Error <- err
				return
			}
			switch msgType {
			case "get_status_result_v0":
				h := body["head"].(map[string]interface{})
				li := body["last_irreversible"].(map[string]interface{})
				c.Status <- &types.GetStatusResult{
					Head:                 types.BlockPosition{
						BlockNum: chain.BlockNum(h["block_num"].(uint32)),
						BlockId:  h["block_id"].(chain.Checksum256),
					},
					LastIrreversible:     types.BlockPosition{
						BlockNum: chain.BlockNum(li["block_num"].(uint32)),
						BlockId:  li["block_id"].(chain.Checksum256),
					},
					TraceBeginBlock:      chain.BlockNum(body["trace_begin_block"].(uint32)),
					TraceEndBlock:        chain.BlockNum(body["trace_end_block"].(uint32)),
					ChainStateBeginBlock: chain.BlockNum(body["chain_state_begin_block"].(uint32)),
					ChainStateEndBlock:   chain.BlockNum(body["chain_state_end_block"].(uint32)),
					ChainId:              body["chain_id"].(chain.Checksum256),
				}
				continue
			case "get_blocks_result_v2":
				h := body["head"].(map[string]interface{})
				li := body["last_irreversible"].(map[string]interface{})
				var pTb *types.BlockPosition
				if tb, ok := body["this_block"]; ok && tb != nil {
					mtb := tb.(map[string]interface{})
					pTb = &types.BlockPosition{
						BlockNum: chain.BlockNum(mtb["block_num"].(uint32)),
						BlockId:  mtb["block_id"].(chain.Checksum256),
					}
				}
				var pPb *types.BlockPosition
				if pb, ok := body["prev_block"]; ok && pb != nil {
					mpb := pb.(map[string]interface{})
					pPb = &types.BlockPosition{
						BlockNum:chain.BlockNum(mpb["block_num"].(uint32)),
						BlockId:  mpb["block_id"].(chain.Checksum256),
					}
				}
				c.Blocks <- &types.GetBlocksResult{
					Head:                 types.BlockPosition{
						BlockNum: chain.BlockNum(h["block_num"].(uint32)),
						BlockId:  h["block_id"].(chain.Checksum256),
					},
					LastIrreversible:     types.BlockPosition{
						BlockNum: chain.BlockNum(li["block_num"].(uint32)),
						BlockId:  li["block_id"].(chain.Checksum256),
					},
					ThisBlock:        pTb,
					PrevBlock:        pPb,
					Block:            body["block"].(chain.Bytes),
					BlockHeader:      body["block_header"].(chain.Bytes),
					Traces:           body["traces"].(chain.Bytes),
					Deltas:           body["deltas"].(chain.Bytes),
				}
				continue
			default:
				c.Error <- fmt.Errorf("unhandled message (%v)", msgType)
				return
			}
		}
	}
}

func (c *Connection) SendRequest(typeName string, body map[string]interface{}) error {
	w, err := c.wsConn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	if err = c.Abi.Encode(w, "request", []interface{}{typeName, body}); err != nil {
		_ = w.Close()
		// workaround go-eosio bug that does not handle empty body properly
		if err.Error() == expectedEmptyBodyError {
			return nil
		}
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) ReadResult(msg []byte) (string, map[string]interface{}, error) {
	r := bytes.NewReader(msg)
	d, err := c.Abi.Decode(r, "result")
	if err != nil {
		return "", nil, err
	}
	ar := d.([]interface{})
	return ar[0].(string), ar[1].(map[string]interface{}), nil
}

const expectedEmptyBodyError = "expected get_status_request_v0 found map[string]interface {}"
func (c *Connection) SendStatusRequest() error {
	return c.SendRequest("get_status_request_v0", map[string]interface{}{})
}

func (c *Connection) SendBlocksRequest(request *types.GetBlocksRequest) error {
	return c.SendRequest("get_blocks_request_v1", map[string]interface{}{
		"start_block_num": uint32(request.StartBlockNum),
		"end_block_num": uint32(request.EndBlockNum),
		"max_messages_in_flight": request.MaxMessagesInFlight,
		"have_positions": toMap(request.HavePositions),
		"irreversible_only": request.IrreversibleOnly,
		"fetch_block": request.FetchBlock,
		"fetch_traces": request.FetchTraces,
		"fetch_deltas": request.FetchDeltas,
		"fetch_block_header": request.FetchBlockHeader,
	})
}

func (c *Connection) SendAckBlocksRequest(request *types.GetBlocksAckRequest) error {
	return c.SendRequest("get_blocks_ack_request_v0", map[string]interface{}{
		"num_messages": request.NumMessages,
	})
}

func toMap(p []types.BlockPosition) []interface{} {
	if p == nil {
		return []interface{}{}
	}
	r := make([]interface{}, len(p))
	for idx, v := range p {
		r[idx] = map[string]interface{}{
			"block_num": v.BlockNum,
			"block_id": v.BlockId,
		}
	}
	return r
}