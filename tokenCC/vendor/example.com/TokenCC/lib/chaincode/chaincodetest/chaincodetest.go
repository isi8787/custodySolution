package chaincodetest

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

//MockChainCode implements chaincode interface. This is provided for unit testing Router methods in src.
type MockChainCode struct {
}

//Init method executes only once while initializing or upgrading mockchaincode
func (t *MockChainCode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke method executes everytime except on initializing or on updating the chain mockchaincode
func (t *MockChainCode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}
