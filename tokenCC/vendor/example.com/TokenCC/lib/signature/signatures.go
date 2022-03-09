package signature


import (
	"crypto/sha256"
	"crypto/ecdsa"

	"encoding/hex"
	"encoding/json"
	b64 "encoding/base64"
	"strings"
	"fmt"
    "math/big"
	"os"
	"strconv"
	"context"

    "example.com/TokenCC/lib/ccsp"
	"example.com/TokenCC/lib/model"
	"example.com/TokenCC/lib/transaction"
	"example.com/TokenCC/lib/account"
	"github.com/coinbase/kryptology/pkg/ted25519/ted25519"

	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/paillier"
	"github.com/coinbase/kryptology/pkg/tecdsa/gg20/dealer"
	"github.com/coinbase/kryptology/pkg/tecdsa/gg20/participant"

	"github.com/dustinxie/ecc"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/btcsuite/btcutil"
	btcchain "github.com/btcsuite/btcd/chaincfg"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/params"


)



/* Define Key structure, with 4 properties.
Structure tags are used by encoding/json library
*/
type ExternalTransaction struct {
	Identifier          string					`json: "identifier" id:"true" mandatory:"true"`
	Status              string                  `json:"status"`
	AccountOwner        SignatureTransaction    `json:"accountOwner"`
	OrgAdmin            SignatureTransaction    `json:"orgAdmin"`
	OrgSignatory		SignatureTransaction	`json:"orgSignatory"`
}

type SignatureTransaction struct {
	SignerID                  string    `json:"signerID"`
	Status                    string    `json:"status"`
	Message                   string    `json:"message"`
	MessageHash               string    `json:"messageHash" id:"true" mandatory:"true"`
	ECParameters              account.ECParameters `json:"ECParameters"`
	TokenId           string    `json:"tokenId"`
	KeyAlias                 string    `json:"keyAlias"`
	SignedMessage string `json:"signedMessage"` //Need to make sure that this is the right place for signed message
	TransactionDate string  `json:"transactionDate"`
	TransactionSubmittedTimeStamp string  `json:"transactionSubmittedTimeStamp"`
}

type ThresholdSignatureTransaction struct {
	SignerID                  string    `json:"signerID"`
	Status                    string    `json:"status"`
	Message                   string    `json:"message"`
	MessageHash               string    `json:"messageHash" id:"true" mandatory:"true"`
	TokenId         		  string    `json:"tokenId"`
	SignedShares 			  []ted25519.PartialSignature    `json:"signedShares"` //Need to make sure that this is the right place for signed message
	TransactionDate			  string  `json:"transactionDate"`
	TransactionSubmittedTimeStamp string  `json:"transactionSubmittedTimeStamp"`
	NoncePub			      string	`json:"noncePub,omitempty"`
	PubKey			          string	`json:"pubKey,omitempty"`
	ShareInformation          map[string]ShareInformation  `json:"shareInformation,omitempty"`
	ApprovalLevel             int       `json:"approvalLevel"`
	TxType                 string   `json:"txType" final:"threshold"`
}

type ECDSAThresholdSignatureTransaction struct {
	SignerID                  string    `json:"signerID"`
	Status                    string    `json:"status"`
	Message                   string    `json:"message"`
	MessageHash               string    `json:"messageHash" id:"true" mandatory:"true"`
	TokenId         		  string    `json:"tokenId"`
	TransactionDate			  string    `json:"transactionDate"`
	Signature			      string    `json:"signature"`
	PubKey			          string	`json:"pubKey,omitempty"`
	ApprovalLevel             int       `json:"approvalLevel"`
	TxType                    string   `json:"txType" final:"ecdsathreshold"`
	TxReceipt                 string   `json:"txReceipt,omitempty"`
}


type ShareInformation struct {
	SignedShare 	   ted25519.PartialSignature    `json:"signedShare,omitempty"`
	NonceShare		   []string   					`json:"nonceShare"`	  
	NoncePubShare	   string   				`json:"noncePubShare"`	 
}

type ThresholdFragment struct {
	SecretShare    string   `json:"secretShare"`
	NonceShare     string   `json:"nonceShare"`
}

type SharedWallet struct {
	SecretShare    string   `json:"secretShare"`
	PubKey     string   `json:"pubKey"`
}

type SignedShare struct {
	ShareIdentifier string `json:"ShareIdentifier"`
	Sig				string `json:"Sig"`
}

type SignatureReciever struct {
	model       *model.Model
	transaction *transaction.TransactionReciever
	account     *account.AccountReciever
}

type SigningNotification struct {
	Org  		string `json:"org"`
	AssetKey	string `json:"assetKey"`
	Asset       []byte `json:"asset"`
}

func GetNewSignatureReciever(m *model.Model, trx *transaction.TransactionReciever, acc *account.AccountReciever) *SignatureReciever {
	var s SignatureReciever
	s.model = m
	s.transaction = trx
	s.account = acc
	return &s
}


