package transaction

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"example.com/TokenCC/lib/model"
	"example.com/TokenCC/lib/util"
)

type Transaction struct {
	AssetType               string  `json:"AssetType" final:"otransaction"`
	TokenId                 string  `json:"TokenId"`
	TransactionId           string  `json:"TransactionId" id:"true"`
	FromAccountId           string  `json:"FromAccountId"`
	ToAccountId             string  `json:"ToAccountId"`
	TransactionType         string  `json:"TransactionType"`
	Amount                  float64 `json:"Amount"`
	Timestamp               string  `json:"Timestamp"`
	NumberOfSubTransactions float64 `json:"NumberOfSubTransactions"`
	HoldingId               string  `json:"HoldingId"`
}

type TransactionReciever struct {
	model *model.Model
}

func GetNewTransactionReciever(m *model.Model) *TransactionReciever {
	var trx TransactionReciever
	trx.model = m
	return &trx
}

const trx_asset_type = "otransaction"

// TODO: validate conditions on Amount, Transaction Type
func (trx *TransactionReciever) CreateTransaction(TokenId string, FromAccountId string, ToAccountId string, TransactionType string, Amount float64, NumberOfSubTransactions float64, HoldingId string) (Transaction, error) {
	err := trx.model.IsTokenType(TokenId)
	if err != nil {
		return Transaction{}, err
	}
	var trxAsset Transaction
	trxAsset.FromAccountId = FromAccountId
	trxAsset.ToAccountId = ToAccountId
	trxAsset.TokenId = TokenId
	trxAsset.TransactionType = TransactionType
	trxAsset.Amount = Amount
	trxAsset.NumberOfSubTransactions = NumberOfSubTransactions
	trxAsset.HoldingId = HoldingId
	currTime, err := trx.model.GetTransactionTimestamp()
	if err != nil {
		return Transaction{}, fmt.Errorf("Error in fetching transaction timestamp from the ledger")
	}

	trxAsset.AssetType = trx_asset_type

	if NumberOfSubTransactions <= 0 || TransactionType == "BULKTRANSFER" {
		trxAsset.TransactionId = trx_asset_type + `~` + trx.model.GetNetworkStub().GetTxID()
	} else {
		h := md5.New()
		h.Write([]byte(fmt.Sprint(NumberOfSubTransactions)))
		k := h.Sum(nil)
		trxAsset.TransactionId = trx_asset_type + `~` + trx.model.GetNetworkStub().GetTxID() + `~` + fmt.Sprintf("%x", k)
	}

	trxAsset.Timestamp = time.Unix(currTime.Seconds, int64(currTime.Nanos)).Format(time.RFC3339)

	err = trx.validateTransactionData(&trxAsset)
	if err != nil {
		return Transaction{}, fmt.Errorf("Error in validating Transaction, %s", err.Error())
	}
	_, err = trx.model.Save(&trxAsset)
	if err != nil {
		return Transaction{}, err
	}

	return trxAsset, nil
}

func (trx *TransactionReciever) get(Id string) (Transaction, error) {
	stub := trx.model.GetNetworkStub()
	assetAsBytes, err := stub.GetState(Id)
	if err != nil {
		return Transaction{}, fmt.Errorf("Error in getting Transaction with Id %s %s", Id, err.Error())
	}
	if assetAsBytes == nil {
		return Transaction{}, fmt.Errorf("Transaction with Id %s does not exist", Id)
	}
	var asset Transaction
	unmarshalError := json.Unmarshal(assetAsBytes, &asset)
	if unmarshalError != nil {
		return Transaction{}, fmt.Errorf("Error in fetching transaction asset: marshalling error %s", unmarshalError.Error())
	}
	if asset.AssetType != trx_asset_type {
		return Transaction{}, fmt.Errorf("Asset with Id %s is not of Transaction DataType", Id)
	}
	return asset, nil
}

func (t *TransactionReciever) getTokenFromBytes(assetAsBytes []byte, result interface{}, Id string) (interface{}, error) {

	var genericResult interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &genericResult)
	if unmarshalError != nil {
		return nil, fmt.Errorf("Error in getting: marshalling error %s", unmarshalError.Error())
	}
	assetTypeFromLedgerString := genericResult.(map[string]interface{})["AssetType"]
	tokenNameFromLedger := genericResult.(map[string]interface{})["Token_name"].(string)
	inputAssetTypeString := reflect.ValueOf(result).Elem().Type().String()
	inputAssetType := strings.Split(inputAssetTypeString, ".")[1]
	if !strings.EqualFold("otoken", assetTypeFromLedgerString.(string)) || !strings.EqualFold(tokenNameFromLedger, inputAssetType) {
		return nil, fmt.Errorf("No asset %s exists with id %s", inputAssetType, Id)
	}
	unmarshalError = json.Unmarshal(assetAsBytes, result)
	if unmarshalError != nil {
		return nil, fmt.Errorf("Error in getting: marshalling error %s", unmarshalError.Error())
	}
	return result, nil
}

