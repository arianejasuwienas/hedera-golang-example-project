package main

import (
    "crypto/ecdsa"
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

func interactWithContract(client *ethclient.Client, contractAddress common.Address, privateKey *ecdsa.PrivateKey, parsedABI abi.ABI) {
    chainId := big.NewInt(296)
    auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
    if err != nil {
        log.Fatalf("Failed to create transactor: %v", err)
    }
    auth.GasLimit = uint64(3000000)

    newGreeting := "Hello, Ethereum!"
    fmt.Printf("Setting new greeting to: %s\n", newGreeting)

    boundContract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

    setGreetingTx, err := boundContract.Transact(auth, "setGreeting", newGreeting)
    if err != nil {
        log.Fatalf("Failed to set greeting: %v", err)
    }
    fmt.Printf("Set greeting transaction hash: %s\n", setGreetingTx.Hash().Hex())

    // Interact with the contract to get the greeting
    var result []interface{}
    err = boundContract.Call(nil, &result, "greet")
    if err != nil {
        log.Fatalf("Failed to get greeting: %v", err)
    }

    fmt.Printf("Greeting: %s\n", result[0].(string))
}