func (s *SignatureReciever) get(Id string) (SignatureTransaction, error) {
	stub := s.model.GetNetworkStub()

	txAsBytes, err := stub.GetState(Id)
	if err != nil {
		return SignatureTransaction{}, fmt.Errorf("error in getting transaction with Id %s %s", Id, err.Error())
	}
	if txAsBytes == nil {
		return SignatureTransaction{}, fmt.Errorf("Transaction with Id %s does not exist", Id)
	}

	var tx SignatureTransaction
	unmarshalError := json.Unmarshal(txAsBytes, &tx)
	if unmarshalError != nil {
		return SignatureTransaction{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
	}
	
	return tx, nil
}

func (s *SignatureReciever) QueryTransaction(transactionId string) (SignatureTransaction, error) {

	if transactionId == "" {
		return SignatureTransaction{}, fmt.Errorf("error in retrieving transaction, transaction id is empty")
	}

	tx, err := s.get(transactionId)
	if err != nil {
		return SignatureTransaction{}, fmt.Errorf("error in getting transaction %s", err.Error())
	}

	return tx, nil
}

func (s *SignatureReciever) QueryExternalTransaction(transactionId string) (ExternalTransaction, error) {

	if transactionId == "" {
		return ExternalTransaction{}, fmt.Errorf("error in retrieving transaction, transaction id is empty")
	}

	stub := s.model.GetNetworkStub()

    txAsBytes, err := stub.GetState(transactionId)
    if err != nil {
        return ExternalTransaction{}, fmt.Errorf("error in getting transaction with Id %s %s", transactionId, err.Error())
    }
    if txAsBytes == nil {
        return ExternalTransaction{}, fmt.Errorf("Transaction with Id %s does not exist", transactionId)
    }

    var tx ExternalTransaction
    unmarshalError := json.Unmarshal(txAsBytes, &tx)
    if unmarshalError != nil {
        return ExternalTransaction{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
    }

    return tx, nil
}

func (s *SignatureReciever) QueryThresholdTransaction(transactionId string) (ThresholdSignatureTransaction, error) {

	if transactionId == "" {
		return ThresholdSignatureTransaction{}, fmt.Errorf("error in retrieving transaction, transaction id is empty")
	}

	stub := s.model.GetNetworkStub()

    txAsBytes, err := stub.GetState(transactionId)
    if err != nil {
        return ThresholdSignatureTransaction{}, fmt.Errorf("error in getting transaction with Id %s %s", transactionId, err.Error())
    }
    if txAsBytes == nil {
        return ThresholdSignatureTransaction{}, fmt.Errorf("Transaction with Id %s does not exist", transactionId)
    }

    var tx ThresholdSignatureTransaction
    unmarshalError := json.Unmarshal(txAsBytes, &tx)
    if unmarshalError != nil {
        return ThresholdSignatureTransaction{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
    }

    return tx, nil
}


func (s *SignatureReciever) PostTransaction(transaction string) (SignatureTransaction, error) {

	if transaction == "" {
		return SignatureTransaction{}, fmt.Errorf("tx cannot be empty")
	}

	var tx SignatureTransaction

	txBytes := []byte(transaction)

	err := json.Unmarshal(txBytes, &tx)
	if err != nil {
		return SignatureTransaction{}, fmt.Errorf("Error decoding TX JSON: %s", err.Error())
	}

	
	//Check that Transaction doesnt exist on ledger already by checking hash
	existingTx , err := s.get(tx.MessageHash)
	if existingTx.MessageHash != "" {
		return SignatureTransaction{}, fmt.Errorf("Transaction Hash already exist on ledger")
	}

	_, err = s.model.Save(&tx)
	if err != nil {
		return SignatureTransaction{}, fmt.Errorf("error in saving tx with Hash %s (SignerId: %s, Transaction Date: %s), Error: %s", tx.MessageHash , tx.SignerID, tx.TransactionDate, err.Error())
	}
	return tx, nil
}

func (s *SignatureReciever) PostSignature(txHash string, signedMsg string, signedDate string, keyAlias string, account_id string) (SignatureTransaction, error) {

	if txHash == "" {
		return SignatureTransaction{}, fmt.Errorf("Transaction Hash cannot be empty")
	}

	//Check that Transaction doesnt exist on ledger already by checking hash
	existingTx , err := s.get(txHash)
	if err != nil {
		return SignatureTransaction{}, fmt.Errorf("Transaction Hash not found ledger %s", err.Error())
	}
	
	isSignatureValid, existingTx, err := s.VerifyTransaction(existingTx, signedMsg, signedDate, keyAlias, account_id)
	if err != nil {
		return SignatureTransaction{}, fmt.Errorf("Signature Verification Failed %s", err.Error())
	}

	if isSignatureValid {
		existingTx.Status = "Verified"	
	} else {
		return SignatureTransaction{}, fmt.Errorf("invalid signature provided")	
	}

	_, err = s.model.Update(&existingTx)
	if err != nil {
		return SignatureTransaction{}, fmt.Errorf("error in updating tx with Hash %s (SignerId: %s, Transaction Date: %s), Error: %s", existingTx.MessageHash , existingTx.SignerID, existingTx.TransactionDate, err.Error())
	}
	return existingTx, nil
}


func (s *SignatureReciever) PrepareNextTransaction(transaction SignatureTransaction) (SignatureTransaction, error) {

	transAsBytes, _ := json.Marshal(transaction)
	message := b64.StdEncoding.EncodeToString(transAsBytes)

	hashedMessage := sha256.Sum256([]byte(message))
	var newTx SignatureTransaction
	newTx.MessageHash = hex.EncodeToString((hashedMessage[:]))
	newTx.Message = message

	return newTx, nil
}

func (s *SignatureReciever) PostExternalTransaction(transaction string) (ExternalTransaction, error) {

	if transaction == "" {
		return ExternalTransaction{}, fmt.Errorf("tx cannot be empty")
	}

	var newExternalTx ExternalTransaction

	txBytes := []byte(transaction)

	err := json.Unmarshal(txBytes, &newExternalTx)
	if err != nil {
		return ExternalTransaction{}, fmt.Errorf("Error decoding TX JSON: %s", err.Error())
	}

	//Check that Transaction doesnt exist on ledger already by checking hash
	existingTx , err := s.get(newExternalTx.Identifier)
	if existingTx.MessageHash != "" {
		return ExternalTransaction{}, fmt.Errorf("Transaction Hash already exist on ledger")
	}

	assetAsBytes, errMarshal := json.Marshal(&newExternalTx)
    if errMarshal != nil {
        return ExternalTransaction{}, fmt.Errorf("error in saving: Asset Id %s marshal error %s", newExternalTx.Identifier,  errMarshal.Error())
    }

	stub := s.model.GetNetworkStub()
    errPut := stub.PutState(newExternalTx.Identifier, assetAsBytes)
    if errPut != nil {
        return ExternalTransaction{}, fmt.Errorf("error in saving: Asset Id %s transaction error %s", newExternalTx.Identifier, errPut.Error())
    }

	return newExternalTx, nil
}

func (s *SignatureReciever) PostExternalTransactionSignature(transactionId string, signedMsg string, signedDate string, keyAlias string, account_id string) (ExternalTransaction, error) {
	if transactionId == "" {
		return ExternalTransaction{}, fmt.Errorf("error in retrieving transaction, transaction id is empty")
	}

	stub := s.model.GetNetworkStub()
	txAsBytes, err := stub.GetState(transactionId)
	if err != nil {
		return ExternalTransaction{}, fmt.Errorf("error in getting transaction with Id %s %s", transactionId, err.Error())
	}
	if txAsBytes == nil {
		return ExternalTransaction{}, fmt.Errorf("Transaction with Id %s does not exist", transactionId)
	}

	var externalOrgtx ExternalTransaction
	unmarshalError := json.Unmarshal(txAsBytes, &externalOrgtx)
	if unmarshalError != nil {
		return ExternalTransaction{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
	}


	switch status := externalOrgtx.Status; status {
	case "Owner":
		nextTx, existingTx , err := s.AdvanceSupport(externalOrgtx.AccountOwner, signedMsg, signedDate, keyAlias, account_id)
		if err != nil {
			return ExternalTransaction{}, fmt.Errorf("Error at Owner: %s", err.Error())
		}
		externalOrgtx.AccountOwner = existingTx
		externalOrgtx.Status = "Admin"
		externalOrgtx.OrgAdmin = nextTx
		externalOrgtx.OrgAdmin.SignerID = "admin"
	case "Admin":
		nextTx, existingTx ,err := s.AdvanceSupport(externalOrgtx.OrgAdmin, signedMsg, signedDate, keyAlias, account_id)
		if err != nil {
			return ExternalTransaction{}, fmt.Errorf("Error at Admin: %s", err.Error())
		}
		externalOrgtx.OrgAdmin = existingTx
		externalOrgtx.Status = "Signatory"
		externalOrgtx.OrgSignatory = nextTx
		externalOrgtx.OrgSignatory.SignerID = "signatory"
	case "Signatory":
		_, existingTx , err := s.AdvanceSupport(externalOrgtx.OrgSignatory, signedMsg, signedDate, keyAlias, account_id)
		if err != nil {
			return ExternalTransaction{}, fmt.Errorf("Error at Signatory: %s", err.Error())
		}
		externalOrgtx.OrgSignatory = existingTx
		externalOrgtx.Status = "Verified"
	default:
		return ExternalTransaction{}, fmt.Errorf("Incorrect Status For: %v", externalOrgtx)
	}

    assetAsBytes, errMarshal := json.Marshal(&externalOrgtx)
    if errMarshal != nil {
        return ExternalTransaction{}, fmt.Errorf("error in saving: Asset Id %s marshal error %s", externalOrgtx.Identifier,  errMarshal.Error())
    }

    errPut := stub.PutState(externalOrgtx.Identifier, assetAsBytes)
    if errPut != nil {
        return ExternalTransaction{}, fmt.Errorf("error in saving: Asset Id %s transaction error %s", externalOrgtx.Identifier, errPut.Error())
    }
	return externalOrgtx, nil
}


func (s *SignatureReciever) VerifyTransaction(existingTx SignatureTransaction, signedMsg string, signedDate string, keyAlias string, account_id string) (bool, SignatureTransaction,  error) {
	existingTx.TransactionSubmittedTimeStamp = signedDate
	existingTx.KeyAlias = keyAlias

	stub := s.model.GetNetworkStub()

	assetAsBytes, err := stub.GetState(account_id)
	if err != nil {
		return false, existingTx , fmt.Errorf("error in getting Account with Id %s %s", account_id, err.Error())
	}
	if assetAsBytes == nil {
		return false, existingTx, fmt.Errorf("Account with Id %s does not exist", account_id)
	}

	var accountAsset account.Account
	unmarshalError := json.Unmarshal(assetAsBytes, &accountAsset)
	if unmarshalError != nil {
		return false, existingTx, fmt.Errorf("marshalling error %s", unmarshalError.Error())
	}

    ecParameters := accountAsset.PublicKeystore[keyAlias]
	existingTx.ECParameters = ecParameters
	existingTx.SignedMessage = signedMsg
	existingTx.Status = "Verification Pending"
	// call verify function

	curveName := ecParameters.CurveName
	curveType := ecParameters.CurveType
	px := new(big.Int)
    px, _ = px.SetString(ecParameters.PX, 10)
	py := new(big.Int)
    py, _ = py.SetString(ecParameters.PY, 10)

	signedMessage, _ := b64.StdEncoding.DecodeString(existingTx.SignedMessage)
	message, _ := b64.StdEncoding.DecodeString(existingTx.Message)
	hashedMessage := sha256.Sum256(message)
	isSignatureValid , _ := ccsp.VerifySignature(curveName,curveType,signedMessage,hashedMessage[:],px.Bytes(),py.Bytes())
	
	return isSignatureValid, existingTx, nil
}


func (s *SignatureReciever) AdvanceSupport(existingTx SignatureTransaction, signedMsg string, signedDate string, keyAlias string, account_id string) (SignatureTransaction, SignatureTransaction, error) {
	isSignatureValid, existingTx,  err := s.VerifyTransaction(existingTx, signedMsg, signedDate, keyAlias, account_id)
	if err != nil {
		return SignatureTransaction{}, existingTx, fmt.Errorf("Signature Verification Failed %s", err.Error())
	}

	if isSignatureValid {
		existingTx.Status = "Verified"	
	} else {

		return SignatureTransaction{}, existingTx, fmt.Errorf("invalid signature provided for %s, the original mesasge %s", signedMsg, existingTx.Message)
	}

	nextTx, err := s.PrepareNextTransaction(existingTx)
	if err != nil {
		return SignatureTransaction{}, existingTx ,fmt.Errorf("error in generating next tx, Error: %s", err.Error())
	}

	return nextTx, existingTx, nil
}

func (s *SignatureReciever) GenerateWalletSharedKey() (string, error) {
	config := ted25519.ShareConfiguration{T: 5, N: 5}
	pub, secretShares, _, _ := ted25519.GenerateSharedKey(&config)
	ss1 :=  b64.StdEncoding.EncodeToString(secretShares[0].Bytes())
	ss2 :=  b64.StdEncoding.EncodeToString(secretShares[1].Bytes())
	ss3 :=  b64.StdEncoding.EncodeToString(secretShares[2].Bytes())
	ss4 :=  b64.StdEncoding.EncodeToString(secretShares[3].Bytes())
	ss5 :=  b64.StdEncoding.EncodeToString(secretShares[4].Bytes())
	pubj := b64.StdEncoding.EncodeToString(pub.Bytes())

	keyshards := strings.Join([]string{ss1, ss2, ss3, ss4, ss5, pubj},",")

	return keyshards ,nil
}

func (s *SignatureReciever) PostDealerKeyFragments() (string, error) {

		stub := s.model.GetNetworkStub()

		// Get new asset from transient map
		transientMap, err := stub.GetTransient()
		if err != nil {
			return "", fmt.Errorf("error getting transient: %v", err)
		}

		// Asset properties are private, therefore they get passed in transient field, instead of func args
		transientAsset, ok := transientMap["asset_properties"]
		if !ok {
			//log error to stdout
			return "", fmt.Errorf("asset not found in the transient map input")
		}

		keyFragmentInfo := string(transientAsset)
			
		keyFragArgs := strings.Split(keyFragmentInfo, ",")
		pubKey := keyFragArgs[len(keyFragArgs)-1] //bpub
	
	
		fr1 := SharedWallet{keyFragArgs[0], pubKey}
		bfr1, _ := json.Marshal(fr1) 
		err = stub.PutPrivateData("_implicit_org_Org1MSP", "SharedWallet", bfr1)
		if err != nil {
			return "", fmt.Errorf("Failed to set asset: %s", "secretshare1")
		}
	
		fr2 := SharedWallet{keyFragArgs[1], pubKey}
		bfr2, _ := json.Marshal(fr2) 
		err = stub.PutPrivateData("_implicit_org_Org2MSP", "SharedWallet", bfr2)
		if err != nil {
			return "", fmt.Errorf("Failed to set asset: %s", "secretshare2")
		}
	
		fr3 := SharedWallet{keyFragArgs[2], pubKey}
		bfr3, _ := json.Marshal(fr3) 
		err = stub.PutPrivateData("_implicit_org_Org3MSP", "SharedWallet", bfr3)
		if err != nil {
			return "", fmt.Errorf("Failed to set asset: %s", "secretshare1")
		}
		
		fr4 := SharedWallet{keyFragArgs[3], pubKey}
		bfr4, _ := json.Marshal(fr4) 
		err = stub.PutPrivateData("_implicit_org_Org4MSP", "SharedWallet", bfr4)
		if err != nil {
			return "", fmt.Errorf("Failed to set asset: %s", "secretshare2")
		}
	
		fr5 := SharedWallet{keyFragArgs[4], pubKey}
		bfr5, _ := json.Marshal(fr5) 
		err = stub.PutPrivateData("_implicit_org_Org5MSP", "SharedWallet", bfr5)
		if err != nil {
			return "", fmt.Errorf("Failed to set asset: %s", "secretshare1")
		}

		errPut := stub.PutState("SharedWalletPublicKey", []byte(pubKey))
		if errPut != nil {
			return "", fmt.Errorf("error in saving public key: %s", errPut.Error())
		}
	
		return "Success", nil
}

func (s *SignatureReciever) GenerateSharedNonce(msg string, orgId string) ([]string, error) {
		message := []byte(msg)

		stub := s.model.GetNetworkStub()

		value, err := stub.GetPrivateData("_implicit_org_"+orgId, "SharedWallet")
		if err != nil {
			return []string{}, fmt.Errorf("Failed to get shared wallet fragment: with error: %s", err)
		}
		if value == nil {
			return []string{}, fmt.Errorf("Shared wallet not found")
		}

		var fragment SharedWallet
		err = json.Unmarshal(value, &fragment)
		if err != nil {
			return []string{}, fmt.Errorf("Error decoding share information for org %s", orgId)
		}

		pubBytes, _ := b64.StdEncoding.DecodeString(fragment.PubKey)	

		pub, err := ted25519.PublicKeyFromBytes(pubBytes)
		if err != nil {
			fmt.Errorf("Failed to get pub key from bytes: %s", err)
		}

		ssBytes, err := b64.StdEncoding.DecodeString(fragment.SecretShare)
		secretShare := ted25519.KeyShareFromBytes(ssBytes)

	
		config := ted25519.ShareConfiguration{T: 5, N: 5}
	
		// Each party generates a nonce and we combine them together into an aggregate one
		noncePub, nonceShare, _, _ := ted25519.GenerateSharedNonce(&config, secretShare, pub, message)

		var nsArray []string
		nsArray = []string{
			b64.StdEncoding.EncodeToString(nonceShare[0].Bytes()),
			b64.StdEncoding.EncodeToString(nonceShare[1].Bytes()),
			b64.StdEncoding.EncodeToString(nonceShare[2].Bytes()),
			b64.StdEncoding.EncodeToString(nonceShare[3].Bytes()),
			b64.StdEncoding.EncodeToString(nonceShare[4].Bytes()),
		}

		nsFinal := strings.Join(nsArray,",")
		
		noncepubj :=  b64.StdEncoding.EncodeToString(noncePub.Bytes())
	
		return []string{nsFinal, noncepubj},nil	
}	

func (s *SignatureReciever) PostInitialThresholdTransaction(transaction string) (ThresholdSignatureTransaction, error) {

		//func (s *SignatureReciever) PostThresholdTransaction(transaction string, bss1 string, bss2 string, bss3 string, bss4 string, bss5 string, bns1 string, bns2 string, bns3 string, bns4 string, bns5 string, bpub string, bnoncepub string) (ThresholdSignatureTransaction, error) {
			stub := s.model.GetNetworkStub()
			if transaction == "" {
				return ThresholdSignatureTransaction{}, fmt.Errorf("tx cannot be empty")
			}
		
			var tx ThresholdSignatureTransaction
		
			txBytes := []byte(transaction)
		
			err := json.Unmarshal(txBytes, &tx)
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("Error decoding TX JSON: %s", err.Error())
			}
					
			//Check that Transaction doesnt exist on ledger already by checking hash
			existingTx , err := s.get(tx.MessageHash)
			if existingTx.MessageHash != "" {
				return ThresholdSignatureTransaction{}, fmt.Errorf("Transaction Hash already exist on ledger")
			}

			pkBytes, err := stub.GetState("SharedWalletPublicKey")
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("error in getting public key for singlw wallet %s",  err.Error())
			}
			
			tx.PubKey = string(pkBytes)
			tx.Status = "Pending"

					
			_, err = s.model.Save(&tx)
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("error in saving tx with Hash %s (SignerId: %s, Transaction Date: %s), Error: %s", tx.MessageHash , tx.SignerID, tx.TransactionDate, err.Error())
			}					
		
			txBytes, _ = json.Marshal(tx)
			var event = SigningNotification{"AllOrgs", tx.MessageHash, txBytes}
			eventAsBytes, _ := json.Marshal(event)
			err = stub.SetEvent("PostInitialThresholdTransaction", eventAsBytes)
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("Failed to create event : %s", err)
			}
		
			return tx, nil
		}	



