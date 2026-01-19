package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// CLI handles command line interface
// Similar to Geth's cmd/geth/main.go
type CLI struct{}

// printUsage prints usage information
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  mine -address ADDRESS - Mine a block with transactions from the mempool")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

// validateArgs validates command line arguments
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// createBlockchain creates a new blockchain DB
func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(address)
	defer bc.db.Close()

	fmt.Println("Done!")
}

// createWallet creates a new wallet
func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

// getBalance gets the balance for an address
func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(address)
	defer bc.db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := bc.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// listAddresses lists all addresses from the wallet file
func (cli *CLI) listAddresses() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

// printChain prints all blocks in the blockchain
func (cli *CLI) printChain() {
	// Need an address to open blockchain, use empty for now
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}

	addresses := wallets.GetAddresses()
	if len(addresses) == 0 {
		log.Panic("ERROR: No wallets found. Create a wallet first.")
	}

	bc := NewBlockchain(addresses[0])
	defer bc.db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Nonce: %d\n", block.Nonce)

		// Validate PoW
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))

		// Print transactions
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()

		// Stop when we reach genesis block
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

// send sends coins from one address to another (adds to mempool)
func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.AddToMempool(tx)

	fmt.Println("Success! Transaction added to Mempool.")
}

// mine mines a block with transactions from the mempool
func (cli *CLI) mine(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Miner address is not valid")
	}

	bc := NewBlockchain(address)
	defer bc.db.Close()

	var txs []*Transaction

	// Get transactions from mempool
	mempool := bc.GetMempool()

	// Check if there are transactions to mine
	if len(mempool) == 0 {
		fmt.Println("No transactions in mempool to mine.")
		return
	}

	// Verify transactions before mining
	for _, tx := range mempool {
		if bc.VerifyTransaction(tx) {
			txs = append(txs, tx)
		} else {
			fmt.Println("ERROR: Invalid transaction found in mempool")
		}
	}

	if len(txs) == 0 {
		fmt.Println("No valid transactions to mine.")
		return
	}

	// Add coinbase transaction
	cbTx := NewCoinbaseTX(address, "")
	txs = append([]*Transaction{cbTx}, txs...) // Coinbase first

	// Mine block
	newBlock := bc.MineBlock(txs)

	// Clear mempool
	bc.ClearMempool()

	fmt.Printf("Success! Mined block: %x\n", newBlock.Hash)
}

// Run parses command line arguments and executes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	mineAddress := mineCmd.String("address", "", "The address to send mining rewards to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "mine":
		err := mineCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if mineCmd.Parsed() {
		if *mineAddress == "" {
			mineCmd.Usage()
			os.Exit(1)
		}
		cli.mine(*mineAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
