package main

import (
    "crypto/ecdsa"
    "log"
    "math/big"
    "os"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

func createTransactor(privateKey *ecdsa.PrivateKey, nonce uint64, gasPrice *big.Int) *bind.TransactOpts {
    auth := bind.NewKeyedTransactor(privateKey)
    auth.Nonce = big.NewInt(int64(nonce))
    auth.Value = big.NewInt(0)
    auth.GasLimit = uint64(3000000)
    auth.GasPrice = gasPrice
    return auth
}

func calculateTransactionCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
    return new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
}

func deployOrLoadContract(client *ethclient.Client, auth *bind.TransactOpts, parsedABI abi.ABI, bin string, initialGreeting string) common.Address {
    addressFilePath := "cache/address"

    var contractAddress common.Address
    if _, err := os.Stat(addressFilePath); err == nil {
        address, err := loadFromFile(addressFilePath)
        if err != nil {
            log.Fatalf("Failed to load contract address: %v", err)
        }
        contractAddress = common.HexToAddress(address)
        log.Printf("Contract already deployed at address: %s", address)
    } else {
        bytecode := common.FromHex(bin)
        address, tx, _, err := bind.DeployContract(auth, parsedABI, bytecode, client, initialGreeting)
        if err != nil {
            log.Fatalf("Failed to deploy contract: %v", err)
        }
        contractAddress = address

        log.Printf("Contract deployed to address: %s", address.Hex())
        log.Printf("Transaction hash: %s", tx.Hash().Hex())

        err = saveToFile(addressFilePath, []byte(address.Hex()))
        if err != nil {
            log.Fatalf("Failed to save contract address: %v", err)
        }
        log.Println("Saved contract address to file.")
    }

    return contractAddress
}