func (s *SignatureReciever) PostTransactionSharedNonce(txHash string, orgId string, nonceShare string, noncePubShare string) (ThresholdSignatureTransaction, error) {

	stub := s.model.GetNetworkStub()

	var tx ThresholdSignatureTransaction

	txAsBytes, err := stub.GetState(txHash)
	if err != nil {
		return ThresholdSignatureTransaction{}, fmt.Errorf("error in getting transaction with Id %s %s", txHash, err.Error())
	}
	if txAsBytes == nil {
		return ThresholdSignatureTransaction{}, fmt.Errorf("Transaction with Id %s does not exist", txHash)
	}

	err = json.Unmarshal(txAsBytes, &tx)
	if err != nil {
		return ThresholdSignatureTransaction{}, fmt.Errorf("Error decoding TX JSON: %s", err.Error())
	}

	nonceShareArray := strings.Split(nonceShare, ",")
	newInfoAdded := false

	if len(tx.ShareInformation) == 0 {
        var shareInformation = make(map[string]ShareInformation)
        shareInformation[orgId] = ShareInformation{NonceShare: nonceShareArray, NoncePubShare: noncePubShare}
        tx.ShareInformation = shareInformation
		newInfoAdded = true
    } else {
        if _, ok := tx.ShareInformation[orgId]; !ok {
			tx.ShareInformation[orgId] = ShareInformation{NonceShare: nonceShareArray, NoncePubShare: noncePubShare}
			newInfoAdded = true
        } 
    }	

	txBytes, _ := json.Marshal(tx)
	var event = SigningNotification{"AllOrgs", tx.MessageHash, txBytes}
	eventAsBytes, _ := json.Marshal(event)


	if (newInfoAdded){
		if len(tx.ShareInformation) < 5 {
			_, err = s.model.Update(&tx)
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("error in saving tx with Hash %s (SignerId: %s, Transaction Date: %s), Error: %s", tx.MessageHash , tx.SignerID, tx.TransactionDate, err.Error())
			}
	
			err = stub.SetEvent("PostTransactionNonceShare", eventAsBytes)
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("Failed to create event : %s", err)
			}
		} else {
			var noncePubArray []ted25519.PublicKey
			for _ , val := range tx.ShareInformation {
				nsBytes, err := b64.StdEncoding.DecodeString(val.NoncePubShare)
				noncePub, err := ted25519.PublicKeyFromBytes(nsBytes)
				if err != nil {
					fmt.Errorf("Failed to get pub key from bytes: %s", err)
				}
				noncePubArray = append(noncePubArray, noncePub)
			}
			noncePub := ted25519.GeAdd(ted25519.GeAdd(ted25519.GeAdd(ted25519.GeAdd(noncePubArray[0], noncePubArray[1]), noncePubArray[2]), noncePubArray[3]), noncePubArray[4])
	
			tx.NoncePub	 = b64.StdEncoding.EncodeToString(noncePub.Bytes())
	
			_, err = s.model.Update(&tx)
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("error in saving tx with Hash %s (SignerId: %s, Transaction Date: %s), Error: %s", tx.MessageHash , tx.SignerID, tx.TransactionDate, err.Error())
			}
	
			err = stub.SetEvent("RequestPartialTransactionSignature", eventAsBytes)
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("Failed to create event : %s", err)
			}
		}
	}
	
	
	return tx, nil
}		