func (trx *TransactionReciever) getTokenById(Id string, result ...interface{}) (interface{}, error) {
	stub := trx.model.GetNetworkStub()

	assetAsBytes, err := stub.GetState(Id)
	if err != nil {
		return nil, fmt.Errorf("Error in getting Token with Id %s %s", Id, err.Error())
	}
	if assetAsBytes == nil {
		return nil, fmt.Errorf("Error in getting: Token with Id %s does not exist", Id)
	}

	if len(result) > 0 {
		return trx.getTokenFromBytes(assetAsBytes, result[0], Id)
	}

	var asset interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &asset)
	if unmarshalError != nil {
		return nil, fmt.Errorf("Error in getting: marshalling error %s", unmarshalError.Error())
	}

	return asset, nil
}

func (trx *TransactionReciever) GetTransactionDetails(trx_id string) (Transaction, error) {
	if trx_id == "" {
		return Transaction{}, fmt.Errorf("Error in retrieving Transaction, Transaction id is empty")
	}
	transactionAsset, err := trx.get(trx_id)
	if err != nil {
		return Transaction{}, fmt.Errorf("Error in getting Transaction %s", err.Error())
	}
	return transactionAsset, nil
}

func (trx *TransactionReciever) GetTokenDecimals(token_id string) (int, error) {
	token, err := trx.getTokenById(token_id)
	if err != nil {
		return 0, err
	}

	tokenMap := token.(map[string]interface{})
	BehaviorValue := tokenMap["Behavior"].([]interface{})
	var behaviors []string
	for i := range BehaviorValue {
		behaviors = append(behaviors, BehaviorValue[i].(string))
	}

	if util.FindInStringSlice(behaviors, "divisible") {
		val, ok := tokenMap["Divisible"]
		if ok && val != nil {
			mintable_properties := val.(map[string]interface{})
			val, ok = mintable_properties["Decimal"]
			if !ok {
				return 0, nil
			}
			return int(val.(float64)), nil
		}
		return 0, nil
	} else {
		return 0, nil
	}
}

// TODO: validations on Amount, TransactionType
func (trx *TransactionReciever) validateTransactionData(asset interface{}) error {
	assetAsBytes, errMarshal := json.Marshal(asset)
	if errMarshal != nil {
		return fmt.Errorf("Error in validating Transaction asset: marshal error %s", errMarshal.Error())
	}
	var transactionObject map[string]interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &transactionObject)
	if unmarshalError != nil {
		return fmt.Errorf("Error in validating Transaction asset: unmarshalling error %s", unmarshalError.Error())
	}
	val, ok := transactionObject["FromAccountId"]
	if !ok {
		return fmt.Errorf("Error in validating Transaction asset: FromAccountId is missing")
	} else if _, ok = val.(string); !ok {
		return fmt.Errorf("Error in validating Transaction asset: FromAccountId is not of type string")
	}

	val, ok = transactionObject["ToAccountId"]
	if !ok {
		return fmt.Errorf("Error in validating Transaction asset: ToAccountId is missing")
	} else if _, ok = val.(string); !ok {
		return fmt.Errorf("Error in validating Transaction asset: ToAccountId is not of type string")
	}

	subTrx, ok := transactionObject["NumberOfSubTransactions"]
	if !ok {
		return fmt.Errorf("Error in validating Transaction asset: NumberOfSubTransactions is missing from Transaction data")
	}
	numSubTrx := subTrx.(float64)
	if numSubTrx < 0 {
		return fmt.Errorf("Error in validating Transaction asset: NumberOfSubTransactions %v cannot be less than 0", numSubTrx)
	}

	timestamp := transactionObject["Timestamp"]
	_, err := time.Parse(time.RFC3339, timestamp.(string))
	if err != nil {
		return fmt.Errorf("Error in validating Transaction asset: Unable to parse timestamp: %s, time is in wrong format. Expected format is RFC3339 for example 2006-01-02T15:04:05Z07:00 ", timestamp)
	}

	// code is copied from tokenModel.go. TODO: refactor into a separate function
	// validate the decimal places in amount
	tokenDecimal, err := trx.GetTokenDecimals(transactionObject["TokenId"].(string))
	if err != nil {
		return err
	}

	if tokenDecimal < util.GetDecimals(transactionObject["Amount"].(float64)) {
		return fmt.Errorf("Error in validating Transaction asset: Amount has greater number of decimal places than maximum decimal places: %v for token_id: %s", tokenDecimal, transactionObject["TokenId"])
	}

	return nil
}

