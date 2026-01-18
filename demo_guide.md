# Blockchain Visualization & Demo Scripts

## Phase 1 Demo: The Ledger Inspector
**Goal:** Visualize the link between blocks.
**Code Requirement:**
In `main.go`, implement a loop that iterates over the blockchain iterator and uses `fmt.Printf` with a template.

**Desired Output Format:**
```text
============= Block 002 =============
Hash:       8a9f...2b1
PrevHash:   b2c1...9d3
PoW:        true
Nonce:      23415
Transactions: 
   - TxID: a1b2... (Coinbase)
=====================================



# Scripts 
# 1. Create Wallets
./blockchain createwallet # Returns ADDRESS_A
./blockchain createwallet # Returns ADDRESS_B

# 2. Mine Genesis (gives coins to A)
./blockchain mine -address ADDRESS_A

# 3. Check Balance (Should be 10)
./blockchain getbalance -address ADDRESS_A 

# 4. Send Coins (A sends 5 to B)
./blockchain send -from ADDRESS_A -to ADDRESS_B -amount 5

# 5. Check Balance AGAIN (Should still be 0 for B, as it's unconfirmed!)
./blockchain getbalance -address ADDRESS_B

# 6. Mine the block
./blockchain mine -address ADDRESS_A

# 7. Final Check (Now B has 5)
./blockchain getbalance -address ADDRESS_B