func (s *SignatureReciever) TSignOrgSharedWallet(key string, orgId string) (ted25519.PartialSignature, error) {
	stub := s.model.GetNetworkStub()

	value, err := stub.GetPrivateData("_implicit_org_"+orgId, "SharedWallet")
	if err != nil {
		return ted25519.PartialSignature{}, fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
	}
	if value == nil {
		return ted25519.PartialSignature{}, fmt.Errorf("Asset not found: %s", key)
	}

	var fragment SharedWallet
	err = json.Unmarshal(value, &fragment)
	if err != nil {
		return ted25519.PartialSignature{}, fmt.Errorf("Error decoding share information for org %s and key: %s", orgId, key)
	}

	ssBytes, err := b64.StdEncoding.DecodeString(fragment.SecretShare)
	secretShare := ted25519.KeyShareFromBytes(ssBytes)

	pubBytes, _ := b64.StdEncoding.DecodeString(fragment.PubKey)	
	pub, err := ted25519.PublicKeyFromBytes(pubBytes)
	if err != nil {
		fmt.Errorf("Failed to get pub key from bytes: %s", err)
	}

	txAsBytes, err := stub.GetState(key)
	if err != nil {
		return ted25519.PartialSignature{}, fmt.Errorf("error in getting transaction with Id %s %s", key, err.Error())
	}
	if txAsBytes == nil {
		return ted25519.PartialSignature{}, fmt.Errorf("Transaction with Id %s does not exist", key)
	}

	var tx ThresholdSignatureTransaction
	unmarshalError := json.Unmarshal(txAsBytes, &tx)
	if unmarshalError != nil {
		return ted25519.PartialSignature{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
	}


	noncepubBytes, _ := b64.StdEncoding.DecodeString(tx.NoncePub)
	noncePub, err := ted25519.PublicKeyFromBytes(noncepubBytes)
	if err != nil {
		return ted25519.PartialSignature{}, fmt.Errorf("Failed to get pub key from bytes: %s", err)
	}

	var org1tmpNonceShareArray []*ted25519.NonceShare
	org1 := tx.ShareInformation["Org1MSP"]
	for _ , val1 := range org1.NonceShare {
		nsBytes, _ := b64.StdEncoding.DecodeString(val1)
		nonceSharesData := ted25519.NonceShareFromBytes(nsBytes)
		org1tmpNonceShareArray = append(org1tmpNonceShareArray, nonceSharesData)
	}

	var org2tmpNonceShareArray []*ted25519.NonceShare
	org2 := tx.ShareInformation["Org2MSP"]
	for _ , val2 := range org2.NonceShare {
		nsBytes, _ := b64.StdEncoding.DecodeString(val2)
		nonceSharesData := ted25519.NonceShareFromBytes(nsBytes)
		org2tmpNonceShareArray = append(org2tmpNonceShareArray, nonceSharesData)
	}

	var org3tmpNonceShareArray []*ted25519.NonceShare
	org3 := tx.ShareInformation["Org3MSP"]
	for _ , val3 := range org3.NonceShare {
		nsBytes, _ := b64.StdEncoding.DecodeString(val3)
		nonceSharesData := ted25519.NonceShareFromBytes(nsBytes)
		org3tmpNonceShareArray = append(org3tmpNonceShareArray, nonceSharesData)
	}

	var org4tmpNonceShareArray []*ted25519.NonceShare
	org4 := tx.ShareInformation["Org4MSP"]
	for _ , val4 := range org4.NonceShare {
		nsBytes, _ := b64.StdEncoding.DecodeString(val4)
		nonceSharesData := ted25519.NonceShareFromBytes(nsBytes)
		org4tmpNonceShareArray = append(org4tmpNonceShareArray, nonceSharesData)
	}

	var org5tmpNonceShareArray []*ted25519.NonceShare
	org5:= tx.ShareInformation["Org5MSP"]
	for _ , val5 := range org5.NonceShare {
		nsBytes, _ := b64.StdEncoding.DecodeString(val5)
		nonceSharesData := ted25519.NonceShareFromBytes(nsBytes)
		org5tmpNonceShareArray = append(org5tmpNonceShareArray, nonceSharesData)
	}

	share := 0
	switch orgId {
    case "Org1MSP":
        share = 0
    case "Org2MSP":
        share = 1
    case "Org3MSP":
        share = 2
	case "Org4MSP":
		share = 3
	case "Org5MSP":
		share = 4
	}

	nonceShares := []*ted25519.NonceShare{
		org1tmpNonceShareArray[0].Add(org2tmpNonceShareArray[0]).Add(org3tmpNonceShareArray[0]).Add(org4tmpNonceShareArray[0]).Add(org5tmpNonceShareArray[0]),
		org1tmpNonceShareArray[1].Add(org2tmpNonceShareArray[1]).Add(org3tmpNonceShareArray[1]).Add(org4tmpNonceShareArray[1]).Add(org5tmpNonceShareArray[1]),
		org1tmpNonceShareArray[2].Add(org2tmpNonceShareArray[2]).Add(org3tmpNonceShareArray[2]).Add(org4tmpNonceShareArray[2]).Add(org5tmpNonceShareArray[2]),
		org1tmpNonceShareArray[3].Add(org2tmpNonceShareArray[3]).Add(org3tmpNonceShareArray[3]).Add(org4tmpNonceShareArray[3]).Add(org5tmpNonceShareArray[3]),
		org1tmpNonceShareArray[4].Add(org2tmpNonceShareArray[4]).Add(org3tmpNonceShareArray[4]).Add(org4tmpNonceShareArray[4]).Add(org5tmpNonceShareArray[4]),
	}


	sig := ted25519.TSign([]byte(tx.Message), secretShare, pub, nonceShares[share], noncePub)
	return *sig, nil
}

func (s *SignatureReciever) PostSignShareWallet(key string, sigShareJSON string, orgId string) (ThresholdSignatureTransaction, error) {
	stub := s.model.GetNetworkStub()

	var tx ThresholdSignatureTransaction

	txAsBytes, err := stub.GetState(key)
	if err != nil {
		return ThresholdSignatureTransaction{}, fmt.Errorf("error in getting transaction with Id %s %s", key, err.Error())
	}
	if txAsBytes == nil {
		return ThresholdSignatureTransaction{}, fmt.Errorf("Transaction with Id %s does not exist", key)
	}

	err = json.Unmarshal(txAsBytes, &tx)
	if err != nil {
		return ThresholdSignatureTransaction{}, fmt.Errorf("Error decoding TX JSON: %s", err.Error())
	}

	var sigShare ted25519.PartialSignature
	err = json.Unmarshal([]byte(sigShareJSON), &sigShare)
	if err != nil {
		return ThresholdSignatureTransaction{}, fmt.Errorf("Error decoding json sig share information for key: %s", key)
	}

	pubBytes, _ := b64.StdEncoding.DecodeString(tx.PubKey)

	pub, err := ted25519.PublicKeyFromBytes(pubBytes)
	if err != nil {
		fmt.Errorf("Failed to get pub key from bytes: %s", err)
	}

	sharedInformation := tx.ShareInformation[orgId]
	sharedInformation.SignedShare = sigShare

	tx.ShareInformation[orgId] = sharedInformation

	shares := 0
	var sigShares []ted25519.PartialSignature
	for _ , val := range tx.ShareInformation {
		if (val.SignedShare.Sig != nil) {
			sigShares = append(sigShares, val.SignedShare)
			shares++
		}
	}

	if (shares == 5) {	
		config := ted25519.ShareConfiguration{T: 5, N: 5}
		var sigShares2 []*ted25519.PartialSignature
		for i := range sigShares {
			sigShares2 = append(sigShares2, &sigShares[i])
		}
		
		sig, _ := ted25519.Aggregate(sigShares2, &config)
		ok, _ := ted25519.Verify(pub, []byte(tx.Message), sig)
		if ok {
			tx.Status = "Verified"
		} else {
			tx.Status = "Not Verified"
			
		}
	
		_, err = s.model.Update(&tx)
		if err != nil {
			return ThresholdSignatureTransaction{}, fmt.Errorf("error in updating tx with Hash %s (SignerId: %s), Error: %s", tx.MessageHash , tx.SignerID, err.Error())
		}

		if (tx.Status == "Not Verified") {
			txBytes, _ := json.Marshal(tx)
			var event = SigningNotification{"AllOrgs", tx.MessageHash, txBytes}
			eventAsBytes, _ := json.Marshal(event)
			err = stub.SetEvent("ThresholdSigningFailed", eventAsBytes)
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("Failed to create event : %s", err)
			}
		} else if (tx.Status == "Verified") {
			txBytes, _ := json.Marshal(tx)
			var event = SigningNotification{"AllOrgs", tx.MessageHash, txBytes}
			eventAsBytes, _ := json.Marshal(event)
			err = stub.SetEvent("ThresholdSigningComplete", eventAsBytes)
			if err != nil {
				return ThresholdSignatureTransaction{}, fmt.Errorf("Failed to create event : %s", err)
			}
		}
	} else {
		_, err = s.model.Update(&tx)
		if err != nil {
			return ThresholdSignatureTransaction{}, fmt.Errorf("error in updating tx with Hash %s (SignerId: %s), Error: %s", tx.MessageHash , tx.SignerID, err.Error())
		}

		txBytes, _ := json.Marshal(tx)
		var event = SigningNotification{"AllOrgs", tx.MessageHash, txBytes}
		eventAsBytes, _ := json.Marshal(event)
		err = stub.SetEvent("CompletePostThresholdTransactionSharedWallet", eventAsBytes)
		if err != nil {
			return ThresholdSignatureTransaction{}, fmt.Errorf("Failed to create event : %s", err)
		}
	}
	
	return tx, nil

}


