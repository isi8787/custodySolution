package model

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"crypto/x509"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
    "sort"
    "github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"example.com/TokenCC/lib/util"
	"example.com/TokenCC/lib/util/validators"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Model struct {
	Stub shim.ChaincodeStubInterface
}

func (m *Model) SetStub(stub shim.ChaincodeStubInterface) {
	m.Stub = stub
}

func GetNewModel(stub shim.ChaincodeStubInterface) *Model {
	var m Model
	m.Stub = stub
	return &m
}

// separate by no escaped commas
var sepPattern *regexp.Regexp = regexp.MustCompile(`((?:^|[^\\])(?:\\\\)*),`)

func splitUnescapedComma(str string) []string {
	ret := []string{}
	indexes := sepPattern.FindAllStringIndex(str, -1)
	last := 0
	for _, is := range indexes {
		ret = append(ret, str[last:is[1]-1])
		last = is[1]
	}
	ret = append(ret, str[last:])
	return ret
}

func (m *Model) GetId(obj interface{}) (string, error) {
	return m.GenerateID(obj, true, "customGetter")
}

func (m *Model) GenerateID(obj interface{}, checkTimestampTag bool, caller string) (string, error) {
	objValue := reflect.ValueOf(obj).Elem()
	objType := objValue.Type()

	for i := 0; i < objType.NumField(); i++ {
		structField := objType.Field(i)
		structFieldName := structField.Name
		_, ok := structField.Tag.Lookup("id")
		if ok {
			_, ok2 := structField.Tag.Lookup("derived")
			if !ok2 {
				idValue := objValue.Field(i)
				idValInterface := idValue
				idString := fmt.Sprintf("%v", idValInterface)
				return idString, nil
			} else {
				if caller == "update" || caller == "save" {
					if objValue.Field(i).String() != "" {
						return "", fmt.Errorf("error in generating the derived key %s. Derived key is required to be omitted from the input, you have passed %s", structFieldName, objValue.Field(i).String())
					}
				}
				derivedTagString := structField.Tag.Get("derived")
				splittedStringArray := splitUnescapedComma(derivedTagString)

				if len(splittedStringArray) < 2 {
					return "", fmt.Errorf("the tag is in improper format. Need atleast 2 fields in derived tag i.e. strategy and format. Please check documentation")
				}
				var propertyMap = make(map[string]string)
				var arguments []string

				for i := 0; i < len(splittedStringArray); i++ {
					var propName string
					var propValue string
					propSplitted := strings.SplitN(splittedStringArray[i], "=", 2)
					if len(propSplitted) < 2 {
						arguments = append(arguments, propSplitted[0])
					} else {
						propName = propSplitted[0]
						propValue = propSplitted[1]
					}
					propertyMap[propName] = propValue
				}
				strategy := propertyMap["strategy"]
				format := propertyMap["format"]
				if strategy == "" || format == "" {
					return "", fmt.Errorf("the tag is not in proper format. Strategy or Format is missing")
				}
				if strategy != "concat" && strategy != "hash" {
					return "", fmt.Errorf("the strategy supported are only concat or hash. You have provided %s", strategy)
				}
				// validating the format
				pattern := []rune(format)
				percent := false
				for j := 0; j < len(pattern); j++ {
					if pattern[j] == '%' {
						for j < len(pattern)-1 && pattern[j+1] == '%' {
							percent = true
							j++
						}
						if percent {
							j++
							percent = false
							continue
						} else if j < len(pattern)-1 && pattern[j+1] == 't' {
							j++
							continue
						} else if j < len(pattern)-1 && pattern[j+1] >= '0' && pattern[j+1] <= '9' {
							ptr := j + 1
							index := ""
							for ptr < len(pattern) && pattern[ptr] >= '0' && pattern[ptr] <= '9' {
								str := string(pattern[ptr])
								index = index + str
								ptr++
							}
							num, _ := strconv.Atoi(index)
							if len(arguments) < num {
								return "", fmt.Errorf("no argument found for position specifier %%%s for derived field %s. Please use valid template string with correct identifiers", index, structFieldName)
							}
						} else {
							j++
							if j < len(pattern) {
								return "", fmt.Errorf("invalid position specifier %%%s  for derived field %s. Please use valid template string with correct identifiers", string(pattern[j]), structFieldName)
							}
						}
					}
				}
				// validation is complete

				newString := ""
				for j := 0; j < len(pattern); j++ {
					if pattern[j] == '%' {
						for j < len(pattern)-1 && pattern[j+1] == '%' {
							percent = true
							newString = newString + string(pattern[j+1])
							j++
						}
						if percent && j < len(pattern)-1 {
							newString = newString + string(pattern[j+1])
							j++
							percent = false
							continue
						}
						if j < len(pattern)-1 && pattern[j+1] == 't' {
							if !checkTimestampTag {
								timestamp, err := m.GetTransactionTimestamp()
								if err != nil {
									return "", fmt.Errorf("getting transaction timestamp failed")
								}

								secs := fmt.Sprintf("%d", timestamp.Seconds)
								nano := fmt.Sprintf("%d", timestamp.Nanos)
								t := secs + nano
								newString = newString + t
								j++
							} else {
								if caller == "update" {
									if objValue.Field(i).String() == "" {
										return "", fmt.Errorf("the derived key %s contains timestamp. Cannot generate derived key. Please pass the original derived key to update the asset", structFieldName)
									} else {
										return objValue.Field(i).String(), nil
									}
								}
								if caller == "customGetter" {
									return "", fmt.Errorf("the derived field %s contains timestamp. Cannot generate derived key for this field", structFieldName)
								}
							}
						} else if j < len(pattern)-1 && pattern[j+1] > '0' && pattern[j+1] <= '9' {
							ptr := j + 1
							index := ""
							for ptr < len(pattern) && pattern[ptr] >= '0' && pattern[ptr] <= '9' {
								index = index + string(pattern[ptr])
								ptr++
							}
							j = j + len(index)
							num, _ := strconv.Atoi(index)
							if len(arguments) >= num {
								val := objValue.FieldByName(arguments[num-1])
								if !val.IsValid() {
									return "", fmt.Errorf("error in generating derived key, field with name %s not found", arguments[num-1])
								}
								if val.Type().Name() != "string" {
									return "", fmt.Errorf("error in generating derived key, field %s is not of type string. Derived keys accept only string types in format", arguments[num-1])
								}
								newString = newString + val.String()
							}
						} else {
							newString = newString + string(pattern[j])
						}
					} else {
						newString = newString + string(pattern[j])
					}
				}
				finalString := newString
				if finalString == "" {
					return "", fmt.Errorf("derived field %s is generated as an empty string after processing the format. Please make sure that all the fields required by the derived key are not omitted or empty", structFieldName)
				}
				if strategy == "concat" {
					algorithmName := propertyMap["algorithm"]
					if algorithmName != "" {
						return "", fmt.Errorf("strategy concat does not require any algorithm")
					}
					objValue.Field(i).Set(reflect.ValueOf(finalString))
					tag := structField.Tag.Get("validate")
					if tag != "-" && tag != "" {
						err := validators.Validate(reflect.ValueOf(finalString).Interface(), tag)
						if err != nil {
							return "", fmt.Errorf("error in validating Id field derived string is %s. %s", finalString, err.Error())
						}
						if len(finalString) > 128 {
							return "", fmt.Errorf("the concatenated string %s is greater than 128 characters", finalString)
						}
					}
					return finalString, nil
				}
				if strategy == "hash" {
					algorithmName := propertyMap["algorithm"]
					if algorithmName == "" {
						algorithmName = "sha256"
					}
					var encodedString string
					if algorithmName == "sha256" {
						h := sha256.New()
						h.Write([]byte(finalString))
						k := h.Sum(nil)
						encodedString = fmt.Sprintf("%x", k)
					} else if algorithmName == "md5" {
						h := md5.New()
						h.Write([]byte(finalString))
						k := h.Sum(nil)
						encodedString = fmt.Sprintf("%x", k)
					} else {
						return "", fmt.Errorf("the hash alogrithm %s is not supported", algorithmName)
					}
					objValue.Field(i).Set(reflect.ValueOf(encodedString))
					//validate string
					tag := structField.Tag.Get("validate")
					if tag != "-" && tag != "" {
						err := validators.Validate(reflect.ValueOf(encodedString).Interface(), tag)
						if err != nil {
							return "", fmt.Errorf("error in validating Id field derived string is %s. %s", encodedString, err.Error())
						}
					}
					return encodedString, nil
				}
			}
		}
	}
	return "", errors.New("id tag is not set")
}

