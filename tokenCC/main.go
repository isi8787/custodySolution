package main

import (
	"example.com/TokenCC/lib/chaincode"
	"example.com/TokenCC/lib/util"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

func main() {
	util.ChaincodeName = "TokenCC"
	err := shim.Start(new(chaincode.ChainCode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
