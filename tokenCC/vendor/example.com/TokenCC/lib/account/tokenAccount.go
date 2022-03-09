package account

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
	"strings"

	"example.com/TokenCC/lib/model"
	"example.com/TokenCC/lib/util"
	"example.com/TokenCC/lib/holding"
	"example.com/TokenCC/lib/transaction"
)

type Account struct {
	AssetType     string  `json:"AssetType" final:"oaccount"`
	AccountId     string  `json:"AccountId" id:"true" mandatory:"true"`
	UserId        string  `json:"UserId"`
	OrgId         string  `json:"OrgId"`
	TokenId       string  `json:"TokenId"`
	TokenName      string  `json:"TokenName"`
	TokenSymbol   string  `json:"TokenSymbol"`
	Balance       float64 `json:"Balance"`
	BalanceOnHold float64 `json:"BalanceOnHold"`
	PublicKeystore   map[string]ECParameters   `json:"keystore"`
}

type ECParameters struct {
	CurveName     string `json:"curveName, omitempty"`
	CurveType     string `json:"curveType, omitempty"`
	PX            string `json:"pX, omitempty"`
	PY            string `json:"pY, omitempty"`
}



type AccountReciever struct {
	model       *model.Model
	transaction *transaction.TransactionReciever
}

func GetNewAccountReciever(m *model.Model, trx *transaction.TransactionReciever) *AccountReciever {
	var a AccountReciever
	a.model = m
	a.transaction = trx
	return &a
}

const user_asset_type = "oaccount"

func (a *AccountReciever) getTokenName(token_id string) (string, error) {
	if token_id == "" {
		return "", fmt.Errorf("unable to generate account_id since token_id is empty")
	}

	tokenAsset, err := a.model.Get(token_id)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve the token object for token-id: %s", token_id)
	}

	if tokenAsset.(map[string]interface{})["AssetType"].(string) == "otoken" {
		return tokenAsset.(map[string]interface{})["Token_name"].(string), nil
	}

	return "", fmt.Errorf("No token asset exists with token-id: %s", token_id)
}

func (a *AccountReciever) getTokenSymbol(token_id string) (string, error) {
	if token_id == "" {
		return "", fmt.Errorf("unable to generate account_id since token_id is empty")
	}

	tokenAsset, err := a.model.Get(token_id)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve the token object for token-id: %s", token_id)
	}

	if tokenAsset.(map[string]interface{})["AssetType"].(string) == "otoken" {
		return tokenAsset.(map[string]interface{})["Token_symbol"].(string), nil
	}

	return "", fmt.Errorf("No token asset exists with token-id: %s", token_id)
}

func (a *AccountReciever) GenerateAccountId(token_id string, org_id string, user_id string) (string, error) {
	if token_id == "" || org_id == "" || user_id == "" {
		return "", fmt.Errorf("unable to generate account_id either token_id or org_id or user_id is empty")
	}

	err := util.ValidateOrgAndUser(org_id, user_id)
	if err != nil {
		return "", err
	}

	tokenName, err := a.getTokenName(token_id)
	if err != nil {
		return "", err
	}

	compiledString := fmt.Sprintf("%s~%s~%s", token_id, org_id, user_id)
	h := sha256.New()
	h.Write([]byte(compiledString))
	k := h.Sum(nil)
	account_id := fmt.Sprintf("%s~%s~%x", util.TokenIdPrefix, tokenName, k)
	return account_id, nil
}

func (a *AccountReciever) CreateAccount(token_id string, org_id string, user_id string, alias string, ecParams string) (Account, error) {
	if token_id == "" {
		return Account{}, fmt.Errorf("token_id cannot be empty")
	}
	err := util.ValidateOrgAndUser(org_id, user_id)
	if err != nil {
		return Account{}, err
	}
	err = a.model.IsTokenType(token_id)
	if err != nil {
		return Account{}, err
	}
	var accountAsset Account
	accountAsset.OrgId = org_id
	accountAsset.UserId = user_id
	accountAsset.TokenId = token_id

	token_symbol, err := a.getTokenSymbol(token_id)
    if err != nil {
        return Account{}, err
    }

	accountAsset.TokenSymbol = token_symbol

	tokenName, err := a.getTokenName(token_id)
	if err != nil {
		return Account{}, err
	}
	accountAsset.TokenName = tokenName

	account_id, err := a.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return Account{}, fmt.Errorf("error in generating account_id %s", err.Error())
	}
	accountAsset.AccountId = account_id

	var pubKeyParams ECParameters
		
	ecBytes := []byte(ecParams)

	err = json.Unmarshal(ecBytes, &pubKeyParams)
	if err != nil{
		return Account{}, fmt.Errorf("Error Reading ECParams JSON in %s", err.Error())
	}

	var keyStoreMap = make(map[string] ECParameters)
	keyStoreMap[alias] = pubKeyParams
	accountAsset.PublicKeystore = keyStoreMap

	err = a.validateAccountData(&accountAsset)
	if err != nil {
		return Account{}, fmt.Errorf("error in validating Account, %s", err.Error())
	}
	_, err = a.model.Save(&accountAsset)
	if err != nil {
		return Account{}, fmt.Errorf("error in saving account with Account_Id %s (Org_Id: %s, User_Id: %s) for token_id %s, Error: %s", account_id, org_id, user_id, token_id, err.Error())
	}
	return accountAsset, nil
}

