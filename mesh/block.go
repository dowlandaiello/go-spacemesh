package mesh

import (
	"bytes"
	"fmt"
	"github.com/davecgh/go-xdr/xdr2"
	"github.com/spacemeshos/go-spacemesh/address"
	"math/big"
	"time"
)

type BlockID uint64
type LayerID uint64
type TransactionId []byte

type BlockHeader struct {
	Id         BlockID
	LayerIndex LayerID
	MinerID    string
	Data       []byte
	Coin       bool
	Timestamp  int64
	BlockVotes []BlockID
	ViewEdges  []BlockID
}

type MiniBlock struct {
	BlockHeader
	TxIds []TransactionId
}

type Block struct {
	BlockHeader
	Txs []*SerializableTransaction
}

type SerializableTransaction struct {
	AccountNonce uint64
	Recipient    *address.Address
	Origin       address.Address //todo: remove this, should be calculated from sig.
	GasLimit     uint64
	Price        []byte
	Amount       []byte
	Payload      []byte
}

func (t *SerializableTransaction) AmountAsBigInt() *big.Int {
	a := &big.Int{}
	a.SetBytes(t.Amount)
	return a
}

func (t *SerializableTransaction) PriceAsBigInt() *big.Int {
	a := &big.Int{}
	a.SetBytes(t.Price)
	return a
}

func NewBlock(id BlockID, layerID LayerID, minerID string, coin bool, data []byte, ts time.Time, viewEdges []BlockID, blockVotes []BlockID, txs []*SerializableTransaction) *Block {
	b := Block{
		BlockHeader: *newBlockHeader(id, layerID, minerID, coin, data, ts.UnixNano(), viewEdges, blockVotes),
		Txs:         txs,
	}
	return &b
}

func newBlockHeader(id BlockID, layerID LayerID, minerID string, coin bool, data []byte, ts int64, viewEdges []BlockID, blockVotes []BlockID) *BlockHeader {
	b := &BlockHeader{
		Id:         id,
		LayerIndex: layerID,
		MinerID:    minerID,
		BlockVotes: blockVotes,
		ViewEdges:  viewEdges,
		Timestamp:  ts,
		Data:       data,
		Coin:       coin,
	}
	return b
}

func newMiniBlock(block *Block) *MiniBlock {
	tids := make([]TransactionId, len(block.Txs))
	for idx, tx := range block.Txs {
		//keep tx's in same order !!!
		tids[idx] = getTransactionId(tx)
	}

	b := MiniBlock{
		BlockHeader: *newBlockHeader(block.ID(), block.Layer(), block.MinerID, block.Coin, block.Data, block.Timestamp, block.ViewEdges, block.BlockVotes),
		TxIds:       tids,
	}
	return &b
}

func NewSerializableTransaction(nonce uint64, origin, recepient address.Address, amount, price *big.Int, gasLimit uint64) *SerializableTransaction {
	return &SerializableTransaction{
		AccountNonce: nonce,
		Price:        price.Bytes(),
		GasLimit:     gasLimit,
		Recipient:    &recepient,
		Origin:       origin, //todo: remove this, should be calculated from sig.
		Amount:       amount.Bytes(),
		Payload:      nil,
	}
}

func (b Block) ID() BlockID {
	return b.Id
}

func (b Block) Layer() LayerID {
	return b.LayerIndex
}

func (b *Block) AddVote(id BlockID) {
	//todo: do this in a sorted manner
	b.BlockVotes = append(b.BlockVotes, id)
}

func (b *Block) AddView(id BlockID) {
	//todo: do this in a sorted manner
	b.ViewEdges = append(b.ViewEdges, id)
}

func (b *Block) AddTransaction(sr *SerializableTransaction) {
	b.Txs = append(b.Txs, sr)
}

type Layer struct {
	blocks []*Block
	index  LayerID
}

func (l *Layer) Index() LayerID {
	return l.index
}

func (l *Layer) Blocks() []*Block {
	return l.blocks
}

func (l *Layer) Hash() []byte {
	return []byte("some hash representing the layer")
}

func (l *Layer) AddBlock(block *Block) {
	block.LayerIndex = l.index
	l.blocks = append(l.blocks, block)
}

func (l *Layer) SetBlocks(blocks []*Block) {
	l.blocks = blocks
}

func NewExistingLayer(idx LayerID, blocks []*Block) *Layer {
	l := Layer{
		blocks: blocks,
		index:  idx,
	}
	return &l
}

func getBlockHeaderBytes(bheader *MiniBlock) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, bheader); err != nil {
		return nil, fmt.Errorf("error marshalling block ids %v", err)
	}
	return w.Bytes(), nil
}

func BlockAsBytes(block Block) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, &block); err != nil {
		return nil, fmt.Errorf("error marshalling block ids %v", err)
	}
	return w.Bytes(), nil
}

func BytesAsBlockHeader(buf []byte) (*MiniBlock, error) {
	b := MiniBlock{}
	_, err := xdr.Unmarshal(bytes.NewReader(buf), &b)
	if err != nil {
		return &b, err
	}
	return &b, nil
}

func BytesAsBlock(buf []byte) (Block, error) {
	b := Block{}
	_, err := xdr.Unmarshal(bytes.NewReader(buf), &b)
	if err != nil {
		return b, err
	}
	return b, nil
}

func TransactionAsBytes(tx *SerializableTransaction) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, &tx); err != nil {
		return nil, fmt.Errorf("error marshalling block ids %v", err)
	}
	return w.Bytes(), nil
}

func BytesAsTransaction(buf []byte) (*SerializableTransaction, error) {
	b := SerializableTransaction{}
	_, err := xdr.Unmarshal(bytes.NewReader(buf), &b)
	if err != nil {
		return &b, err
	}
	return &b, nil
}

func NewExistingBlock(id BlockID, layerIndex LayerID, data []byte) *Block {
	b := Block{
		BlockHeader: BlockHeader{
			Id:         BlockID(id),
			BlockVotes: make([]BlockID, 0, 10),
			ViewEdges:  make([]BlockID, 0, 10),
			LayerIndex: LayerID(layerIndex),
			Data:       data},
	}
	return &b
}

func NewLayer(layerIndex LayerID) *Layer {
	return &Layer{
		index:  layerIndex,
		blocks: make([]*Block, 0, 10),
	}
}