func (trx *TransactionReciever) Update(asset interface{}) (interface{}, error) {
	err := trx.validateTransactionData(asset)
	if err != nil {
		return nil, fmt.Errorf("Error in validating Transaction Data %s", err.Error())
	}
	err = util.SetField(asset, "TransactionId", "")
	if err != nil {
		return nil, err
	}
	return trx.model.Update(asset)
}

func (trx *TransactionReciever) History(trx_id string) ([]interface{}, error) {
	if trx_id == "" {
		return nil, fmt.Errorf("Error in retrieving transaction, trx_id cannot be empty")
	}
	_, err := trx.get(trx_id)
	if err != nil {
		return nil, fmt.Errorf("Error in getting Transaction %s", err.Error())
	}
	return trx.model.GetHistoryById(trx_id)
}

func (trx *TransactionReciever) GetByRange(start_trx_id string, end_trx_id string) (bytes.Buffer, error) {
	return trx.model.GetByRangeFromLedger(start_trx_id, end_trx_id)
}

func (trx *TransactionReciever) Delete(trx_id string) (interface{}, error) {
	// validate if a trasaction with trx_id exists
	_, err := trx.get(trx_id)
	if err != nil {
		return nil, err
	}

	return trx.model.Delete(trx_id)
}

func (trx *TransactionReciever) GetTransactionsHistory(trx_id string) (interface{}, error) {
	if trx_id == "" {
		return nil, fmt.Errorf("Transaction id cannot be empty %s", trx_id)
	}
	transaction, err := trx.GetTransactionDetails(trx_id)
	if err != nil {
		return nil, fmt.Errorf("Error in getting transaction with id %s, %s", trx_id, err.Error())
	}
	trx_history, err := trx.History(trx_id)
	if err != nil {
		return nil, fmt.Errorf("Error in getting history for trx_id %s %s", trx_id, err.Error())
	}

	response := make(map[string]interface{})
	response["transaction_id"] = trx_id
	response["history"] = trx_history

	if transaction.AssetType == "BULKTRANSFER" && transaction.NumberOfSubTransactions >= 0 {
		var subTransactionArray []map[string]interface{}
		for i := 1; i <= int(transaction.NumberOfSubTransactions); i++ {
			subTransactionId := trx_id + "~" + util.Getmd5Hash(strconv.Itoa(int(transaction.NumberOfSubTransactions)))
			subtransactionHistory, _ := trx.History(subTransactionId)
			subTransactiondata := make(map[string]interface{})
			subTransactiondata["transaction_id"] = subTransactionId
			subTransactiondata["history"] = subtransactionHistory
			subTransactionArray = append(subTransactionArray, subTransactiondata)
		}
		response["sub_transactions"] = subTransactionArray
	}
	return response, nil
}

func (trx *TransactionReciever) DeleteHistoricalTransactions(referenceTime string) (interface{}, error) {
	_, err := time.Parse(time.RFC3339, referenceTime)
	if err != nil {
		return nil, err
	}

	query := "SELECT key, valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType') = 'otransaction' AND DATETIME(json_extract(valueJson, '$.Timestamp')) < DATETIME('" + referenceTime + "')"
	results, err := trx.model.Query(query)

	if err != nil {
		return nil, fmt.Errorf("Error in executing the SQL query: %s", err)
	}

	transaction_ids := make([]string, 0)

	for _, elem := range results {
		var currTrxAsset Transaction

		data, ok := elem.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Error in typecasting to query record to map[string]interface{}")
		}

		err = json.Unmarshal([]byte(data["valueJson"].(string)), &currTrxAsset)
		if err != nil {
			return nil, fmt.Errorf("Error in typecasting to query record to Transaction asset: %s", err)
		}

		transaction_ids = append(transaction_ids, currTrxAsset.TransactionId)
		trx.Delete(currTrxAsset.TransactionId)
	}

	result := make(map[string]interface{})

	if len(transaction_ids) == 0 {
		result["msg"] = `No transaction older than date : ` + referenceTime + ` is available.`
		result["transactions"] = make([]string, 0)
		return result, nil
	}

	result["msg"] = `Successfuly deleted transaction older than date:` + referenceTime
	result["transactions"] = transaction_ids
	return result, nil
}