func (a *AccountReciever) GetUserPubKey() (ECParameters, error) {
	pub, err := a.model.GetUserPublicKey()
	if err != nil {
		return ECParameters{}, fmt.Errorf("unable to get Public Key %s", err.Error())
	}

	pxy := strings.Split(pub, ":") 
	fullPub := ECParameters{CurveName: "SECP", CurveType: "P256r1", PX: pxy[0], PY: pxy[1]}
	return fullPub, nil
}

func (a *AccountReciever) get(Id string) (Account, error) {
	stub := a.model.GetNetworkStub()

	assetAsBytes, err := stub.GetState(Id)
	if err != nil {
		return Account{}, fmt.Errorf("error in getting Account with Id %s %s", Id, err.Error())
	}
	if assetAsBytes == nil {
		return Account{}, fmt.Errorf("Account with Id %s does not exist", Id)
	}

	var asset Account
	unmarshalError := json.Unmarshal(assetAsBytes, &asset)
	if unmarshalError != nil {
		return Account{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
	}
	if asset.AssetType != user_asset_type {
		return Account{}, fmt.Errorf("Account with Id %s does not exist", Id)
	}
	return asset, nil
}

func (a *AccountReciever) GetAccount(account_id string) (Account, error) {
	if account_id == "" {
		return Account{}, fmt.Errorf("error in retrieving account, account id is empty")
	}
	accountAsset, err := a.get(account_id)
	if err != nil {
		return Account{}, fmt.Errorf("error in getting account %s", err.Error())
	}
	return accountAsset, nil
}

func (a *AccountReciever) validateAccountData(asset interface{}) error {
	assetAsBytes, errMarshal := json.Marshal(asset)
	if errMarshal != nil {
		return fmt.Errorf("marshal error %s", errMarshal.Error())
	}
	var accountObject map[string]interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &accountObject)
	if unmarshalError != nil {
		return fmt.Errorf("unmarshalling error %s", unmarshalError.Error())
	}
	val, ok := accountObject["Balance"]
	if !ok {
		return fmt.Errorf("balance is missing")
	}
	balance, ok := val.(float64)
	if !ok {
		return fmt.Errorf("balance is not of type float64")
	}
	if balance < 0 {
		return fmt.Errorf("balance %v cannot be less than 0", balance)
	}
	holdval, ok := accountObject["BalanceOnHold"]
	if !ok {
		return fmt.Errorf("on hold balance is missing from account data")
	}
	holdbalance, ok := holdval.(float64)
	if !ok {
		return fmt.Errorf("on hold balance is not of type float64")
	}
	if holdbalance < 0 {
		return fmt.Errorf("on hold balance %v cannot be less than 0", holdbalance)
	}
	return nil
}

func (a *AccountReciever) Update(asset interface{}) (interface{}, error) {
	err := a.validateAccountData(asset)
	if err != nil {
		return nil, fmt.Errorf("error in validating Account Data %s", err.Error())
	}
	return a.model.Update(asset)
}

func (a *AccountReciever) History(account_id string) ([]interface{}, error) {
	if account_id == "" {
		return nil, fmt.Errorf("error in retrieving account, account id caanot be empty")
	}
	_, err := a.get(account_id)
	if err != nil {
		return nil, fmt.Errorf("error in getting account %s", err.Error())
	}
	return a.model.GetHistoryById(account_id)
}

func (a *AccountReciever) GetUserByAccountById(account_id string) (interface{}, error) {
	accountAsset, err := a.GetAccount(account_id)
	if err != nil {
		return nil, fmt.Errorf("error in getting account with id %s", account_id)
	}

	result := make(map[string]interface{})

	result["token_id"] = accountAsset.TokenId
	result["user_id"] = accountAsset.UserId
	result["org_id"] = accountAsset.OrgId
	return result, nil
}

