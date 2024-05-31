package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os/exec"
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