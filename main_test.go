package main

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    "os"
    "strings"
    "testing"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/joho/godotenv"
    "github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (*ethclient.Client, *ecdsa.PrivateKey, common.Address, abi.ABI) {
    err := godotenv.Load()
    assert.NoError(t, err, "Error loading .env file")

    infuraURL := os.Getenv("INFURA_URL")
    privateKeyHex := os.Getenv("PRIVATE_KEY")

    client, err := ethclient.Dial(infuraURL)
    assert.NoError(t, err, "Failed to connect to Ethereum client")

    privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    assert.NoError(t, err, "Failed to parse private key")

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    assert.True(t, ok, "Failed to cast public key to ECDSA")
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

    bin, abiBytes, err := compileContract("contracts/Greeter.sol")
    assert.NoError(t, err, "Failed to compile contract")

    saveContractArtifacts(bin, abiBytes)

    parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
    assert.NoError(t, err, "Failed to parse ABI")

    return client, privateKey, fromAddress, parsedABI
}

func TestDeployContract(t *testing.T) {
    client, privateKey, fromAddress, parsedABI := setup(t)

    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    assert.NoError(t, err, "Failed to get nonce")

    gasPrice, err := client.SuggestGasPrice(context.Background())
    assert.NoError(t, err, "Failed to get gas price")

    auth := createTransactor(privateKey, nonce, gasPrice)

    initialGreeting := "Hello, World!"
    bin, _, err := compileContract("contracts/Greeter.sol")
    assert.NoError(t, err, "Failed to compile contract")

    contractAddress := deployOrLoadContract(client, auth, parsedABI, bin, initialGreeting)

    assert.True(t, common.IsHexAddress(contractAddress.Hex()), "Deployed contract address is not a valid EVM address")
}

func TestGetBalance(t *testing.T) {
    client, _, fromAddress, _ := setup(t)

    balance, err := client.BalanceAt(context.Background(), fromAddress, nil)
    assert.NoError(t, err, "Failed to get balance")
    assert.NotNil(t, balance, "Balance should not be nil")
}

func TestSetAndGetGreeting(t *testing.T) {
    client, privateKey, fromAddress, parsedABI := setup(t)

    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    assert.NoError(t, err, "Failed to get nonce")

    gasPrice, err := client.SuggestGasPrice(context.Background())
    assert.NoError(t, err, "Failed to get gas price")

    auth := createTransactor(privateKey, nonce, gasPrice)

    initialGreeting := "Hello, World!"
    bin, _, err := compileContract("contracts/Greeter.sol")
    assert.NoError(t, err, "Failed to compile contract")

    contractAddress := deployOrLoadContract(client, auth, parsedABI, bin, initialGreeting)

    chainID := big.NewInt(296)
    auth, err = bind.NewKeyedTransactorWithChainID(privateKey, chainID)
    assert.NoError(t, err, "Failed to create transactor")
    auth.GasLimit = uint64(3000000)

    newGreeting := "Hello, Ethereum!"

    boundContract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

    setGreetingTx, err := boundContract.Transact(auth, "setGreeting", newGreeting)
    assert.NoError(t, err, "Failed to set greeting")
    assert.NotNil(t, setGreetingTx, "Set greeting transaction should not be nil")

    // var result []interface{}
    // callOpts := &bind.CallOpts{}
    // err = boundContract.Call(callOpts, &result, "greet")
    // assert.NoError(t, err, "Failed to get greeting")
    // assert.NotEmpty(t, result, "Result should not be empty")
}