// Save writes the asset to the ledger
func (m *Model) Save(args ...interface{}) (interface{}, error) {
	stub := m.GetNetworkStub()
	obj := args[0]

	if len(args) < 1 {
		return nil, fmt.Errorf("no object passed for saving in the ledger")
	}

	_, err := util.SetFinaTagInput(obj)
	if err != nil {
		return nil, fmt.Errorf("setting final input for token failed %s", err.Error())
	}

	id, idErr := m.GenerateID(obj, false, "save")
	if idErr != nil {
		return nil, fmt.Errorf("error in getting Id. %s", idErr.Error())
	}

	_, err = m.Get(id)
	if err == nil {
		return nil, fmt.Errorf("error in saving: asset already exist in ledger with Id %s ", id)
	}

	if len(args) > 1 {
		ptrToAsset := obj
		assetValue := reflect.ValueOf(ptrToAsset).Elem()
		metadata := args[1]
		metdataField := assetValue.FieldByName("Metadata")
		metdataField.Set(reflect.ValueOf(metadata))
	}

	assetAsBytes, errMarshal := json.Marshal(obj)
	if errMarshal != nil {
		return nil, fmt.Errorf("error in saving: Asset Id %s marshal error %s", id, errMarshal.Error())
	}

	errPut := stub.PutState(id, assetAsBytes)
	if errPut != nil {
		return nil, fmt.Errorf("error in saving: Asset Id %s transaction error %s", id, errPut.Error())
	}

	return obj, nil
}