func (s *SignatureReciever) GetPrivateOrg(key string, orgId string) (string, error) {
	stub := s.model.GetNetworkStub()

	value, err := stub.GetPrivateData("_implicit_org_" + orgId, key)
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", key)
	}
	return string(value), nil
}

func (s *SignatureReciever) GetAllPendingTxs() (interface{}, error) {
	query := `{"selector": {"status": "Pending", "txType": "threshold"}}`
	return s.model.Query(query)
}

func (s *SignatureReciever) GetAllThresholdTxs() (interface{}, error) {
	query := `{"selector": {"txType": "threshold"}}`
	return s.model.Query(query)
}


/////////////////////////////////
var (
	testPrimes = []*big.Int{
		B10("186141419611617071752010179586510154515933389116254425631491755419216243670159714804545944298892950871169229878325987039840135057969555324774918895952900547869933648175107076399993833724447909579697857041081987997463765989497319509683575289675966710007879762972723174353568113668226442698275449371212397561567"),
		B10("94210786053667323206442523040419729883258172350738703980637961803118626748668924192069593010365236618255120977661397310932923345291377692570649198560048403943687994859423283474169530971418656709749020402756179383990602363122039939937953514870699284906666247063852187255623958659551404494107714695311474384687"),
		B10("62028909880050184794454820320289487394141550306616974968340908736543032782344593292214952852576535830823991093496498970213686040280098908204236051130358424961175634703281821899530101130244725435470475135483879784963475148975313832483400747421265545413510460046067002322131902159892876739088034507063542087523"),
		B10("321804071508183671133831207712462079740282619152225438240259877528712344129467977098976100894625335474509551113902455258582802291330071887726188174124352664849954838358973904505681968878957681630941310372231688127901147200937955329324769631743029415035218057960201863908173045670622969475867077447909836936523"),
		B10("52495647838749571441531580865340679598533348873590977282663145916368795913408897399822291638579504238082829052094508345857857144973446573810004060341650816108578548997792700057865473467391946766537119012441105169305106247003867011741811274367120479722991749924616247396514197345075177297436299446651331187067"),
		B10("118753381771703394804894143450628876988609300829627946826004421079000316402854210786451078221445575185505001470635997217855372731401976507648597119694813440063429052266569380936671291883364036649087788968029662592370202444662489071262833666489940296758935970249316300642591963940296755031586580445184253416139"),
	}
	dealerParams = &dealer.ProofParams{
		N:  B10("135817986946410153263607521492868157288929876347703239389804036854326452848342067707805833332721355089496671444901101084429868705550525577068432132709786157994652561102559125256427177197007418406633665154772412807319781659630513167839812152507439439445572264448924538846645935065905728327076331348468251587961"),
		H1: B10("130372793360787914947629694846841279927281520987029701609177523587189885120190605946568222485341643012763305061268138793179515860485547361500345083617939280336315872961605437911597699438598556875524679018909165548046362772751058504008161659270331468227764192850055032058007664070200355866555886402826731196521"),
		H2: B10("44244046835929503435200723089247234648450309906417041731862368762294548874401406999952605461193318451278897748111402857920811242015075045913904246368542432908791195758912278843108225743582704689703680577207804641185952235173475863508072754204128218500376538767731592009803034641269409627751217232043111126391"),
	}
	k256Verifier = func(pubKey *curves.EcPoint, hash []byte, sig *curves.EcdsaSignature) bool {
		btcPk := &btcec.PublicKey{
			Curve: btcec.S256(),
			X:     pubKey.X,
			Y:     pubKey.Y,
		}
		btcSig := btcec.Signature{
			R: sig.R,
			S: sig.S,
		}
		return btcSig.Verify(hash, btcPk)
	}
)


func getParams(msg *string, t, n *uint32) {
	argCount := len(os.Args[1:])

	if argCount > 0 {
		*msg = os.Args[1]

	}
	if argCount > 1 {
		val, _ := strconv.Atoi(os.Args[2])
		*t = uint32(val)
	}
	if argCount > 2 {
		val, _ := strconv.Atoi(os.Args[3])
		*n = uint32(val)
	}

}
func genPrimesArray(count int) []struct{ p, q *big.Int } {
	primesArray := make([]struct{ p, q *big.Int }, 0, count)
	for len(primesArray) < count {
		for i := 0; i < len(testPrimes) && len(primesArray) < count; i++ {
			for j := 0; j < len(testPrimes) && len(primesArray) < count; j++ {
				if i == j {
					continue
				}
				keyPrime := struct {
					p, q *big.Int
				}{
					testPrimes[i], testPrimes[j],
				}
				primesArray = append(primesArray, keyPrime)
			}
		}
	}
	return primesArray
}

func B10(s string) *big.Int {
	x, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("Couldn't derive big.Int from string")
	}
	return x
}

type GenTECDSA struct {
	PK string `json:"pk"`
	Shares map[uint32]string `json:"shares"`
	PubShares map[uint32]string `json:"pubshares"`
	PaillierKeys map[uint32]string `json:"paillierKeys"`
	UserId   string `json:"userId"`
	TokenId  string `json:"tokenId"`
}

type ApprovalTECDSA struct {
	Approval bool `json:"approval"`
	OrgId    string `json:"orgId"`
	TxHash   string `json:"txHash"`
	Participant string `json:"participant,omitempty"`
	SK   	 string    `json:"sk,omitmepty"`
	UserId   string `json:"userId"`
	TokenId  string `json:"tokenId"`
}

type ECDSAParticipant struct {
	PK string `json:"pk"`
	Share string `json:"share"`
	PubShare string `json:"pubshare"`
	PaillierKey string `json:"paillierKey"`
}

type ECDSAPublicInfo struct {
	PK string `json:"pk"`
	PubShares map[uint32]string `json:"pubshares"`
	PubKeys map[uint32]string `json:"pubkeys"`
}

type ECDSAPublicInfoImport struct {
	PK string `json:"pk"`
	PubShares map[string]string `json:"pubshares"`
	PubKeys map[string]string `json:"pubkeys"`
}

func (s *SignatureReciever) GenerateECDSAWalletSharedKey(userId string, tokenId string) (GenTECDSA, error) {

	var newecdsa GenTECDSA

	newecdsa.UserId = userId
	newecdsa.TokenId = tokenId

	tshare := uint32(5)
	nshare := uint32(5)
	//msg := "Hello" 

	//m := []byte(msg)
	//msgHash := sha256.Sum256(m)

	//getParams(&msg, &tshare, &nshare)

	k256 := btcec.S256()

	ikm, _ := dealer.NewSecret(k256)

	pk, sharesMap, _ := dealer.NewDealerShares(k256, tshare, nshare, ikm)


	pkjson , _ := pk.MarshalJSON()
	newecdsa.PK = string(pkjson)

	var sharesstrings = make(map[uint32]string, tshare)
	for i := range sharesMap {
		jsonres, _ := sharesMap[i].MarshalJSON()
		sharesstrings[i] = string(jsonres)
	}

	newecdsa.Shares = sharesstrings

	pubSharesMap, _ := dealer.PreparePublicShares(sharesMap)

	var pubsharesstrings = make(map[uint32]string, tshare)

	for i := range pubSharesMap {
		jsonres, _ := json.Marshal(pubSharesMap[i])
		pubsharesstrings[i] = string(jsonres)
	}

	newecdsa.PubShares = pubsharesstrings

	//keyPrimesArray := genPrimesArray(2)

	keyPrimesArray := genPrimesArray(5)
	var paillierstring = make(map[uint32]string, tshare)
	for i := range sharesMap {
		paillerkey, _ := paillier.NewSecretKey(keyPrimesArray[i-1].p, keyPrimesArray[i-1].q)
		skjson , _ := paillerkey.MarshalJSON()
		paillierstring[i] = string(skjson)
	}

	newecdsa.PaillierKeys = paillierstring
	

	return newecdsa ,nil
}



func (s *SignatureReciever) PostECDSAWalletSharedKey() (ECDSAPublicInfo, error) {

	stub := s.model.GetNetworkStub()

	// Get new asset from transient map
	transientMap, err := stub.GetTransient()
	if err != nil {
		return ECDSAPublicInfo{}, fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		//log error to stdout
		return ECDSAPublicInfo{}, fmt.Errorf("asset not found in the transient map input")
	}

		
	var newecdsa GenTECDSA
	err = json.Unmarshal(transientAssetJSON, &newecdsa)
	if err != nil {
		return ECDSAPublicInfo{}, fmt.Errorf("marshalling error %s", string(transientAssetJSON))
	}

	var pubKeyMap = make(map[uint32]string, 5)
	for i:=1 ; i <= len(newecdsa.Shares) ; i++ {
		participant := ECDSAParticipant{newecdsa.PK, newecdsa.Shares[uint32(i)], newecdsa.PubShares[uint32(i)], newecdsa.PaillierKeys[uint32(i)]}
		participantBytes, _ := json.Marshal(participant) 
		err := stub.PutPrivateData("_implicit_org_Org" + strconv.Itoa(i) + "MSP", "ECDSASharedWallet-" + newecdsa.UserId + "-" + newecdsa.TokenId, participantBytes)
		if err != nil {
			return ECDSAPublicInfo{}, fmt.Errorf("Failed to set asset secret share for %d", i)
		}

		
		var paillier paillier.SecretKey
		paillier.UnmarshalJSON([]byte(newecdsa.PaillierKeys[uint32(i)]))

		pskjson2 , err := paillier.PublicKey.MarshalJSON()
		if err != nil {
			return ECDSAPublicInfo{}, fmt.Errorf("Failed to get decode paillier key")
		}
		pubKeyMap[uint32(i)] = string(pskjson2)

	}



	publicInfo := ECDSAPublicInfo{newecdsa.PK, newecdsa.PubShares, pubKeyMap}
	publicInfoBytes, _ := json.Marshal(publicInfo)

	errPut := stub.PutState("SharedWalletECDSAPublicKey-" + newecdsa.UserId + "-" + newecdsa.TokenId, publicInfoBytes)
	if errPut != nil {
		return ECDSAPublicInfo{}, fmt.Errorf("error in saving public key: %s", errPut.Error())
	}

	return publicInfo, nil
}

