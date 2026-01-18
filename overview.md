# Project: Go-Protocol-Chain (Geth Lite)

## Motivation
I am working on task.md with the context of overview.md.

You are my Lead Protocol Engineer mentor. I want to build this blockchain from scratch to understand the fundamentals.

- Do not write the full code at once. Guide me through example Step 1: The Anatomy of a Block, do it One by One before moving to Step 2

- Execute task by task and after completion or sub-completion of each prompt, explain why we need byte conversion for the headers when calculating the hash 


## Objective
To build a distributed blockchain in Go that mimics the core architecture of Ethereum (Geth) to learn protocol engineering.

## Architecture Guidelines
The system is divided into three distinct layers:

### 1. The Core Layer (The Ledger)
* **Model:** Account-Based (Address -> Balance), NOT UTXO.
* **Consensus:** Proof of Work (Simplified Ethash).
* **Hashing:** SHA-256.
* **Database:** In-memory persistence (store blocks as files/serialized bytes).

### 2. The Cryptography Layer (Identity)
* **Elliptic Curve:** `crypto/ecdsa` (secp256k1).
* **Addressing:** Derived from Public Key (Last 20 bytes of hash).
* **Signing:** Transactions must be signed by the sender's private key.

### 3. The Network Layer (P2P)
* **Protocol:** TCP.
* **Discovery:** Manual peer addition (for simplicity).
* **Message Types:**
    * `VERSION`: Handshake (Compare chain height).
    * `ADDR`: Exchange peer lists.
    * `BLOCK`: Propagate a new block.
    * `INV`: Inventory (Announce new data).
    * `GETDATA`: Request specific block data.

## Key Terminology
* **Mempool:** A waiting area for valid but unmined transactions.
* **State:** The current snapshot of all account balances.
* **Genesis:** The hardcoded first block.