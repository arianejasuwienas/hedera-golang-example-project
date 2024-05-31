1. Install go: https://go.dev/doc/install
2. sudo apt install solc
3. https://docs.soliditylang.org/en/latest/installing-solidity.html
2. go get github.com/ethereum/go-ethereum/ethclient
3. go get github.com/ethereum/go-ethereum
   go get github.com/ethereum/go-ethereum/accounts/abi/bind
   go get github.com/ethereum/go-ethereum/crypto
4. go get -t hedera-golang-example-project
4. go get github.com/joho/godotenv
5. export GOPATH=$HOME/go
   export PATH=$PATH:$GOPATH/bin
5. go install github.com/ethereum/go-ethereum/cmd/abigen@latest
3. go build .
4. ./hedera-golang-example-project - run tests