func (m *Model) GenerateCompositeKey(indexName string, attributes []string) (string, error) {
	stub := m.GetNetworkStub()
	if len(attributes) == 0 {
		const errorMessage = "attributes param is expected to be an array of string"
		return "", fmt.Errorf(errorMessage)
	}

	compositeKey, err := stub.CreateCompositeKey(indexName, attributes)
	if err != nil {
		return "", fmt.Errorf("failed creating composite Key")
	}
	return compositeKey, nil
}

func (m *Model) GetByCompositeKey(key string, columns []string, index int) (interface{}, error) {

	stub := m.GetNetworkStub()

	resultsIterator, err := stub.GetStateByPartialCompositeKey(key, columns)
	if err != nil {
		return nil, fmt.Errorf("error in returning iterator: %d", resultsIterator)
	}

	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error in getting GetByCompositeKey: iteration error %s", err.Error())
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(queryResult.Key)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

		returnedID := compositeKeyParts[index]

		assetAsBytes, err := stub.GetState(returnedID)
		if err != nil {
			return nil, fmt.Errorf("error in getting Asset with id %s %s", returnedID, err.Error())
		}
		if assetAsBytes == nil {
			return nil, fmt.Errorf("error in getting: Asset with id %s does not exist", returnedID)
		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(returnedID)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(assetAsBytes))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	var result []interface{}
	unmarshalError := json.Unmarshal(buffer.Bytes(), &result)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting history by id: unmarshalling error %s", unmarshalError.Error())
	}
	return result, nil
}

func (m *Model) GetTransactionId() string {
	return m.GetNetworkStub().GetTxID()
}

func (m *Model) GetTransactionTimestamp() (*timestamp.Timestamp, error) {
	return m.GetNetworkStub().GetTxTimestamp()
}

func (m *Model) GetChannelID() string {
	return m.GetNetworkStub().GetChannelID()
}

func (m *Model) GetCreator() ([]byte, error) {
	return m.GetNetworkStub().GetCreator()
}

func (m *Model) GetSignedProposal() (*peer.SignedProposal, error) {
	return m.GetNetworkStub().GetSignedProposal()
}

func (m *Model) GetArgs() [][]byte {
	return m.GetNetworkStub().GetArgs()
}

func (m *Model) GetStringArgs() []string {
	return m.GetNetworkStub().GetStringArgs()
}