func (a *AccountReciever) GetAccountTransactionHistory(account_id string) (interface{}, error) {
	if account_id == "" {
		return nil, fmt.Errorf("error in retrieving account, account id is empty")
	}
	_, err := a.GetAccount(account_id)
	if err != nil {
		return nil, err
	}

	userHistoryArray, err := a.History(account_id)
	if err != nil {
		return nil, fmt.Errorf("error in getting history for account_id %s err %s", account_id, err.Error())
	}

	fmt.Println("transaction history : ", userHistoryArray)

	var transactionHistoryArray []map[string]interface{}

	for i := 0; i < len(userHistoryArray); i++ {
		historyItem := userHistoryArray[i].(map[string]interface{})
		trx_id := historyItem["TxId"].(string)
		transactionId := "otransaction~" + trx_id
		transaction_data, err := a.transaction.GetTransactionDetails(transactionId)
		if err != nil {
			fmt.Println("error in getting transaction for transaction_id", transactionId)
			break
		}

		transaction_result := make(map[string]interface{})
		transaction_result["transaction_id"] = transaction_data.TransactionId
		transaction_result["transacted_amount"] = transaction_data.Amount
		transaction_result["timestamp"], _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", historyItem["Timestamp"].(string))
		transaction_result["token_id"] = transaction_data.TokenId
		val := historyItem["Value"].(map[string]interface{})
		transaction_result["balance"] = val["Balance"]
		transaction_result["onhold_balance"] = val["BalanceOnHold"]

		switch transaction_data.TransactionType {
		case "MINT":
			transaction_result["transacted_account"] = transaction_data.ToAccountId
			transaction_result["transaction_type"] = "MINT"
			transactionHistoryArray = append(transactionHistoryArray, transaction_result)
		case "BURN":
			transaction_result["transacted_account"] = transaction_data.FromAccountId
			transaction_result["transaction_type"] = "BURN"
			transactionHistoryArray = append(transactionHistoryArray, transaction_result)
		case "ONHOLD":
			transaction_result["transacted_account"] = transaction_data.FromAccountId
			transaction_result["transaction_type"] = "ONHOLD"
			transaction_result["holding_id"] = transaction_data.HoldingId
			transactionHistoryArray = append(transactionHistoryArray, transaction_result)
		case "EXECUTEHOLD":
			if transaction_data.FromAccountId == account_id {
				transaction_result["transacted_account"] = transaction_data.FromAccountId
				transaction_result["transaction_type"] = "EXECUTEHOLD"
			} else if transaction_data.ToAccountId == account_id {
				transaction_result["transaction_account"] = transaction_data.ToAccountId
				transaction_result["transaction_type"] = "CREDIT"
			} else {
				return nil, fmt.Errorf("invalid transaction found for transaction id %s, type %s", transaction_result["transaction_id"], transaction_data.TransactionType)
			}
			transaction_result["holding_id"] = transaction_data.HoldingId
			transactionHistoryArray = append(transactionHistoryArray, transaction_result)
		case "RELEASEHOLD":
			transaction_result["transacted_account"] = transaction_data.FromAccountId
			transaction_result["transaction_type"] = "RELEASEHOLD"
			transaction_result["holding_id"] = transaction_data.HoldingId
			transactionHistoryArray = append(transactionHistoryArray, transaction_result)
		case "TRANSFER":
			if transaction_data.FromAccountId == account_id {
				transaction_result["transacted_account"] = transaction_data.FromAccountId
				transaction_result["transaction_type"] = "DEBIT"
			} else if transaction_data.ToAccountId == account_id {
				transaction_result["transacted_account"] = transaction_data.ToAccountId
				transaction_result["transaction_type"] = "CREDIT"
			} else {
				return nil, fmt.Errorf("invalid transaction found for transaction id %s, type %s", transaction_result["transaction_id"], transaction_data.TransactionType)
			}
			transactionHistoryArray = append(transactionHistoryArray, transaction_result)
		case "BULKTRANSFER":
			fmt.Println("bulkTransfer")
			subTrxResults := make([]map[string]interface{}, 0)
			if transaction_data.NumberOfSubTransactions > 0 {
				for i := 1; i <= int(transaction_data.NumberOfSubTransactions); i++ {
					compiledString := fmt.Sprintf("%d", i)
					h := md5.New()
					h.Write([]byte(compiledString))
					k := h.Sum(nil)
					// compiledStringEncoded := base64.StdEncoding.EncodeToString(k)
					subTransactionId := fmt.Sprintf("%s~%x", transactionId, k)
					fmt.Println("subTransactionId", subTransactionId)
					subTransaction, err := a.transaction.GetTransactionDetails(subTransactionId)
					if err != nil {
						continue
					}
					subTrxObj := make(map[string]interface{})
					if subTransaction.FromAccountId == account_id {
						subTrxObj["transacted_account"] = transaction_data.FromAccountId
						subTrxObj["transaction_type"] = "DEBIT"
						subTrxObj["transaction_id"] = subTransaction.TransactionId
						subTrxObj["transacted_amount"] = subTransaction.Amount
						subTrxResults = append(subTrxResults, subTrxObj)
					} else if subTransaction.ToAccountId == account_id {
						subTrxObj["transacted_account"] = transaction_data.ToAccountId
						subTrxObj["transaction_type"] = "CREDIT"
						subTrxObj["transaction_id"] = subTransaction.TransactionId
						subTrxObj["transacted_amount"] = subTransaction.Amount
						subTrxResults = append(subTrxResults, subTrxObj)
					}
				}
				transaction_result["transaction_type"] = "BULKTRANSFER"
				transaction_result["sub_transactions"] = subTrxResults
				transactionHistoryArray = append(transactionHistoryArray, transaction_result)
			}
		default:
			fmt.Println("transaction type not recognized ", transaction_data.TransactionType)
			continue
		}
	}
	return transactionHistoryArray, nil
}

