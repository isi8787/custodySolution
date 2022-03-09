package chaincode

import (
	"fmt"

	"example.com/TokenCC/lib/trxcontext"
	"example.com/TokenCC/lib/util"
	"example.com/TokenCC/src"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type ChainCode struct {
}

//Init Function Executes only once while initializing or upgrading chaincode
func (t *ChainCode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	chaincodeRouter := new(src.Router)
	chaincodeRouter.Ctx = trxcontext.GetNewCtx(stub)
	return util.ExecuteMethod(chaincodeRouter, "Init", stub, args)
}

// Invoke Function Executes everytime except on initializing or on updating the chain code
func (t *ChainCode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoking " + function)
	chaincodeRouter := new(src.Router)
	chaincodeRouter.Ctx = trxcontext.GetNewCtx(stub)
	return util.ExecuteMethod(chaincodeRouter, function, stub, args)
}