func (m *Model) GetNetworkStub() shim.ChaincodeStubInterface {
	return m.Stub
}

func ConvertAssetFromBytes(assetAsBytes []byte, result interface{}, AssetType string) (interface{}, error) {

	var genericResult interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &genericResult)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting: marshalling error %s", unmarshalError.Error())
	}
	assetTypeFromLedgerString := genericResult.(map[string]interface{})["AssetType"].(string)
	inputAssetType := AssetType
	if inputAssetType == "" {
		inputAssetTypeString := reflect.ValueOf(result).Elem().Type().String()
		inputAssetType = strings.Split(inputAssetTypeString, ".")[1]
	}

	// Putting the check for backward compatibility
	if inputAssetType != assetTypeFromLedgerString && !checkOldAssetType(assetTypeFromLedgerString, inputAssetType) {
		return nil, fmt.Errorf("AssetType Mismatch")
	}
	unmarshalError = json.Unmarshal(assetAsBytes, result)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting: marshalling error %s", unmarshalError.Error())
	}
	return result, nil
}

func (m *Model) IsTokenType(token_id string) error {
	tokenData, err := m.Get(token_id)
	if err != nil {
		return fmt.Errorf("error in getting token with token_id %s %s", token_id, err.Error())
	}
	tokenMap, ok := tokenData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("no token exist with token_id %s", token_id)
	}
	assetType, ok := tokenMap["AssetType"]
	if !ok {
		return fmt.Errorf("no token exist with token_id %s", token_id)
	}
	if assetType.(string) != "otoken" {
		return fmt.Errorf("no token exist with token_id %s", token_id)
	}
	return nil
}

func (m *Model) Get(Id string, result ...interface{}) (interface{}, error) {
	stub := m.GetNetworkStub()

	assetAsBytes, err := stub.GetState(Id)

	if err != nil {
		return nil, fmt.Errorf("error in getting Asset with Id %s %s", Id, err.Error())
	}

	if assetAsBytes == nil {
		return nil, fmt.Errorf("error in getting: Asset with Id %s does not exist", Id)
	}

	if len(result) > 0 {
		_, err := ConvertAssetFromBytes(assetAsBytes, result[0], "")
		if err != nil {
			return nil, fmt.Errorf("no asset exist with id %s %s", Id, err.Error())
		}
	}
	var genericResult interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &genericResult)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting: marshalling error %s", unmarshalError.Error())
	}
	return genericResult, nil
}

// Update the asset to the ledger
func (m *Model) Update(args ...interface{}) (interface{}, error) {
	stub := m.GetNetworkStub()

	obj := args[0]
	_, err := util.SetFinaTagInput(obj)
	if err != nil {
		return nil, fmt.Errorf("setting final input for model failed %s", err.Error())
	}

	id, idErr := m.GenerateID(obj, true, "update")
	if idErr != nil {
		return nil, fmt.Errorf("error in getting ID: %s", idErr.Error())
	}

	assetAsBytes, err := stub.GetState(id)
	if err != nil {
		return nil, fmt.Errorf("error in updating: Unable to get the asset from ledger with id %s %s", id, err.Error())
	}
	if assetAsBytes == nil {
		return nil, fmt.Errorf("error in updating: Asset with id %s does not exist", id)
	}

	assetAsBytes, errMarshal := json.Marshal(obj)
	if errMarshal != nil {
		return nil, fmt.Errorf("error in updating: Asset Id %s marshal error %s", id, errMarshal.Error())
	}

	errPut := stub.PutState(id, assetAsBytes)
	if errPut != nil {
		return nil, fmt.Errorf("error in updating: Asset Id %s marshal error %s", id, errPut.Error())
	}

	return obj, nil
}

// Delete deletes the asset from the ledger
func (m *Model) Delete(Id string) (interface{}, error) {
	stub := m.GetNetworkStub()

	assetAsBytes, err := stub.GetState(Id)
	if err != nil {
		return nil, fmt.Errorf("error in deleting: could not find asset with Id %s %s", Id, err.Error())
	}
	if assetAsBytes == nil {
		return nil, fmt.Errorf("error in deleting: Asset with Id %s does not exist", Id)
	}

	errPut := stub.DelState(Id)
	if errPut != nil {
		return nil, fmt.Errorf("error in deleting: failed to delete asset with Id %s error %s", Id, errPut.Error())
	}

	var result interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &result)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in deleting: marshalling error %s", unmarshalError.Error())
	}
	return result, nil
}

