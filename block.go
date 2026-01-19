package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"time"
)

// Block represents a single block in the blockchain
type Block struct {
	Timestamp     int64          // When the block was created (Unix timestamp)
	Transactions  []*Transaction // The transactions in this block
	PrevBlockHash []byte         // Hash of the previous block (creates the chain link)
	Hash          []byte         // Hash of the current block (the block's fingerprint)
	Nonce         int            // Number used in Proof of Work mining
}

// NewBlock creates and returns a new Block
// Similar to Geth's miner.worker.commitNewWork() + Seal()
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{}, // Will be calculated by PoW
		Nonce:         0,        // Will be found by PoW
	}

	// Run Proof of Work to mine the block
	// This is similar to consensus.Engine.Seal() in Geth
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

// CalculateHash calculates the hash of the block
func (b *Block) CalculateHash() []byte {
	// Combine all the block headers into one byte array
	headers := b.PrepareData()

	// Calculate SHA-256 hash
	hash := sha256.Sum256(headers)

	return hash[:]
}

// PrepareData prepares the block data for hashing
// This is where we convert all headers to bytes
func (b *Block) PrepareData() []byte {
	// Hash all transactions
	txHashes := b.HashTransactions()

	// We'll concatenate: PrevBlockHash + TxHashes + Timestamp + Nonce
	data := bytes.Join(
		[][]byte{
			b.PrevBlockHash,
			txHashes,
			IntToHex(b.Timestamp),
			IntToHex(int64(b.Nonce)),
		},
		[]byte{},
	)

	return data
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}

	return buff.Bytes()
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	txHash := sha256.Sum256(bytes.Join(transactions, []byte{}))

	return txHash[:]
}

// String returns a human-readable representation of the block
func (b *Block) String() string {
	return fmt.Sprintf("Block:\n"+
		"  Timestamp:     %d\n"+
		"  PrevBlockHash: %x\n"+
		"  Hash:          %x\n"+
		"  Transactions:  %d\n"+
		"  Nonce:         %d\n",
		b.Timestamp,
		b.PrevBlockHash,
		b.Hash,
		len(b.Transactions),
		b.Nonce,
	)
}

// Serialize serializes the block for storage
// Similar to Geth's RLP encoding (rlp.EncodeToBytes)
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block from bytes
// Similar to Geth's RLP decoding (rlp.DecodeBytes)
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		panic(err)
	}

	return &block
}