func (s *SignatureReciever) PostInitialECDSAThresholdTransaction(transaction string) (ECDSAThresholdSignatureTransaction, error) {

	stub := s.model.GetNetworkStub()
	if transaction == "" {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("tx cannot be empty")
	}

	var tx ECDSAThresholdSignatureTransaction

	txBytes := []byte(transaction)

	err := json.Unmarshal(txBytes, &tx)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("Error decoding TX JSON: %s", err.Error())
	}
			
	//Check that Transaction doesnt exist on ledger already by checking hash
	existingTx , err := s.get(tx.MessageHash)
	if existingTx.MessageHash != "" {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("Transaction Hash already exist on ledger")
	}

	pkInfoAsBytes, err := stub.GetState("SharedWalletECDSAPublicKey-" + tx.SignerID + "-" + tx.TokenId)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error in getting pkInfo with Id %s %s", "SharedWalletECDSAPublicKey-" + tx.SignerID + "-" + tx.TokenId, err.Error())
	}
	if pkInfoAsBytes == nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("pkInfo does not exist")
	}

	var pkInfo ECDSAPublicInfoImport
	err = json.Unmarshal(pkInfoAsBytes, &pkInfo)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("Error decoding pk Info: %s",err)
	}
	
	tx.PubKey = pkInfo.PK
	tx.Status = "Pending"

			
	_, err = s.model.Save(&tx)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error in saving tx with Hash %s (SignerId: %s, Transaction Date: %s), Error: %s", tx.MessageHash , tx.SignerID, tx.TransactionDate, err.Error())
	}


	txBytes, err = json.Marshal(tx)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("Error encoding TX JSON: %s", err.Error())
	}

	txBytes, _ = json.Marshal(tx)
	var event = SigningNotification{"AllOrgs", tx.MessageHash, txBytes}
	eventAsBytes, _ := json.Marshal(event)
	err = stub.SetEvent("PostInitialECDSAThresholdTransaction", eventAsBytes)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("Failed to create event : %s", err)
	}

	return tx, nil
}


func (s *SignatureReciever) PrepareApproveECDSATx() (string, error) {
	var result string
	stub := s.model.GetNetworkStub()

	// Get new asset from transient map
	transientMap, err := stub.GetTransient()
	if err != nil {
		return result, fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		//log error to stdout
		return result, fmt.Errorf("asset not found in the transient map input")
	}
		
	var approvalEcdsa ApprovalTECDSA
	err = json.Unmarshal(transientAssetJSON, &approvalEcdsa)
	if err != nil {
		return result, fmt.Errorf("marshalling error %s", string(transientAssetJSON))
	}

	caller_org_id := s.model.GetCreatorMspId()
	approvalEcdsa.OrgId = caller_org_id


	if (approvalEcdsa.Approval){
		value, err := stub.GetPrivateData("_implicit_org_"+ approvalEcdsa.OrgId, "ECDSASharedWallet-" + approvalEcdsa.UserId + "-" + approvalEcdsa.TokenId)
		if err != nil {
			return result, fmt.Errorf("Failed to get asset: %s with error: %s", "ECDSASharedWallet-"+ approvalEcdsa.UserId + "-" + approvalEcdsa.TokenId, err)
		}
		if value == nil {
			return result, fmt.Errorf("Asset not found: %s", approvalEcdsa.TxHash)
		}
	
		var participantInfo ECDSAParticipant
		err = json.Unmarshal(value, &participantInfo)
		if err != nil {
			return result, fmt.Errorf("Error decoding share information for org %s and key: %s", approvalEcdsa.OrgId, approvalEcdsa.TxHash)
		}		

		var share dealer.Share
		share.UnmarshalJSON([]byte(participantInfo.Share))
			
		var paillier paillier.SecretKey
		paillier.UnmarshalJSON([]byte(participantInfo.PaillierKey))

		p := participant.Participant{share, &paillier}

		pjson, _ := p.MarshalJSON()
		approvalEcdsa.Participant = string(pjson)
		approvalEcdsa.SK = participantInfo.PaillierKey

		approvalEcdsaBytes, _ := json.Marshal(approvalEcdsa) 
		return string(approvalEcdsaBytes), nil

	}


	return result, nil
}


func (s *SignatureReciever) ApproveECDSATx() (string, error) {
	var result string
	stub := s.model.GetNetworkStub()

	// Get new asset from transient map
	transientMap, err := stub.GetTransient()
	if err != nil {
		return result, fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		//log error to stdout
		return result, fmt.Errorf("asset not found in the transient map input")
	}
		
	var approvalEcdsa ApprovalTECDSA
	err = json.Unmarshal(transientAssetJSON, &approvalEcdsa)
	if err != nil {
		return result, fmt.Errorf("marshalling error %s: Errormsg: %s", string(transientAssetJSON), err)
	}

	if (approvalEcdsa.Approval){

		approvalEcdsaBytes, _ := json.Marshal(approvalEcdsa) 
		err = stub.PutPrivateData("_implicit_org_Org1MSP", approvalEcdsa.TxHash + ":" + approvalEcdsa.OrgId, approvalEcdsaBytes)
		if err != nil {
			return result, fmt.Errorf("Failed to set asset: %s", "secretshare" + approvalEcdsa.OrgId)
		}
		
		var event = SigningNotification{"Org1MSP", approvalEcdsa.TxHash, nil}
		eventAsBytes, _ := json.Marshal(event)
		err = stub.SetEvent("ApproveECDSATx", eventAsBytes)
		if err != nil {
			return result, fmt.Errorf("Failed to create event : %s", err)
		}

		return "Success", nil

	}


	return result, nil
}


