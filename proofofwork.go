package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

// Difficulty target bits (similar to Bitcoin/Ethereum difficulty)
// Lower value = harder difficulty
// In Geth, this is called "difficulty" and is dynamically adjusted
const targetBits = 16

// maxNonce is the maximum value for nonce to prevent infinite loops
const maxNonce = math.MaxInt64

// ProofOfWork represents the proof-of-work consensus mechanism
// In Geth, this is part of the consensus.Engine interface
type ProofOfWork struct {
	block  *Block   // The block we're mining
	target *big.Int // The target threshold (difficulty)
}

// NewProofOfWork creates a new ProofOfWork instance
// Similar to Geth's ethash.New() or clique.New()
func NewProofOfWork(b *Block) *ProofOfWork {
	// Create the target: 1 << (256 - targetBits)
	// This creates a number with leading zeros
	// Example: targetBits=16 means hash must start with 16 zero bits
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}
	return pow
}

// prepareData prepares the data to be hashed
// In Geth, this is similar to how block headers are serialized for hashing
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	pow.block.Nonce = nonce
	return pow.block.PrepareData()
}

// Run performs the proof-of-work mining
// This is the core mining loop - similar to Geth's ethash.Seal() method
// Returns: nonce (the solution) and hash (the resulting block hash)
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining block with %d transaction(s)\n", len(pow.block.Transactions))

	// The mining loop - keep trying nonces until we find a valid hash
	for nonce < maxNonce {
		// Prepare data with current nonce
		data := pow.prepareData(nonce)

		// Calculate hash
		hash = sha256.Sum256(data)

		// Print progress every 100000 attempts (optional, for visualization)
		if nonce%100000 == 0 {
			fmt.Printf("\r%x", hash)
		}

		// Convert hash to big.Int for comparison
		hashInt.SetBytes(hash[:])

		// Check if hash is less than target (i.e., has enough leading zeros)
		// This is the "proof" - we found a nonce that produces a valid hash
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x\n", hash)
			break
		} else {
			// Try next nonce
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate validates the proof-of-work
// In Geth, this is similar to the consensus.Engine.VerifyHeader() method
// This is what other nodes do when they receive a block
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	// Recreate the hash using the block's nonce
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	// Check if the hash meets the difficulty requirement
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