// Query runs the given transaction on the peer
func (m *Model) Query(queryString string) ([]interface{}, error) {
	stub := m.GetNetworkStub()
	fmt.Printf("Query: queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("Query: iteration error %s", err.Error())
	}

	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	var result []interface{}
	unmarshalError := json.Unmarshal(buffer.Bytes(), &result)
	if unmarshalError != nil {
		return nil, fmt.Errorf("Query: unmarshalling result error %s", unmarshalError.Error())
	}
	return result, nil
}

func (m *Model) GetByRangeFromLedger(startKey string, endKey string) (bytes.Buffer, error) {
	stub := m.GetNetworkStub()
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)

	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error in getting by range: %s", err.Error())
	}

	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return bytes.Buffer{}, fmt.Errorf("error in getting by range: iteration error %s", err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")
	return buffer, nil
}

func checkOldAssetType(assetTypeString string, assetType string) bool {
	splittedStringSlice := strings.Split(assetTypeString, ".")
	if len(splittedStringSlice) < 2 {
		return false
	}
	if splittedStringSlice[1] == assetType {
		return true
	}
	return false
}

func FilterRangeResultsByAssetType(assetType string, buffer bytes.Buffer) ([]map[string]interface{}, error) {
	var result []interface{}
	unmarshalError := json.Unmarshal(buffer.Bytes(), &result)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting by range: unmarshalling error %s", unmarshalError.Error())
	}

	var resultAssets []map[string]interface{}
	for i := 0; i < len(result); i++ {
		entry := result[i].(map[string]interface{})
		value := entry["Record"]
		mapAsset := value.(map[string]interface{})
		assetTypeString := mapAsset["AssetType"].(string)

		// Keeping backward compatibility we are checking for oldAssetTypes as well for normal non token assets
		if assetTypeString == assetType || checkOldAssetType(assetTypeString, assetType) {
			resultAssets = append(resultAssets, mapAsset)
		}
	}

	return resultAssets, nil
}
func (m *Model) getCreatorMap() (map[string]interface{}, error) {
	stub := m.GetNetworkStub()

	creator, err := stub.GetCreator()
	if err != nil {
		return nil, err
	}
	//Create a SerializedIdentity to hold Unmarshal GetCreator() result
	sId := &msp.SerializedIdentity{}
	//Unmarshal the creator from []byte to structure
	err1 := proto.Unmarshal(creator, sId)
	if err1 != nil {
		return nil, err1
	}

	response := make(map[string]interface{})
	response["mspid"] = sId.Mspid
	return response, nil
}

func (m *Model) GetTransientMap() (map[string][]byte, error) {
	transientMap, err := m.GetNetworkStub().GetTransient()
	if err != nil {
		return nil, fmt.Errorf("transient map is not found in the transaction request")
	}
	return transientMap, nil
}

func (m *Model) GetTransientMapKey(key string) ([]byte, error) {
	transientMap, err := m.GetNetworkStub().GetTransient()
	if err != nil {
		return nil, fmt.Errorf("transient map is not found in the transaction request")
	}
	if userId, ok := transientMap[key]; ok {
		return userId, nil
	} else {
		return nil, fmt.Errorf("key %s is not found in transient map", key)
	}
}

func (m *Model) GetObpRestUser() string {
	restUserBytes, err := m.GetTransientMapKey("bcsRestClientId")
	if err != nil {
		return ""
	}
	return string(restUserBytes)
}

func (m *Model) GetCreatorMspId() string {
	creatorMap, err := m.getCreatorMap()
	if err != nil {
		return ""
	}
	return creatorMap["mspid"].(string)
}