func (s *SignatureReciever) PerformECDSARounds(key string) (string, error) {
	var result string
	
	stub := s.model.GetNetworkStub()

	var listApproval []ApprovalTECDSA
	for i:=1; i<6;i++{

		value, err := stub.GetPrivateData("_implicit_org_Org1MSP", key + ":Org" + strconv.Itoa(i) + "MSP")
		if err != nil {
			return result, fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
		}
		if value == nil {
			return result, fmt.Errorf("Asset not found: %s", key)
		}

		var approvalEcdsa ApprovalTECDSA
		err = json.Unmarshal(value, &approvalEcdsa)
		if err != nil {
			return result, fmt.Errorf("Error decoding share information for org %d and key: %s", i, key)
		}

		listApproval = append(listApproval, approvalEcdsa)
	}

	tshare := uint32(5)
	if (len(listApproval) == 5) {
		txAsBytes, err := stub.GetState(key)
		if err != nil {
			return result, fmt.Errorf("error in getting tx with Id %s: %s", key, err.Error())
		}
		if txAsBytes == nil {
			return result, fmt.Errorf("tx does not exist")
		}
	
		var tx ECDSAThresholdSignatureTransaction
		err = json.Unmarshal(txAsBytes, &tx)
		if err != nil {
			return result, fmt.Errorf("Error decoding pk Info: %s",err)
		}

		var msgHash []byte
		switch status := tx.TokenId; status {
		case "ETH":
			var ethtx types.Transaction
			err = json.Unmarshal([]byte(tx.Message), &ethtx)
			if err != nil {
				return result, fmt.Errorf("error decoding eth tx")
			}
			chainID := big.NewInt(3)
			signereth := types.NewEIP155Signer(chainID)
			h := signereth.Hash(&ethtx)
			msgHash = h[:]
		case "BTC":
			return result, fmt.Errorf("Token  not supported: %s", tx.TokenId)
		default:
			return result, fmt.Errorf("Token  not supported: %s", tx.TokenId)
		}
		
		//m := []byte(tx.Message)
		//msgHash := sha256.Sum256(m)

		pkInfoAsBytes, err := stub.GetState("SharedWalletECDSAPublicKey-" + tx.SignerID + "-" + tx.TokenId)
		if err != nil {
			return result, fmt.Errorf("error in getting pkInfo with Id %s %s", "SharedWalletECDSAPublicKey-" + tx.SignerID + "-" + tx.TokenId, err.Error())
		}
		if pkInfoAsBytes == nil {
			return result, fmt.Errorf("pkInfo does not exist")
		}
	
		var pkInfo ECDSAPublicInfoImport
		err = json.Unmarshal(pkInfoAsBytes, &pkInfo)
		if err != nil {
			return result, fmt.Errorf("Error decoding pk Info: %s",err)
		}

		k256 := btcec.S256()
		var pk curves.EcPoint
	
		pk.UnmarshalJSON([]byte(pkInfo.PK))
	
		

		//////

		var pubSharesMap = make(map[uint32]*dealer.PublicShare)

		var pubshare1 dealer.PublicShare
		var pubshare2 dealer.PublicShare
		var pubshare3 dealer.PublicShare
		var pubshare4 dealer.PublicShare
		var pubshare5 dealer.PublicShare
		
		json.Unmarshal([]byte(pkInfo.PubShares["1"]),&pubshare1)
		json.Unmarshal([]byte(pkInfo.PubShares["2"]),&pubshare2)
		json.Unmarshal([]byte(pkInfo.PubShares["3"]),&pubshare3)
		json.Unmarshal([]byte(pkInfo.PubShares["4"]),&pubshare4)
		json.Unmarshal([]byte(pkInfo.PubShares["5"]),&pubshare5)
		
		pubSharesMap[1] = &pubshare1
		pubSharesMap[2] = &pubshare2
		pubSharesMap[3] = &pubshare3
		pubSharesMap[4] = &pubshare4
		pubSharesMap[5] = &pubshare5	

		////////
		var pubKeysMap = make(map[uint32]*paillier.PublicKey)
		
		var pubkey1 paillier.PublicKey
		var pubkey2 paillier.PublicKey
		var pubkey3 paillier.PublicKey
		var pubkey4 paillier.PublicKey
		var pubkey5 paillier.PublicKey

		pubkey1.UnmarshalJSON([]byte(pkInfo.PubKeys["1"]))
		pubkey2.UnmarshalJSON([]byte(pkInfo.PubKeys["2"]))
		pubkey3.UnmarshalJSON([]byte(pkInfo.PubKeys["3"]))
		pubkey4.UnmarshalJSON([]byte(pkInfo.PubKeys["4"]))
		pubkey5.UnmarshalJSON([]byte(pkInfo.PubKeys["5"]))
		
		pubKeysMap[1] = &pubkey1
		pubKeysMap[2] = &pubkey2
		pubKeysMap[3] = &pubkey3
		pubKeysMap[4] = &pubkey4
		pubKeysMap[5] = &pubkey5

		///////

		signersMap := make(map[uint32]*participant.Signer, tshare)
		for i , val := range listApproval {
			var p participant.Participant
			p.UnmarshalJSON([]byte(val.Participant))

			var paillier paillier.SecretKey
			paillier.UnmarshalJSON([]byte(val.SK))

			p.SK = &paillier
			proofParams := &dealer.TrustedDealerKeyGenType{
				ProofParams: dealerParams,
			}
	
			signersMap[uint32(i)+1], _ = p.PrepareToSign(&pk, k256Verifier, k256, proofParams, pubSharesMap, pubKeysMap)

		}

	
		for i := range signersMap {
			fmt.Printf("Signer Map: %x\n", signersMap[i])
		}
	
		// Run signing rounds
		// Sign Round 1
		signerOut := make(map[uint32]*participant.Round1Bcast, tshare)
		for i, s := range signersMap {
			signerOut[i], _, err = s.SignRound1()
			if err != nil {
				return result , fmt.Errorf("Error SignRound1: %s",err)
			}
		}
	
		// Sign Round 2
		p2p := make(map[uint32]map[uint32]*participant.P2PSend)
		for i, s := range signersMap {
			in := make(map[uint32]*participant.Round1Bcast, tshare-1)
			for j := range signersMap {
				if i == j {
					continue
				}
				in[j] = signerOut[j]
			}
			p2p[i], err = s.SignRound2(in, nil) // TODO: fix me later
			if err != nil {
				return result , fmt.Errorf("Error SignRound2: %s",err)
			}
		}
	
		// Sign Round 3
		r3Bcast := make(map[uint32]*participant.Round3Bcast, tshare)
		for i, s := range signersMap {
			in := make(map[uint32]*participant.P2PSend, tshare-1)
			for j := range signersMap {
				if i == j {
					continue
				}
				in[j] = p2p[j][i]
			}
			r3Bcast[i], err = s.SignRound3(in)
			if err != nil {
				return result , fmt.Errorf("Error SignRound3: %s",err)
			}
		}
	
		// Sign Round 4
		r4Bcast := make(map[uint32]*participant.Round4Bcast, tshare)
		for i, s := range signersMap {
			in := make(map[uint32]*participant.Round3Bcast, tshare-1)
			for j := range signersMap {
				if i == j {
					continue
				}
				in[j] = r3Bcast[j]
			}
			r4Bcast[i], err = s.SignRound4(in)
			if err != nil {
				return result , fmt.Errorf("Error SignRound4: %s",err)
			}
		}
	
		// Sign Round 5
		r5Bcast := make(map[uint32]*participant.Round5Bcast, tshare)
		r5P2p := make(map[uint32]map[uint32]*participant.Round5P2PSend, tshare)
		for i, s := range signersMap {
			in := make(map[uint32]*participant.Round4Bcast, tshare-1)
			for j := range signersMap {
				if i == j {
					continue
				}
				in[j] = r4Bcast[j]
			}
			r5Bcast[i], r5P2p[i], err = s.SignRound5(in)
			if err != nil {
				return result , fmt.Errorf("Error SignRound5: %s",err)
			}
		}
	
		// Sign Round 6
		r6Bcast := make(map[uint32]*participant.Round6FullBcast, tshare)
		for i, s := range signersMap {
			in := make(map[uint32]*participant.Round5Bcast, tshare-1)
			for j := range signersMap {
				if i == j {
					continue
				}
				in[j] = r5Bcast[j]
			}
			r6Bcast[i], err = s.SignRound6Full(msgHash[:], in, r5P2p[i])
			if err != nil {
				return result , fmt.Errorf("Error SignRound6: %s",err)
			}
		}
	
		// Signature output
		var sig *curves.EcdsaSignature
		for i, s := range signersMap {
			in := make(map[uint32]*participant.Round6FullBcast, tshare-1)
			for j := range signersMap {
				if i == j {
					continue
				}
				in[j] = r6Bcast[j]
			}
	
			sig, _ = s.SignOutput(in)
	
		}
	
		fmt.Printf("\nOverall signature: (%d %d)\n", sig.R, sig.S)
	
		publicKey := ecdsa.PublicKey{
			Curve: ecc.P256k1(), //secp256k1
			X:     pk.X,
			Y:     pk.Y,
		}
	
		rtn := ecdsa.Verify(&publicKey, msgHash[:], sig.R, sig.S)
		if rtn {
			sigjson, err := json.Marshal(sig)
			if err != nil {
				return result, nil
			}
			return string(sigjson), nil
		} else {
			return "Not Verified", nil
		}

	}

	return result, nil
}


func (s *SignatureReciever) PostECDSASignature(txHash string, signature string) (ECDSAThresholdSignatureTransaction, error) {	

	var tx ECDSAThresholdSignatureTransaction
	stub := s.model.GetNetworkStub()
	txAsBytes, err := stub.GetState(txHash)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error in getting tx with Id %s %s", txHash, err.Error())
	}
	if txAsBytes == nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("tx does not exist")
	}

	err = json.Unmarshal(txAsBytes, &tx)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error decoding tx with Id %s %s", txHash, err.Error())
	}

	if tx.Status == "Pending" {
		var msgHash []byte
		switch status := tx.TokenId; status {
		case "ETH":
			var ethtx types.Transaction
			err = json.Unmarshal([]byte(tx.Message), &ethtx)
			if err != nil {
				return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error decoding eth tx")
			}
			chainID := big.NewInt(3)
			signereth := types.NewEIP155Signer(chainID)
			h := signereth.Hash(&ethtx)
			msgHash = h[:]
		case "BTC":
			return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("Token  not supported: %s", tx.TokenId)
		default:
			return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("Token  not supported: %s", tx.TokenId)
		}
	
		var pk curves.EcPoint
		pk.UnmarshalJSON([]byte(tx.PubKey))	
		publicKey := ecdsa.PublicKey{
			Curve: ecc.P256k1(), //secp256k1
			X:     pk.X,
			Y:     pk.Y,
		}
	
		var sig curves.EcdsaSignature
		err = json.Unmarshal([]byte(signature), &sig)
		if err != nil {
			return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error unmarshalling signature for tx %s transaction error %s", txHash, err.Error())
	
		}
	
		rtn := ecdsa.Verify(&publicKey, msgHash[:], sig.R, sig.S)
		if rtn {
			tx.Status = "Verified"
			tx.Signature = signature
		} else {
			tx.Status = "Not Verified"
		}
	
		txAsBytes , err = json.Marshal(tx)
		if err != nil {
			return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error encoding: Asset Id %s transaction error %s", txHash, err.Error())
	
		}
	
		err = stub.PutState(txHash, txAsBytes)
		if err != nil {
			return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error in saving: Asset Id %s transaction error %s", txHash, err.Error())
		}
	
		if tx.Status == "Verified" {
			txBytes, _ := json.Marshal(tx)
			var event = SigningNotification{"Org1MSP", tx.MessageHash, txBytes}
			eventAsBytes, _ := json.Marshal(event)
			err = stub.SetEvent("PostECDSASignature", eventAsBytes)
			if err != nil {
				return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("Failed to create event : %s", err)
			}
		}
	}	

	return tx, nil
	
}

func (s *SignatureReciever) UpdateTxReceipt(txHash string, txreceipt string) (ECDSAThresholdSignatureTransaction, error) {	
	var tx ECDSAThresholdSignatureTransaction

	stub := s.model.GetNetworkStub()
	txAsBytes, err := stub.GetState(txHash)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error in getting tx with Id %s %s", txHash, err.Error())
	}
	if txAsBytes == nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("tx does not exist")
	}

	err = json.Unmarshal(txAsBytes, &tx)
	if err != nil {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error decoding tx with Id %s %s", txHash, err.Error())
	}

	if tx.Status == "Verified" {
		tx.TxReceipt = txreceipt
		tx.Status = "Submitted"
		txAsBytes, err = json.Marshal(tx)
		if err != nil {
			return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error in json updated marshal %s transaction error %s", txHash, err.Error())
		}
		err = stub.PutState(txHash, txAsBytes)
		if err != nil {
			return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error in saving: Asset Id %s transaction error %s", txHash, err.Error())
		}
	}

	return tx, nil
	
}

