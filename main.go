package main

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/joho/godotenv"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    infuraURL := os.Getenv("INFURA_URL")
    privateKeyHex := os.Getenv("PRIVATE_KEY")

    client, err := ethclient.Dial(infuraURL)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Connected to Ethereum client")

    privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        log.Fatalf("Failed to parse private key: %v", err)
    }

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        log.Fatal("Failed to cast public key to ECDSA")
    }
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    fmt.Printf("Using address: %s\n", fromAddress.Hex())

    balance, err := client.BalanceAt(context.Background(), fromAddress, nil)
    if err != nil {
        log.Fatalf("Failed to get balance: %v", err)
    }
    fmt.Printf("Account balance: %s\n", balance.String())

    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        log.Fatalf("Failed to get nonce: %v", err)
    }
    fmt.Printf("Using nonce: %d\n", nonce)

    gasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        log.Fatalf("Failed to get gas price: %v", err)
    }
    fmt.Printf("Using gas price: %s\n", gasPrice.String())

    bin, abiBytes, err := compileContract("contracts/Greeter.sol")
    if err != nil {
        log.Fatalf("Failed to compile contract: %v", err)
    }

    saveContractArtifacts(bin, abiBytes)

    parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
    if err != nil {
        log.Fatalf("Failed to parse ABI: %v", err)
    }

    auth := createTransactor(privateKey, nonce, gasPrice)

    totalCost := calculateTransactionCost(auth.GasLimit, gasPrice)
    if balance.Cmp(totalCost) < 0 {
        log.Fatalf("Insufficient funds: balance %s, needed %s", balance.String(), totalCost.String())
    }

    initialGreeting := "Hello, World!"
    contractAddress := deployOrLoadContract(client, auth, parsedABI, bin, initialGreeting)
    interactWithContract(client, contractAddress, privateKey, parsedABI)
}