func (m *Model) GetUserId() (string, error) {
	stub := m.GetNetworkStub()
	bcsRest := m.GetObpRestUser()

	if runtime.GOOS == "windows" {
		return bcsRest, nil
	} else {
		//mspId := m.GetCreatorMspId()
		x509, err := cid.GetX509Certificate(stub)
		if err != nil {
			return "", fmt.Errorf("error in getting certificate %s", err.Error())
		}
		enrollmentUserId := x509.Subject.CommonName
		var userId string
		if bcsRest != "" {
			userId = bcsRest
		} else {
			userId = enrollmentUserId
		}
		return userId, nil
	}
}


func (m *Model) GetUserPublicKey() (string, error) {
	stub := m.GetNetworkStub()
	bcsRest := m.GetObpRestUser()

	if runtime.GOOS == "windows" {
		return bcsRest, nil
	} else {
		//mspId := m.GetCreatorMspId()
		x509bytes, err := cid.GetX509Certificate(stub)
		if err != nil {
			return "", fmt.Errorf("error in getting certificate %s", err.Error())
		}

		pub, err := x509.ParsePKIXPublicKey(x509bytes.RawSubjectPublicKeyInfo)
		if err != nil {
			panic("failed to parse DER encoded public key: " + err.Error())
		}

		switch pub := pub.(type) {
		case *rsa.PublicKey:
			return "", fmt.Errorf(" RSA public key Not Supported")
		case *dsa.PublicKey:
			return "", fmt.Errorf("DSA public key Not Supported")
		case *ecdsa.PublicKey:
			return (pub.X.String() + ":" + pub.Y.String()), nil
		case ed25519.PublicKey:
			return "", fmt.Errorf("ED25519 public key Not Supported")
		default:
			return "", fmt.Errorf("X509 public key format Not Supported")
		}
		
		return "", fmt.Errorf("Unable to get Key")
	}
}

// GetByRange gets all the assets with key between the provided range
func (m *Model) GetByRange(startKey string, endKey string, asset ...interface{}) ([]map[string]interface{}, error) {
	if len(asset) > 0 {
		buffer, err := m.GetByRangeFromLedger(startKey, endKey)
		if err != nil {
			return nil, err
		}
		inputAssetTypeString := reflect.TypeOf(asset[0]).Elem().Elem().String()
		inputAssetType := strings.Split(inputAssetTypeString, ".")[1]
		resultAssets, err := FilterRangeResultsByAssetType(inputAssetType, buffer)
		if err != nil {
			return nil, err
		}
		mapBytes, err := json.Marshal(resultAssets)
		if err != nil {
			return nil, fmt.Errorf("error in marshalling map %s", err.Error())
		}
		err = json.Unmarshal(mapBytes, asset[0])
		if err != nil {
			return nil, fmt.Errorf("error in unmarshalling map %s", err.Error())
		}
		return resultAssets, nil
	}
	buffer, err := m.GetByRangeFromLedger(startKey, endKey)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	unmarshalError := json.Unmarshal(buffer.Bytes(), &result)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting by range: unmarshalling error %s", unmarshalError.Error())
	}
	return result, nil
}

// GetHistoryById gets the history of an asset from the ledger
// GetHistoryById gets the history of an asset from the ledger
func (m *Model) GetHistoryById(Id string) ([]interface{}, error) {
	recordKey := Id
	stub := m.GetNetworkStub()

	resultsIterator, err := stub.GetHistoryForKey(recordKey)
	if err != nil {
		return nil, fmt.Errorf("error in getting history by id: %s", err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	var responses []*queryresult.KeyModification

	for resultsIterator.HasNext() {

		response, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error in getting history by id: iteration error %s", err.Error())
		}
		responses = append(responses, response)
	}

	sort.Slice(responses, func(i, j int) bool {
		t1 := time.Unix(responses[i].Timestamp.Seconds, int64(responses[i].Timestamp.Nanos))
		t2 := time.Unix(responses[j].Timestamp.Seconds, int64(responses[j].Timestamp.Nanos))
		return t1.After(t2)
	})

	for _, response := range responses {
		if err != nil {
			return nil, fmt.Errorf("error in getting history by id: iteration error %s", err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		// corresponding value null. Else, we will write the response.Values
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")
		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	var result []interface{}
	unmarshalError := json.Unmarshal(buffer.Bytes(), &result)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting history by id: unmarshalling error %s", unmarshalError.Error())
	}

	return result, nil
}