func (s *SignatureReciever) QueryECDSAThresholdTransaction(transactionId string) (ECDSAThresholdSignatureTransaction, error) {

	if transactionId == "" {
		return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error in retrieving transaction, transaction id is empty")
	}

	stub := s.model.GetNetworkStub()

    txAsBytes, err := stub.GetState(transactionId)
    if err != nil {
        return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("error in getting transaction with Id %s %s", transactionId, err.Error())
    }
    if txAsBytes == nil {
        return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("Transaction with Id %s does not exist", transactionId)
    }

    var tx ECDSAThresholdSignatureTransaction
    unmarshalError := json.Unmarshal(txAsBytes, &tx)
    if unmarshalError != nil {
        return ECDSAThresholdSignatureTransaction{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
    }

    return tx, nil
}

func (s *SignatureReciever) GetAllPendingECDSATxs() (interface{}, error) {
	query := `{"selector": {"status": "Pending", "txType": "ecdsathreshold"}}`
	return s.model.Query(query)
}

func (s *SignatureReciever) GetAllThresholdECDSATxs() (interface{}, error) {
	query := `{"selector": {"txType": "ecdsathreshold"}}`
	return s.model.Query(query)
}

func (s *SignatureReciever) GetPublicInfo(userId string, tokenId string) ([]byte, error) {	
	stub := s.model.GetNetworkStub()
	pkInfoAsBytes, err := stub.GetState("SharedWalletECDSAPublicKey-" + userId + "-" + tokenId)
	if err != nil {
		return nil, fmt.Errorf("error in getting pkInfo with Id %s %s", "SharedWalletECDSAPublicKey-"+userId + "-" +  tokenId, err.Error())
	}
	if pkInfoAsBytes == nil {
		return nil, fmt.Errorf("pkInfo does not exist")
	}

	return pkInfoAsBytes, nil
	
}

func (s *SignatureReciever) GetWalletId(userId string, tokenId string) (string, error) {	
	stub := s.model.GetNetworkStub()
	pkInfoAsBytes, err := stub.GetState("SharedWalletECDSAPublicKey-" + userId + "-" + tokenId)
	if err != nil {
		return "", fmt.Errorf("error in getting pkInfo with Id %s %s", "SharedWalletECDSAPublicKey-"+userId + "-" +  tokenId, err.Error())
	}
	if pkInfoAsBytes == nil {
		return "", fmt.Errorf("pkInfo does not exist")
	}

	var pk ECDSAPublicInfo
	err = json.Unmarshal(pkInfoAsBytes, &pk)
	if err != nil {
		return "", fmt.Errorf("error in decoding pkInfo json with Id %s %s", "SharedWalletECDSAPublicKey-"+userId + "-" +  tokenId, err.Error())
	}

	var pubKeyK ecdsa.PublicKey
	err = json.Unmarshal([]byte(pk.PK), &pubKeyK)
	if err != nil {
		return "", fmt.Errorf("error in decoding pkInfo json with Id %s %s", "SharedWalletECDSAPublicKey-"+userId + "-" +  tokenId, err.Error())
	}

	switch status := tokenId; status {
	case "ETH":
		address := crypto.PubkeyToAddress(pubKeyK).Hex()
		return address, nil	
	case "BTC":
		compressedBytes := SerializeUncompressed(&pubKeyK)
		addresspubkey, _ := btcutil.NewAddressPubKey(compressedBytes, &btcchain.TestNet3Params)
		address, err := btcutil.DecodeAddress(addresspubkey.EncodeAddress(), &btcchain.TestNet3Params)
		if err != nil {
			return "", fmt.Errorf("Error generating BTC address err: %v\n", err)
		}
		return address.EncodeAddress(), nil	
	default:
		return "", fmt.Errorf("Token address generation not supported: %s", tokenId)
	}
    
	
}

func (s *SignatureReciever) PrepareEthTx(userId string, tokenId string, destination string, ethvalue string) (ETHTX, error) {	
	var result ETHTX

	stub := s.model.GetNetworkStub()
	pkInfoAsBytes, err := stub.GetState("SharedWalletECDSAPublicKey-" + userId + "-" + tokenId)
	if err != nil {
		return result, fmt.Errorf("error in getting pkInfo with Id %s %s", "SharedWalletECDSAPublicKey-"+userId + "-" +  tokenId, err.Error())
	}
	if pkInfoAsBytes == nil {
		return result, fmt.Errorf("pkInfo does not exist")
	}

	var pk ECDSAPublicInfo
	err = json.Unmarshal(pkInfoAsBytes, &pk)
	if err != nil {
		return result, fmt.Errorf("error in decoding pkInfo json with Id %s %s", "SharedWalletECDSAPublicKey-"+userId + "-" +  tokenId, err.Error())
	}

	var pubKeyK ecdsa.PublicKey
	err = json.Unmarshal([]byte(pk.PK), &pubKeyK)
	if err != nil {
		return result, fmt.Errorf("error in decoding pkInfo json with Id %s %s", "SharedWalletECDSAPublicKey-"+userId + "-" +  tokenId, err.Error())
	}

	client, err := ethclient.Dial("https://ropsten.infura.io/v3/234372f56a5d46f5820dc9780ad2f0b0")
	if err != nil {
		return result, fmt.Errorf("Error getting client: %s", err.Error())
	}
		
	fromAddress :=  crypto.PubkeyToAddress(pubKeyK) 
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return result, fmt.Errorf("Error getting nonce: %s", err.Error())
	}

	ethamount := new(big.Float)
	ethamount, ok := ethamount.SetString(ethvalue)
	if ok != true  {
		return result, fmt.Errorf("error converting eth value to big float")
	}
	value := etherToWei(ethamount) // in wei (1 eth)
	gasLimit := uint64(21000)                // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return result, fmt.Errorf("error getting gas value %s", err.Error())
	}
	
	toAddress := common.HexToAddress(destination)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return result, fmt.Errorf("Error getting chainId: %s", err.Error())
	}
	
	chainID = big.NewInt(3)
	signereth := types.NewEIP155Signer(chainID)
	h := signereth.Hash(tx)
	hstring, err := json.Marshal(h)
	if err != nil {
		return result, fmt.Errorf("Error generating hash string: %s", err.Error())
	}

	txstring, err := json.Marshal(tx)
	if err != nil {
		return result, fmt.Errorf("Error generating tx string: %s", err.Error())
	}
	result.Hash = string(hstring)
	result.TX = string(txstring)

	return result, nil
}

func (s *SignatureReciever) SubmitEthTx(txhash string) (string, error) {	
	var result string
	stub := s.model.GetNetworkStub()

    txAsBytes, err := stub.GetState(txhash)
    if err != nil {
        return result, fmt.Errorf("error in getting transaction with Id %s %s", txhash, err.Error())
    }
    if txAsBytes == nil {
        return result, fmt.Errorf("Transaction with Id %s does not exist", txhash)
    }

    var tx ECDSAThresholdSignatureTransaction
    unmarshalError := json.Unmarshal(txAsBytes, &tx)
    if unmarshalError != nil {
        return result, fmt.Errorf("marshalling error %s", unmarshalError.Error())
    }

	var ethtx types.Transaction
	err = json.Unmarshal([]byte(tx.Message), &ethtx)
	if err != nil {
		return result, fmt.Errorf("error decoding eth tx")
	}

	var signature curves.EcdsaSignature
	err = json.Unmarshal([]byte(tx.Signature), &signature)
	if err != nil {
        return result, fmt.Errorf("error unmarshalling signature for tx %s transaction error %s", txhash, err.Error())

	}

	sigbytes := append(signature.R.Bytes(),signature.S.Bytes()...)
	if (signature.V > 0){
		sigbytes = append(sigbytes,big.NewInt(int64(signature.V)).Bytes()...)
	} else if (signature.V == 0) {
		empty := []byte{0}
		sigbytes = append(sigbytes,empty[0])
	}


	client, err := ethclient.Dial("https://ropsten.infura.io/v3/234372f56a5d46f5820dc9780ad2f0b0")
	if err != nil {
		return result, fmt.Errorf("Error getting client: %s", err.Error())
	}

	chainID := big.NewInt(3)
	signereth := types.NewEIP155Signer(chainID)
		
	signedTx, err := ethtx.WithSignature(signereth, sigbytes)
	if err != nil {
		return result, fmt.Errorf("Error appending signature: %s", err.Error())
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return result, fmt.Errorf("Error sending tx via client: %s", err.Error())
	}

	result = signedTx.Hash().Hex()

	return result, nil
}

type ETHTX struct{
	Hash string `json:"hash"`   //common.Hash
	TX   string `json:"tx"`     // types.Transaction
}

func etherToWei(eth *big.Float) *big.Int {
	truncInt, _ := eth.Int(nil)
	truncInt = new(big.Int).Mul(truncInt, big.NewInt(params.Ether))
	fracStr := strings.Split(fmt.Sprintf("%.18f", eth), ".")[1]
	fracStr += strings.Repeat("0", 18 - len(fracStr))
	fracInt, _ :=  new(big.Int).SetString(fracStr, 10)
	wei := new(big.Int).Add(truncInt, fracInt)
	return wei;
}