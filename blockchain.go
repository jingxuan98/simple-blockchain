package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"go.etcd.io/bbolt"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const mempoolBucket = "mempool"

// Blockchain represents the blockchain with database persistence
// Similar to Geth's core.BlockChain
type Blockchain struct {
	tip []byte    // Hash of the last block in the chain (the "tip")
	db  *bbolt.DB // Database connection
}

// BlockchainIterator is used to iterate over blockchain blocks
// Similar to Geth's iterator pattern
type BlockchainIterator struct {
	currentHash []byte
	db          *bbolt.DB
}

// MineBlock mines a new block with the provided transactions
// Similar to Geth's miner.worker.commitNewWork()
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte

	// Verify all transactions
	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	// Read the last block hash from the database
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// Create and mine new block
	newBlock := NewBlock(transactions, lastHash)

	// Save the new block to database
	err = bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXO finds all unspent transaction outputs
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// AddToMempool adds a transaction to the mempool
func (bc *Blockchain) AddToMempool(tx *Transaction) {
	err := bc.db.Update(func(txn *bbolt.Tx) error {
		b := txn.Bucket([]byte(mempoolBucket))
		if b == nil {
			return errors.New("Mempool bucket does not exist")
		}

		key := tx.ID
		value := tx.Serialize()

		err := b.Put(key, value)
		return err
	})
	if err != nil {
		log.Panic(err)
	}
}

// GetMempool returns all transactions in the mempool
func (bc *Blockchain) GetMempool() []*Transaction {
	var txs []*Transaction

	err := bc.db.View(func(txn *bbolt.Tx) error {
		b := txn.Bucket([]byte(mempoolBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var tx Transaction
			decoder := gob.NewDecoder(bytes.NewReader(v))
			err := decoder.Decode(&tx)
			if err != nil {
				log.Panic(err)
			}
			txs = append(txs, &tx)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return txs
}

// ClearMempool wipes the mempool
func (bc *Blockchain) ClearMempool() {
	err := bc.db.Update(func(txn *bbolt.Tx) error {
		err := txn.DeleteBucket([]byte(mempoolBucket))
		if err != nil {
			return err
		}
		_, err = txn.CreateBucket([]byte(mempoolBucket))
		return err
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

// Iterator returns a BlockchainIterator
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}

// Next returns the next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash
	return block
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	height := 0
	bci := bc.Iterator()

	for {
		block := bci.Next()
		height++
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return height
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()
		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		// For now, blindly update the tip.
		// Ideally we should check if new block's work > current tip's work.
		// But since we sync in order (Genesis -> Tip), the last added block should be the tip.
		_ = lastBlock // suppress unused variable if needed, or remove lines above

		err = b.Put([]byte("l"), block.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = block.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// NewBlockchain creates a new Blockchain with genesis block
// Similar to Geth's core.NewBlockChain()
func NewBlockchain(address, nodeID string) *Blockchain {
	var tip []byte

	// Open database
	dbPath := fmt.Sprintf(dbFile, nodeID)
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			// No blockchain exists
			if address == "" {
				fmt.Println("No existing blockchain found. Please create one first using 'createblockchain'.")
				os.Exit(1)
			}

			// Create genesis block
			fmt.Println("No existing blockchain found. Creating a new one...")
			cbtx := NewCoinbaseTX(address, "Genesis Block")
			genesis := NewBlock([]*Transaction{cbtx}, []byte{})

			// Create bucket
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			// Store genesis block
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			// Store last block hash
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}

			// Create mempool bucket
			_, err = tx.CreateBucket([]byte(mempoolBucket))
			if err != nil {
				log.Panic(err)
			}

			tip = genesis.Hash
		} else {
			// Blockchain exists, load the tip
			tip = b.Get([]byte("l"))

			// Ensure mempool bucket exists (migration for existing DBs)
			if tx.Bucket([]byte(mempoolBucket)) == nil {
				_, err = tx.CreateBucket([]byte(mempoolBucket))
				if err != nil {
					log.Panic(err)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}
