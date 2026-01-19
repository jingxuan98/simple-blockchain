package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

const walletFile = "wallets.dat"

// Wallets stores a collection of wallets
// Similar to Geth's accounts.Manager
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()

	return &wallets, err
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())

	ws.Wallets[address] = wallet

	return address
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var walletsData map[string][]byte
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&walletsData)
	if err != nil {
		log.Panic(err)
	}

	// Reconstruct wallets from serialized data
	for address, data := range walletsData {
		// Extract private key and public key
		curve := elliptic.P256()
		privKey := new(ecdsa.PrivateKey)
		privKey.PublicKey.Curve = curve
		privKey.D = new(big.Int).SetBytes(data[:32])

		pubKey := data[32:]
		privKey.PublicKey.X = new(big.Int).SetBytes(pubKey[:len(pubKey)/2])
		privKey.PublicKey.Y = new(big.Int).SetBytes(pubKey[len(pubKey)/2:])

		wallet := &Wallet{*privKey, pubKey}
		ws.Wallets[address] = wallet
	}

	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile() {
	var content bytes.Buffer

	// Create a map to store serializable wallet data
	walletsData := make(map[string][]byte)

	for address, wallet := range ws.Wallets {
		// Serialize wallet as: privateKey bytes + publicKey bytes
		privKeyBytes := wallet.PrivateKey.D.Bytes()
		data := append(privKeyBytes, wallet.PublicKey...)
		walletsData[address] = data
	}

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(walletsData)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
