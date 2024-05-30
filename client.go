package main

import (
    "bytes"
    "context"
    "crypto/ecdsa"
    "encoding/json"
    "fmt"
    "log"
    "math/big"
    "os"
    "os/exec"
    "strings"
    "io/ioutil"
    "path/filepath"
    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/joho/godotenv"
)

type CompilationResult struct {
    Contracts map[string]struct {
        ABI      json.RawMessage
        Bin      string
        Metadata string
    } `json:"contracts"`
}

func compileContract(filePath string) (string, []byte, error) {
    cmd := exec.Command("solc", "--combined-json", "abi,bin", filePath)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return "", nil, fmt.Errorf("failed to compile contract: %v", err)
    }
    var result CompilationResult
    err = json.Unmarshal(out.Bytes(), &result)
    if err != nil {
        return "", nil, fmt.Errorf("failed to parse solc output: %v", err)
    }
    for _, contract := range result.Contracts {
        abiBytes, err := json.Marshal(contract.ABI)
        if err != nil {
            return "", nil, fmt.Errorf("failed to marshal ABI: %v", err)
        }
        bin := contract.Bin
        return bin, abiBytes, nil
    }

    return "", nil, fmt.Errorf("no contract found in solc output")
}

func loadFromFile(filename string) (string, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

func saveToFile(filename string, data []byte) error {
    return ioutil.WriteFile(filename, data, 0644)
}

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
    fmt.Println("we have a connection")
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
    cacheDir := "cache"
    if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
        err = os.Mkdir(cacheDir, 0755)
        if err != nil {
            log.Fatalf("Failed to create cache directory: %v", err)
        }
    }
    err = saveToFile(filepath.Join(cacheDir, "Greeter.bin"), []byte(bin))
    if err != nil {
        log.Fatalf("Failed to save bytecode: %v", err)
    }
    err = saveToFile(filepath.Join(cacheDir, "Greeter.abi"), abiBytes)
    if err != nil {
        log.Fatalf("Failed to save ABI: %v", err)
    }
    fmt.Println("Saved ABI and bytecode to cache files.")
    parsedABI, err := abi.JSON(bytes.NewReader(abiBytes))
    if err != nil {
        log.Fatalf("Failed to parse ABI: %v", err)
    }
    bytecode := common.FromHex(bin)
    auth := bind.NewKeyedTransactor(privateKey)
    auth.Nonce = big.NewInt(int64(nonce))
    auth.Value = big.NewInt(0)
    auth.GasLimit = uint64(3000000)
    auth.GasPrice = gasPrice
    totalCost := new(big.Int).Mul(big.NewInt(int64(auth.GasLimit)), gasPrice)
    fmt.Printf("Total transaction cost: %s\n", totalCost.String())

    if balance.Cmp(totalCost) < 0 {
        log.Fatalf("Insufficient funds: balance %s, needed %s", balance.String(), totalCost.String())
    }
    initialGreeting := "Hello, World!"
    fmt.Printf("Transaction details:\n")
    fmt.Printf("  From Address: %s\n", fromAddress.Hex())
    fmt.Printf("  Nonce: %d\n", nonce)
    fmt.Printf("  Gas Price: %s\n", gasPrice.String())
    fmt.Printf("  Gas Limit: %d\n", auth.GasLimit)
    fmt.Printf("  Initial Greeting: %s\n", initialGreeting)
    addressFilePath := "cache/address"
    var contractAddress common.Address
    if _, err := os.Stat(addressFilePath); err == nil {
        address, err := loadFromFile(addressFilePath)
        if err != nil {
            log.Fatalf("Failed to load contract address: %v", err)
        }
        contractAddress = common.HexToAddress(address)
        fmt.Printf("Contract already deployed at address: %s\n", address)
    } else {
        address, tx, _, err := bind.DeployContract(auth, parsedABI, bytecode, client, initialGreeting)
        if err != nil {
            log.Fatalf("Failed to deploy contract: %v", err)
        }
        contractAddress = address

        fmt.Printf("Contract deployed to address: %s\n", address.Hex())
        fmt.Printf("Transaction hash: %s\n", tx.Hash().Hex())

        err = saveToFile(addressFilePath, []byte(address.Hex()))
        if err != nil {
            log.Fatalf("Failed to save contract address: %v", err)
        }
        fmt.Println("Saved contract address to file.")
    }
    interactWithContract(client, contractAddress, privateKey, parsedABI)
}

func interactWithContract(client *ethclient.Client, contractAddress common.Address, privateKey *ecdsa.PrivateKey, parsedABI abi.ABI) {
    chainId := big.NewInt(296)
    auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
    auth.GasLimit = uint64(3000000)
    newGreeting := "Hello, Ethereum!"
    fmt.Printf("Setting new greeting to: %s\n", newGreeting)

    boundContract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

    setGreetingTx, err := boundContract.Transact(auth, "setGreeting", newGreeting)
    if err != nil {
        log.Fatalf("Failed to set greeting: %v", err)
    }
    fmt.Printf("Set greeting transaction hash: %s\n", setGreetingTx.Hash().Hex())
    callMsg := ethereum.CallMsg{
        To:   &contractAddress,
        Data: common.FromHex("0xcfae3217"), // greeting method selector.
    }

    res, err := client.CallContract(context.Background(), callMsg, nil)
    if err != nil {
        log.Fatalf("Failed to get greeting: %v", err)
    }
    length := new(big.Int).SetBytes(res[32:64]).Int64()
    greeting := string(res[64 : 64+length])
    greeting = strings.TrimRight(greeting, "\x00")

    log.Printf("Greeting: %s", greeting)
}