func (a *AccountReciever) GetAccountBalance(account_id string) (map[string]interface{}, error) {
	if account_id == "" {
		return nil, fmt.Errorf("error in retrieving account, account id is empty")
	}
	userAsset, err := a.GetAccount(account_id)
	if err != nil {
		return nil, err
	}
	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Current Balance of %s is %v", account_id, userAsset.Balance)
	response["user_balance"] = userAsset.Balance
	return response, nil
}

func (a *AccountReciever) GetOnHoldIDs(account_id string) (map[string]interface{}, error) {
	if account_id == "" {
		return nil, fmt.Errorf("error in retrieving account, account id is empty")
	}
	_, err := a.GetAccount(account_id)
	if err != nil {
		return nil, err
	}

	holdStartId := util.TokenHoldRangeStartKey
	holdEndId := util.TokenHoldRangeStartKey + util.EndKeyChar

	var holds []holding.Hold

	holdistBuffer, err := a.model.GetByRangeFromLedger(holdStartId, holdEndId)
	if err != nil {
		return nil, fmt.Errorf("error in getting admin list from ledger %s", err.Error())
	}
	holdListMap, err := util.FilterRangeResultsByAssetType(holding.Hold_asset_type, holdistBuffer)
	if err != nil {
		return nil, fmt.Errorf("error in converting range results to admin list %s", err.Error())
	}
	mapBytes, err := json.Marshal(holdListMap)
	if err != nil {
		return nil, fmt.Errorf("error in marshalling map %s", err.Error())
	}
	err = json.Unmarshal(mapBytes, &holds)
	if err != nil {
		return nil, fmt.Errorf("error in unmarshalling map %s", err.Error())
	}
	var holdingIds []string
	for _, element := range holds {
		// do not return completely executed OR released holds
		// only return the hold assets relevant to given acconut_id
		if element.Quantity > 0 && element.FromAccountId == account_id {
			holdingIds = append(holdingIds, element.HoldingId)
		}
	}

	response := make(map[string]interface{})
	response["holding_ids"] = holdingIds
	response["msg"] = fmt.Sprintf("Holding Ids are: %v", holdingIds)
	return response, nil
}

func (a *AccountReciever) GetAccountOnHoldBalance(account_id string) (map[string]interface{}, error) {
	if account_id == "" {
		return nil, fmt.Errorf("error in retrieving account, account id is empty")
	}
	userAsset, err := a.GetAccount(account_id)
	if err != nil {
		return nil, err
	}
	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Total Holding Balance of Account Id %s (org_id: %s, user_id: %s) is %v", account_id, userAsset.OrgId, userAsset.UserId, userAsset.BalanceOnHold)
	response["holding_balance"] = userAsset.BalanceOnHold
	return response, nil
}

func (a *AccountReciever) GetAllAccounts() (interface{}, error) {
	query := `{"selector": {"AssetType": "oaccount"}}`
	return a.model.Query(query)
}

func (a *AccountReciever) GetAccountsByUser(org_id string, user_id string) (interface{}, error) {
    query := fmt.Sprintf(`{"selector": {"AssetType": "oaccount","OrgId":"%s","UserId":"%s"}}`, org_id, user_id)
	return a.model.Query(query)
}
