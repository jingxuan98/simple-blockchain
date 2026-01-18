# Implementation Roadmap & Verification

## Phase 1: The Core Engine (Storage & Mining)
* [ ] **Step 1.1:** Create `Block` struct & Hash logic.
* [ ] **Step 1.2:** Implement `ProofOfWork` (Target bits logic).
* [ ] **Step 1.3:** Create `Blockchain` struct with DB persistence.
* [ ] **Step 1.4:** **VERIFICATION (The "Visual Chain"):**
    * Create a CLI command `printchain`.
    * Output: Format the block data cleanly (e.g., "PrevHash: x... Hash: y... Nonce: 1234").
    * Success Criteria: Run the app, add 3 blocks, restart app, print chain. The blocks must still be there (persistence works).

## Phase 2: Wallets & Transactions (The State Machine)
* [ ] **Step 2.1:** Implement `Wallet` generation & Address derivation.
* [ ] **Step 2.2:** Implement `Transaction` struct & Signing logic.
* [ ] **Step 2.3:** Implement `Mempool` & Update Mining logic.
* [ ] **Step 2.4:** **VERIFICATION (The "Spending" Demo):**
    * Create CLI commands: `createwallet`, `getbalance`, `send`.
    * Success Criteria:
        1. Create two wallets (Alice, Bob).
        2. Alice sends 10 coins to Bob.
        3. `getbalance Bob` shows 0 (because it's in mempool).
        4. Mine a block.
        5. `getbalance Bob` shows 10.

## Phase 3: P2P Networking (The Nervous System)
* [ ] **Step 3.1:** Setup TCP Listener & Command Parser.
* [ ] **Step 3.2:** Implement Handshake (`VERSION` message).
* [ ] **Step 3.3:** Implement Block Propagation (`INV`, `GETDATA`, `BLOCK`).
* [ ] **Step 3.4:** **VERIFICATION (The "Twin Nodes" Demo):**
    * Create a startup flag `NODE_ID` to run multiple instances.
    * Success Criteria:
        1. Open Terminal A (Node 3000).
        2. Open Terminal B (Node 3001).
        3. Connect B to A.
        4. Mine a block on A.
        5. Watch Terminal B logs automatically show: "Received new block from ...".