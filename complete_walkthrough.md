# The Complete Blockchain Guide (Phases 1-3) 🏆

Welcome to your fully functional Proof-of-Work Blockchain implementation! This guide covers everything from mining the first block to syncing nodes over P2P.

## 🌟 Quick Start Cheat Sheet
1.  **Create Identity**: `go run . createwallet`
2.  **Start Chain**: `go run . createblockchain -address <YOUR_ADDR>`
3.  **Start Node**: `go run . startnode`
4.  **Mine**: `go run . mine -address <YOUR_ADDR>`

---

## 🏗️ Phase 1: The Core Engine
**What we built:** Immutable Block storage, Proof of Work mining, and Persistence (BoltDB).

### Verification
Check the chain integrity.
```bash
go run . printchain
```
*Expected Output:* A list of blocks linked by Previous Hash, with validated PoW nonces.

---

## 💰 Phase 2: Wallets & Transactions
**What we built:** Elliptic Curve Cryptography (Wallets), UTXO Transaction Model, Mempool.

### Verification (The Spending Demo)
1.  **Create Wallets**:
    ```bash
    go run . createwallet # (Alice) -> Save Address A
    go run . createwallet # (Bob)   -> Save Address B
    ```
2.  **Alice sends 5 coins to Bob**:
    ```bash
    go run . send -from <ADDR_A> -to <ADDR_B> -amount 5
    ```
3.  **Mine the Transaction**:
    ```bash
    go run . mine -address <ADDR_A>
    ```
4.  **Check Bob's Balance**:
    ```bash
    go run . getbalance -address <ADDR_B>
    ```
    *Result:* `Balance: 5`

---

## 🌐 Phase 3: P2P Networking
**What we built:** TCP Server, Dynamic Node Discovery, Block Synchronization Protocol (`version` -> `getblocks` -> `inv` -> `getdata`).

### Verification (The Twin Nodes Sync)
Simulate a network on your local machine using `NODE_ID`.

**Terminal A (Node 3000 - Leader):**
```bash
export NODE_ID=3000
go run . createwallet
go run . createblockchain -address <ADDR_3000>
go run . startnode
```

**Terminal B (Node 3001 - Follower):**
```bash
export NODE_ID=3001
go run . createwallet
go run . createblockchain -address <ADDR_3001> # Initial Genesis
go run . startnode
```
*Terminal B connects to Terminal A...*

**Trigger Sync:**
1.  Stop Node A (`Ctrl+C`).
2.  Mine blocks on Node A:
    ```bash
    go run . mine -address <ADDR_3000>
    ```
3.  Restart Node A (`go run . startnode`).
4.  Watch **Terminal B** logs. It will announce:
    > "Received inv command"
    > "Recevied a new block!"

**Final Check:**
On Node 3001:
```bash
go run . printchain
```
You will see the blocks mined by Node A!
