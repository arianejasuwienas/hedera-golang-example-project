package main

import (
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "fmt"
)

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

func saveContractArtifacts(bin string, abiBytes []byte) {
    cacheDir := "cache"
    if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
        err = os.Mkdir(cacheDir, 0755)
        if err != nil {
            log.Fatalf("Failed to create cache directory: %v", err)
        }
    }

    err := saveToFile(filepath.Join(cacheDir, "Greeter.bin"), []byte(bin))
    if err != nil {
        log.Fatalf("Failed to save bytecode: %v", err)
    }
    err = saveToFile(filepath.Join(cacheDir, "Greeter.abi"), abiBytes)
    if err != nil {
        log.Fatalf("Failed to save ABI: %v", err)
    }

    fmt.Println("Saved ABI and bytecode to cache files.")
}

func saveAddressToFile(address string) {
    addressFilePath := "cache/address"
    err := saveToFile(addressFilePath, []byte(address))
    if err != nil {
        log.Fatalf("Failed to save contract address: %v", err)
    }
    fmt.Println("Saved contract address to file.")
}
