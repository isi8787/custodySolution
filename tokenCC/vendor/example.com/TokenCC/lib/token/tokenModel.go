package token

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"example.com/TokenCC/lib/account"
	"example.com/TokenCC/lib/holding"
	"example.com/TokenCC/lib/model"
	tokenRole "example.com/TokenCC/lib/role"
	"example.com/TokenCC/lib/util"
	"example.com/TokenCC/lib/transaction"
)

const ochain_token_type_string = "otoken"
const metadata_asset_type = "ometadata"

type TokenReciever struct {
	model       *model.Model
	tokenRole   *tokenRole.RoleReciever
	holding     *holding.HoldReciever
	account     *account.AccountReciever
	transaction *transaction.TransactionReciever
}

type BasicToken struct {
	AssetType  string   `json:"AssetType" final:"otoken"`
	Token_id   string   `json:"Token_id" id:"true" mandatory:"true" validate:"regexp=^[A-Za-z0-9][A-Za-z0-9_-]*$,max=16"`
	Token_name string   `json:"Token_name"`
	Token_symbol string   `json:"Token_symbol"`
	Token_desc string   `json:"Token_desc" validate:"max=256"`
	Token_type string   `json:"Token_type" final:"fungible" validate:"regexp=^fungible$"`
	Behavior   []string `json:"Behavior" final:"[\"divisible\",\"mintable\",\"transferable\",\"burnable\",\"roles\"]"`

	Roles map[string]interface{} `json:"Roles" final:"{\"minter_role_name\":\"minter\"}"`

	Mintable map[string]interface{} `json:"Mintable" final:"{\"Max_mint_quantity\":100000000000}"`

	Divisible map[string]interface{} `json:"Divisible" final:"{\"Decimal\":8}"`

	Currency_name           string      `json:"Currency_name" validate:"string"`
	Token_to_currency_ratio int         `json:"Token_to_currency_ratio" validate:"int"`
	Metadata                interface{} `json:"Metadata,omitempty"`
}

func GetNewTokenReciever(m *model.Model, t *tokenRole.RoleReciever, h *holding.HoldReciever, a *account.AccountReciever, trx *transaction.TransactionReciever) *TokenReciever {
	var tokenReciever TokenReciever
	tokenReciever.model = m
	tokenReciever.account = a
	tokenReciever.tokenRole = t
	tokenReciever.holding = h
	tokenReciever.transaction = trx
	return &tokenReciever
}

type TokenMetadata struct {
	AssetType           string  `json:"AssetType" final:"ometadata"`
	Metadata_id         string  `json:"Metadata_id" id:"true" derived:"strategy=concat,format=%1~%2~%3,AssetType,Token_name,Token_id"`
	Token_id            string  `json:"Token_id"`
	Token_name			string	`json:"Token_name"`
	Total_supply        float64 `json:"Total_supply"`
	Total_minted_amount float64 `json:"Total_minted_amount"`
}

func (t *TokenReciever) formatNumber(number float64, precision int) string {
	formatString := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(formatString, number)
}

func (t *TokenReciever) increment(a float64, b float64, limit int) (float64, error) {
	numA, err := strconv.ParseFloat(t.formatNumber(a, limit), 64)
	if err != nil {
		return 0, fmt.Errorf("error in incrmenting,converting to number with precision %f", a)
	}
	numB, err := strconv.ParseFloat(t.formatNumber(b, limit), 64)
	if err != nil {
		return 0, fmt.Errorf("error in incrementing, converting to number with precision %f", b)
	}
	fmt.Println(numA, numB)
	numC, err := strconv.ParseFloat(t.formatNumber(numA+numB, limit), 64)
	if err != nil {
		return 0, fmt.Errorf("error in incrementing, converting to number to precision %f", b)
	}
	return numC, nil
}

func (t *TokenReciever) decrement(a float64, b float64, limit int) (float64, error) {
	numA, err := strconv.ParseFloat(t.formatNumber(a, limit), 64)
	if err != nil {
		return 0, fmt.Errorf("error in decrementing, converting to number to precision %f %s", a, err.Error())
	}
	numB, err := strconv.ParseFloat(t.formatNumber(b, limit), 64)
	if err != nil {
		return 0, fmt.Errorf("error in decrementing, converting to number to precision %f %s", b, err.Error())
	}
	fmt.Println(numA, numB)
	numC, err := strconv.ParseFloat(t.formatNumber(numA-numB, limit), 64)
	if err != nil {
		return 0, fmt.Errorf("error in decrementing, converting to number to precision %f %s", b, err.Error())
	}
	return numC, nil
}

func (t *TokenReciever) getTokenMetadataFromLedgerById(id string, obj interface{}) (interface{}, error) {
	stub := t.model.GetNetworkStub()
	assetAsBytes, err := stub.GetState(id)
	if err != nil {
		return nil, fmt.Errorf("error in getting tokenMetadata %s", err.Error())
	}
	if assetAsBytes == nil {
		return nil, fmt.Errorf("error in getting tokenMetadata asset with id %s does not exist", id)
	}

	_, err = model.ConvertAssetFromBytes(assetAsBytes, obj, metadata_asset_type)
	if err != nil {
		return nil, fmt.Errorf("token metadata with id %s does not exist %s", id, err.Error())
	}
	return obj, nil
}

func (t *TokenReciever) RoleCheck(account_id string, tokenAsset interface{}) (bool, error) {
	if account_id == "" {
		return false, fmt.Errorf("account id must be a non-empty string")
	}
	_, err := t.account.GetAccount(account_id)
	if err != nil {
		return false, fmt.Errorf("error in getting account with account id %s %s", account_id, err.Error())
	}
	ok, err := t.checkMinterRole(account_id, tokenAsset)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	ok, err = t.checkBurnAllowed(account_id, tokenAsset)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	ok, err = t.checkNotary(account_id, tokenAsset)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	return false, nil
}

func (t *TokenReciever) createTokenMetadata(token_id string, token_name string) (interface{}, error) {
	var tokenMetadata TokenMetadata

	tokenMetadata.Token_id = token_id
	tokenMetadata.Token_name = token_name
	tokenMetadata.Total_minted_amount = 0
	tokenMetadata.Total_supply = 0
	return t.model.Save(&tokenMetadata)
}

func (t *TokenReciever) getTokenMetadata(token_id string, token_name string) (TokenMetadata, error) {
	var asset TokenMetadata
	asset.Token_id = token_id
	asset.AssetType = metadata_asset_type
	asset.Token_name = token_name

	id, err := t.model.GetId(&asset)
	if err != nil {
		return TokenMetadata{}, fmt.Errorf("error in generating the id for tokenMetadata %s", err.Error())
	}

	_, err = t.getTokenMetadataFromLedgerById(id, &asset)
	if err != nil {
		return TokenMetadata{}, fmt.Errorf("error in getting metadata with id %s %s", id, err.Error())
	}

	return asset, nil
}

func (t *TokenReciever) updateTokenMetadata(asset TokenMetadata) (interface{}, error) {
	asset.Metadata_id = ""
	return t.model.Update(&asset)
}

func (t *TokenReciever) checkBehaviors(tokenid string, behavior string) (bool, error) {
	tokenAsset, err := t.model.Get(tokenid)

	if err != nil {
		return false, err
	}

	tokenAssetMap := tokenAsset.(map[string]interface{})

	behaviorsInterface := tokenAssetMap["Behavior"].([]interface{})

	var behaviors []string

	for i := 0; i < len(behaviorsInterface); i++ {
		behaviors = append(behaviors, behaviorsInterface[i].(string))
	}

	return util.FindInStringSlice(behaviors, behavior), nil
}

func (t *TokenReciever) getTokenFromBytes(assetAsBytes []byte, result interface{}, Id string) (interface{}, error) {

	var genericResult interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &genericResult)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting: marshalling error %s", unmarshalError.Error())
	}

	unmarshalError = json.Unmarshal(assetAsBytes, result)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting: marshalling error %s", unmarshalError.Error())
	}
	return result, nil
}

func (t *TokenReciever) validateTokenProperties(asset map[string]interface{}) error {

	_, ok := asset["Token_id"]
	if !ok {
		return fmt.Errorf("error in saving: TokenId is not present")
	}
	if len(asset["Token_id"].(string)) == 0 {
		return fmt.Errorf("error in saving: TokenId cannot be empty %s ", asset["Token_id"].(string))
	}
	if len(asset["Token_id"].(string)) > 16 {
		return fmt.Errorf("error in saving: Length of Token_id cannot be greater than 16 characters %s ", asset["Token_id"].(string))
	}
	expression, err := regexp.Compile(util.RegexForTokenId)
	if err != nil {
		return fmt.Errorf("validating ids failed, regex could not be compiled %s", err.Error())
	}
	if !expression.MatchString(asset["Token_id"].(string)) {
		return fmt.Errorf("token id %s is not in accepted format, it should start with alphanumeric and can include '-' and '_'", asset["Token_id"].(string))
	}
	_, ok = asset["Token_desc"]
	if !ok {
		return fmt.Errorf("error in saving: TokenDesc is not present")
	}
	if len(asset["Token_desc"].(string)) > 256 {
		return fmt.Errorf("error in saving: TokenDesc must be less than 256 characters %s ", asset["Token_desc"].(string))
	}

	_, ok = asset["Token_type"]
	if !ok {
		return fmt.Errorf("error in saving: TokenType is not present")
	}
	if asset["Token_type"] != "fungible" {
		return fmt.Errorf("error in saving: TokenType must be fungible")
	}

	_, ok = asset["Behavior"]
	if !ok {
		return fmt.Errorf("error in saving: Behaviors properties missing from the token")
	}

	BehaviorValue := asset["Behavior"].([]interface{})
	var behaviors []string
	for i := range BehaviorValue {
		behaviors = append(behaviors, BehaviorValue[i].(string))
	}

	if !util.FindInStringSlice(behaviors, "mintable") {
		return fmt.Errorf("error in saving: mintable behavior is mandatory")
	}
	val, ok := asset["Mintable"]
	if ok && val != nil {
		mintable_properties := val.(map[string]interface{})
		val, ok = mintable_properties["Max_mint_quantity"]
		if ok && val.(float64) < 0 {
			return fmt.Errorf("error in saving: MaxMintQuantity must be atleast 0")
		}
	}

	if !util.FindInStringSlice(behaviors, "transferable") {
		return fmt.Errorf("error in saving: transferable behavior is mandatory")
	}
	if util.FindInStringSlice(behaviors, "divisible") {
		val, ok := asset["Divisible"]
		if ok && val != nil {
			mintable_properties := val.(map[string]interface{})
			val, ok = mintable_properties["Decimal"]
			if ok {
				if val.(float64) < 0 || val.(float64) > 8 {
					return fmt.Errorf("error in saving: Property DecimalPlaces must be between 0 and 8")
				}
				if util.GetDecimals(val.(float64)) != 0 {
					return fmt.Errorf("error in saving: Property DecimalPlaces must be an integer %v", val.(float64))
				}
			}
		}
	}
	return nil
}

func (t *TokenReciever) Save(args ...interface{}) (interface{}, error) {
	stub := t.model.GetNetworkStub()
	obj := args[0]
	_, err := util.SetFinaTagInput(obj)
	if err != nil {
		return nil, fmt.Errorf("setting final input for token failed %s", err.Error())
	}

	id, idErr := t.model.GenerateID(obj, false, "save")
	if idErr != nil {
		return nil, fmt.Errorf("error in getting Id. %s", idErr.Error())
	}

	_, err = t.Get(id)
	if err == nil {
		return nil, fmt.Errorf("error in saving: asset already exist in ledger with Id %s ", id)
	}

	assetAsBytes, errMarshal := json.Marshal(obj)
	if errMarshal != nil {
		return nil, fmt.Errorf("error in saving: Asset Id %s marshal error %s", id, errMarshal.Error())
	}

	var token interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &token)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in saving token: marshalling error %s", unmarshalError.Error())
	}

	asset := token.(map[string]interface{})
	if err := t.validateTokenProperties(asset); err != nil {
		return nil, err
	}

	errPut := stub.PutState(id, assetAsBytes)
	if errPut != nil {
		return nil, fmt.Errorf("error in saving: Asset Id %s transaction error %s", id, errPut.Error())
	}

	_, err = t.createTokenMetadata(id, asset["Token_name"].(string))
	if err != nil {
		return nil, fmt.Errorf("creating token metadata failed err %s", err.Error())
	}

	return obj, nil
}

func (t *TokenReciever) Get(Id string, result ...interface{}) (interface{}, error) {
	stub := t.model.GetNetworkStub()

	assetAsBytes, err := stub.GetState(Id)
	if err != nil {
		return nil, fmt.Errorf("error in getting Token with Id %s %s", Id, err.Error())
	}
	if assetAsBytes == nil {
		return nil, fmt.Errorf("error in getting: Token with Id %s does not exist", Id)
	}

	if len(result) > 0 {
		return t.getTokenFromBytes(assetAsBytes, result[0], Id)
	}

	var asset interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &asset)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in getting: marshalling error %s", unmarshalError.Error())
	}

	return asset, nil
}

func (t *TokenReciever) Update(asset interface{}) (interface{}, error) {

	assetAsBytes, errMarshal := json.Marshal(asset)
	if errMarshal != nil {
		return nil, fmt.Errorf("error in updating:  marshaling error %s", errMarshal.Error())
	}

	var token interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &token)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error in updating: unmarshalling error %s", unmarshalError.Error())
	}

	assetInterface := token.(map[string]interface{})
	if err := t.validateTokenProperties(assetInterface); err != nil {
		return nil, err
	}

	return t.model.Update(asset)
}

func filterRangeResultsByAssetTypeAndTokenName(assetType string,  buffer bytes.Buffer) ([]map[string]interface{}, error) {
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

		if assetTypeString == assetType {
			resultAssets = append(resultAssets, mapAsset)
		}
	}
	return resultAssets, nil
}

func (t *TokenReciever) GetAllTokens() (interface{}, error) {
	query := `{"selector": {"AssetType": "otoken"}}`
	return t.model.Query(query)
}

func (t *TokenReciever) getDecimal(value float64) int {
	s := fmt.Sprintf("%v", value)

	stringSplitted := strings.Split(s, ".")
	if len(stringSplitted) == 1 {
		return 0
	}

	return len(stringSplitted[1])
}

func (t *TokenReciever) checkMinterRole(account_id string, tokenAsset interface{}) (bool, error) {
	structValue := reflect.ValueOf(tokenAsset).Elem()

	asset_type := structValue.FieldByName("AssetType").String()
	rolesValue := structValue.FieldByName("Roles")
	if !rolesValue.IsValid() {
		return true, nil
	}
	roles := rolesValue.Interface()

	if asset_type == "" || asset_type != "otoken" {
		return false, fmt.Errorf("input value %v is not of type token", tokenAsset)
	}

	role_name, ok := roles.(map[string]interface{})["minter_role_name"]
	if !ok {
		return true, nil
	}
	check, err := t.IsInRole(role_name.(string), account_id, tokenAsset)

	if err != nil {
		return false, err
	}

	return check, nil
}

func (t *TokenReciever) checkMintQuantity(quantity float64, tokenAsset interface{}) (bool, error) {
	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return false, err
	}
	maxMintQuantity, err := t.GetMaxMintQuantity(token_id)
	if err != nil {
		return false, err
	}
	value, err := t.GetTotalMintedTokens(tokenAsset)
	if err != nil {
		return false, err
	}
	totalMintedQuantity := value["quantity"].(float64)
	if maxMintQuantity != 0 {
		if quantity > maxMintQuantity {
			return false, fmt.Errorf("quantity to mint is greater than the maximum mintable quantity %v for token_id: %s", maxMintQuantity, token_id)
		}
		if quantity+totalMintedQuantity > maxMintQuantity {
			return false, fmt.Errorf("quantity: %v remaining to be minted for token_id: %s", maxMintQuantity-totalMintedQuantity, token_id)
		}
	}
	return true, nil
}

func (t *TokenReciever) checkMintAllowed(account_id string, quantity float64, tokenAsset interface{}) (bool, error) {
	roleCheck, err := t.checkMinterRole(account_id, tokenAsset)
	if err != nil {
		return false, fmt.Errorf("error in checking minter role %s", err.Error())
	}
	if !roleCheck {
		return false, nil
	}
	_, err = t.checkMintQuantity(quantity, tokenAsset)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (t *TokenReciever) checkBurnAllowed(account_id string, tokenAsset interface{}) (bool, error) {

	structValue := reflect.ValueOf(tokenAsset).Elem()

	asset_type := structValue.FieldByName("AssetType").String()
	rolesValue := structValue.FieldByName("Roles")
	if !rolesValue.IsValid() {
		return true, nil
	}

	roles := rolesValue.Interface()

	if asset_type == "" || asset_type != "otoken" {
		return false, fmt.Errorf("input value %v is not of type token", tokenAsset)
	}

	role_name, ok := roles.(map[string]interface{})["burner_role_name"]
	if !ok {
		return true, nil
	}

	check, err := t.IsInRole(role_name.(string), account_id, tokenAsset)

	if err != nil {
		return false, err
	}

	return check, nil
}

func (t *TokenReciever) checkNotary(account_id string, tokenAsset interface{}) (bool, error) {

	structValue := reflect.ValueOf(tokenAsset).Elem()
	asset_type := structValue.FieldByName("AssetType").String()
	rolesValue := structValue.FieldByName("Roles")

	if !rolesValue.IsValid() {
		return true, nil
	}
	roles := rolesValue.Interface()

	if asset_type == "" || asset_type != "otoken" {
		return false, fmt.Errorf("input value %v is not of type token", structValue)
	}

	role_name, ok := roles.(map[string]interface{})["notary_role_name"]
	if !ok {
		return true, nil
	}

	check, err := t.IsInRole(role_name.(string), account_id, tokenAsset)

	if err != nil {
		return false, err
	}

	return check, nil
}

func (t *TokenReciever) Mint(quantity float64, tokenAsset interface{}) (interface{}, error) {
	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}

	userId, err := t.model.GetUserId()
	if err != nil {
		return nil, fmt.Errorf("error in retrieving user_id of the caller %s", err.Error())
	}
	orgId := t.model.GetCreatorMspId()

	account_id, err := t.account.GenerateAccountId(token_id, orgId, userId)
	if err != nil {
		return nil, err
	}

	return t.mintTo(account_id, quantity, tokenAsset)
}

func (t *TokenReciever) mintTo(account_id string, quantity float64, tokenAsset interface{}) (interface{}, error) {

	userAsset, err := t.account.GetAccount(account_id)
	if err != nil {
		userid, _ := t.model.GetUserId()
		return nil, fmt.Errorf("account %s does not exist, please create Account first for the user with org_id %s and user_id %s", account_id, t.model.GetCreatorMspId(), userid)
	}

	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}

	if userAsset.TokenId != token_id {
		return nil, fmt.Errorf("account %s does not hold tokens with id %s", userAsset.AccountId, token_id)
	}

	check, err := t.checkBehaviors(token_id, "mintable")
	if err != nil {
		return "", err
	}
	if !check {
		return "", fmt.Errorf("mintable behavior is not configured for token_id: %s!. Please add this behavior to continue", token_id)
	}

	auth, err := t.checkMintAllowed(account_id, quantity, tokenAsset)
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("caller is not authorized to mint tokens, does not have minter role")
	}

	if quantity <= 0 {
		return "", fmt.Errorf("error in Minting: Quantity must be a positive value")
	}

	tokenDecimal, err := t.GetDecimals(token_id)
	if err != nil {
		return nil, err
	}
	if tokenDecimal < util.GetDecimals(quantity) {
		return "", fmt.Errorf("quantity has greater number of decimal places than maximum decimal places: %v for token_id: %s", tokenDecimal, token_id)
	}

	// userAsset.Balance += quantity

	updatedBalance, err := t.increment(userAsset.Balance, quantity, tokenDecimal)
	if err != nil {
		return nil, fmt.Errorf("Error in updating balance %s", err.Error())
	}
	userAsset.Balance = updatedBalance

	t.account.Update(&userAsset)

	_, err = t.transaction.CreateTransaction(token_id, "", account_id, "MINT", quantity, 0, "")
	if err != nil {
		return nil, fmt.Errorf("error in creating transaction entry %s", err.Error())
	}

	token_name, err := util.GetTokenName(tokenAsset)
	if err != nil {
		return nil, err
	}

	tokenMetadata, err := t.getTokenMetadata(token_id, token_name);
	if err != nil {
		return nil, fmt.Errorf("error in getting token metadata for token %s %s", token_id, err.Error())
	}

	//tokenMetadata.Total_minted_amount += quantity
	updated_Total_minted_amount, err := t.increment(tokenMetadata.Total_minted_amount, quantity, tokenDecimal)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new totalMintedAmount %s", err)
	}
	tokenMetadata.Total_minted_amount = updated_Total_minted_amount

	//tokenMetadata.Total_supply += quantity
	updated_Total_supply, err := t.increment(tokenMetadata.Total_supply, quantity, tokenDecimal)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new totalSupply %s", err)
	}
	tokenMetadata.Total_supply = updated_Total_supply
	_, err = t.updateTokenMetadata(tokenMetadata)
	if err != nil {
		return nil, fmt.Errorf("error in updating the token metadata for token %s %s", token_id, err.Error())
	}

	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Successfully minted %v tokens to account %s (org_id : %s, user_id : %s)", quantity, account_id, userAsset.OrgId, userAsset.UserId)
	return response, nil

}

func (t *TokenReciever) Transfer(to_account_id string, quantity float64, tokenAsset interface{}) (interface{}, error) {
	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}

	userId, err := t.model.GetUserId()
	if err != nil {
		return nil, fmt.Errorf("error in retrieving user_id of the caller %s", err.Error())
	}
	orgId := t.model.GetCreatorMspId()

	account_id, err := t.account.GenerateAccountId(token_id, orgId, userId)
	if err != nil {
		return nil, err
	}

	return t.transferFrom(account_id, to_account_id, quantity, tokenAsset)
}

func (t *TokenReciever) transferFrom(from_account_id string, to_account_id string, quantity float64, inputStruct interface{}) (interface{}, error) {

	if from_account_id == to_account_id {
		return "", fmt.Errorf("unable to initiate transfer. From Account and To Account are same")
	}

	if quantity <= 0 {
		return nil, fmt.Errorf("unable to initiate transfer. Quantity must be greater than 0")
	}

	typ := reflect.TypeOf(inputStruct).Elem()
	name := "Token_id"
	structValue := reflect.ValueOf(inputStruct).Elem()
	_, ok := typ.FieldByName(name)
	if !ok {
		return "", fmt.Errorf("token_id field is missing from the %s asset", typ.Name())
	}
	val := structValue.FieldByName(name)
	tokenid := val.String()
	check, err := t.checkBehaviors(tokenid, "transferable")
	if err != nil {
		return "", err
	}
	if !check {
		return "", fmt.Errorf("transferable behavior is not configured for token_id: %s!. Please add this behavior to continue", tokenid)
	}

	tokenDecimal, err := t.GetDecimals(tokenid)
	if err != nil {
		return nil, err
	}

	if tokenDecimal < util.GetDecimals(quantity) {
		return "", fmt.Errorf("quantity has greater number of decimal places than maximum decimal places: %v for token_id: %s", tokenDecimal, tokenid)
	}

	fromUser, err := t.account.GetAccount(from_account_id)
	if err != nil {
		return "", fmt.Errorf("from-Account id %s (org_id : %s, user_id : %s) for %s is not a valid account! Please account for the user", from_account_id, fromUser.OrgId, fromUser.UserId, tokenid)
	}
	toUser, err := t.account.GetAccount(to_account_id)
	if err != nil {
		return "", fmt.Errorf("to-Account id %s for %s is not a valid account! Please register user first", to_account_id, tokenid)
	}

	if toUser.TokenId != fromUser.TokenId {
		return "", fmt.Errorf("Transferring of tokens of different token_id is not allowed")
	}

	fromUserBalance := fromUser.Balance

	if fromUserBalance == 0 {
		return "", fmt.Errorf("from-Account id %s (org_id : %s, user_id : %s)  has 0 balance for token %s", from_account_id, fromUser.OrgId, fromUser.UserId, fromUser.TokenId)
	}

	if fromUserBalance < quantity {
		return "", fmt.Errorf("from-Account id %s (org_id : %s, user_id : %s) for token %s has insufficient balance for transfer!`", from_account_id, fromUser.OrgId, fromUser.UserId, tokenid)
	}

	updated_Balance, err := t.decrement(fromUser.Balance, quantity, tokenDecimal)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new Balance for from_account %s", err)
	}
	fromUser.Balance = updated_Balance

	t.account.Update(&fromUser)

	//toUser.Balance = toUser.Balance + quantity
	updated_Balance, err = t.increment(toUser.Balance, quantity, tokenDecimal)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new Balance for to_account %s", err)
	}
	toUser.Balance = updated_Balance
	t.account.Update(&toUser)

	_, err = t.transaction.CreateTransaction(tokenid, from_account_id, to_account_id, "TRANSFER", quantity, 0, "")
	if err != nil {
		return nil, fmt.Errorf("error in creating transaction asset %s", err.Error())
	}

	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("successfully transferred %v tokens from account %s (org_id : %s, user_id : %s) to account %s (org_id : %s, user_id : %s)", quantity, from_account_id, fromUser.OrgId, fromUser.UserId, to_account_id, toUser.OrgId, toUser.UserId)

	return response, nil
}

func (t *TokenReciever) Burn(orgId string, userId string, quantity float64, tokenAsset interface{}) (interface{}, error) {

	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}

	account_id, err := t.account.GenerateAccountId(token_id, orgId, userId)
	if err != nil {
		return nil, err
	}

	return t.burnFrom(account_id, quantity, tokenAsset)
}

func (t *TokenReciever) burnFrom(account_id string, quantity float64, tokenAsset interface{}) (interface{}, error) {

	if quantity <= 0 {
		return nil, fmt.Errorf("cannot burn tokens, quantity must be positive number %v", quantity)
	}

	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}

	check, err := t.checkBehaviors(token_id, "burnable")
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, fmt.Errorf("burnable behavior is not configured for token_id: %s!. Please add this behavior to continue", token_id)
	}

	auth, err := t.checkBurnAllowed(account_id, tokenAsset)
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("caller is not authorized to burn tokens, does not have burner role")
	}

	tokenDecimal, err := t.GetDecimals(token_id)
	if err != nil {
		return nil, err
	}

	if tokenDecimal < util.GetDecimals(quantity) {
		return "", fmt.Errorf("quantity has greater number of decimal places than maximum decimal places: %v for token_id: %s", tokenDecimal, token_id)
	}

	if err != nil {
		return "", fmt.Errorf("unable to generate timestamp %s", err.Error())
	}

	userAsset, err := t.account.GetAccount(account_id)
	if err != nil {
		return nil, err
	}
	if userAsset.TokenId != token_id {
		return nil, fmt.Errorf("account %s does not holds tokens of id %s", userAsset.AccountId, token_id)
	}
	if userAsset.Balance == 0 {
		return "", fmt.Errorf("account id %s (org_id : %s, user_id : %s)  has 0 balance for token %s", account_id, userAsset.OrgId, userAsset.UserId, token_id)
	}

	if userAsset.Balance < quantity {
		return nil, fmt.Errorf("account does not have sufficient balance to burn tokens. Account Balance : %v", userAsset.Balance)
	}

	//userAsset.Balance -= quantity
	updated_Balance, err := t.decrement(userAsset.Balance, quantity, tokenDecimal)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new balance for burner account %s", err)
	}
	userAsset.Balance = updated_Balance
	t.account.Update(&userAsset)

	_, err = t.transaction.CreateTransaction(token_id, "", account_id, "BURN", quantity, 0, "")
	if err != nil {
		return nil, fmt.Errorf("error in creating transaction asset %s", err.Error())
	}
	token_name, err := util.GetTokenName(tokenAsset)
	if err != nil {
		return nil, err
	}
	tokenMetadata, err := t.getTokenMetadata(token_id,token_name)
	if err != nil {
		return nil, fmt.Errorf("error in getting token metadata for token %s %s", token_id, err.Error())
	}

	//tokenMetadata.Total_supply -= quantity
	updated_Total_supply, err := t.decrement(tokenMetadata.Total_supply, quantity, tokenDecimal)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new Total_supply after burning tokens %s", err)
	}
	tokenMetadata.Total_supply = updated_Total_supply
	_, err = t.updateTokenMetadata(tokenMetadata)
	if err != nil {
		return nil, fmt.Errorf("error in updating the token metadata for token %s %s", token_id, err.Error())
	}

	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Successfully burned %v tokens from account id: %s (org_id : %s, user_id : %s)", quantity, account_id, userAsset.OrgId, userAsset.UserId)
	return response, nil
}

func (t *TokenReciever) Hold(operation_id string, to_account_id string, notary_account_id string, quantity float64, time_to_expiration string, tokenAsset interface{}) (interface{}, error) {

	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}
	userId, err := t.model.GetUserId()
	if err != nil {
		return nil, fmt.Errorf("error in retrieving user_id of the caller %s", err.Error())
	}
	orgId := t.model.GetCreatorMspId()

	from_account_id, err := t.account.GenerateAccountId(token_id, orgId, userId)
	if err != nil {
		return nil, err
	}

	return t.holdFrom(operation_id, from_account_id, to_account_id, notary_account_id, quantity, time_to_expiration, tokenAsset)
}

func (t *TokenReciever) holdFrom(operation_id string, from_account_id string, to_account_id string, notary_account_id string, quantity float64, time_to_expiration string, inputStruct interface{}) (interface{}, error) {
	if operation_id == "" {
		return nil, fmt.Errorf("operation_id cannot be an empty string")
	}
	typ := reflect.TypeOf(inputStruct).Elem()
	name := "Token_id"
	structValue := reflect.ValueOf(inputStruct).Elem()
	_, ok := typ.FieldByName(name)
	if !ok {
		return "", fmt.Errorf("token_id field is missing from the %s asset", typ.Name())
	}
	val := structValue.FieldByName(name)
	token_id := val.String()

	if from_account_id == notary_account_id {
		return "", fmt.Errorf("From-Account-Id cannot be the same as Notary Account-Id")
	}

	if to_account_id == notary_account_id {
		return "", fmt.Errorf("To-Account-Id cannot be the same as Notary Account-Id")
	}

	result, err := t.checkBehaviors(token_id, "holdable")
	if err != nil {
		return "", err
	}
	if !result {
		return "", fmt.Errorf("holdable behavior is not configured for token_id: %s. Please add this behavior to continue", token_id)
	}

	notaryAccount, err := t.account.GetAccount(notary_account_id)
	if err != nil {
		return nil, fmt.Errorf("no Notary account exist with account_id %s", notary_account_id)
	}

	auth, err := t.checkNotary(notary_account_id, inputStruct)
	if err != nil {
		return nil, err
	}

	if !auth {
		return nil, fmt.Errorf("notary account id %s (org_id : %s, user_id : %s) is not authorized to act as a Notary", notary_account_id, notaryAccount.OrgId, notaryAccount.UserId)
	}

	if from_account_id == to_account_id {
		return "", fmt.Errorf("from-Account id is same as To-account Id")
	}
	if from_account_id == notary_account_id {
		return "", fmt.Errorf("from-Account id cannot be same as Notary-account Id")
	}

	from_account, err := t.account.GetAccount(from_account_id)
	if err != nil {
		return "", fmt.Errorf("error in getting from_account, please create account first, from_account_id %s %s", from_account_id, err.Error())
	}

	to_account, err := t.account.GetAccount(to_account_id)
	if err != nil {
		return "", fmt.Errorf("error in getting to_account, please create account first, to_account_id %s %s", to_account_id, err.Error())
	}

	if from_account.TokenId != token_id {
		return "", fmt.Errorf("from_account %s is holding different tokens than %s", from_account_id, token_id)
	}

	if to_account.TokenId != token_id {
		return "", fmt.Errorf("from_account %s is holding different tokens than %s", from_account_id, token_id)
	}

	if to_account.TokenId != from_account.TokenId {
		return "", fmt.Errorf("From_Account %s and To_account %s hold different tokens", from_account_id, to_account_id)
	}

	if from_account.Balance < float64(quantity) {
		return "", fmt.Errorf("from-Account Id %s (org_id : %s, user_id : %s) for %s has insufficient balance for transfer", from_account_id, from_account.OrgId, from_account.UserId, token_id)
	}

	if err != nil {
		return "", fmt.Errorf("unable to generate timestamp %s", err.Error())
	}

	if time_to_expiration != "0" {
		expirationTime, err := time.Parse(time.RFC3339, time_to_expiration)

		if err != nil {
			return "", fmt.Errorf("unable to parse time_to_expiration: %s, time is in wrong format. Expected format is RFC3339 for example 2006-01-02T15:04:05+07:00 ", time_to_expiration)
		}

		stub := t.model.GetNetworkStub()
		tx, err := stub.GetTxTimestamp()
		if err != nil {
			return nil, err
		}
		timestamp := time.Unix(tx.Seconds, int64(tx.Nanos)).Format(time.RFC3339)

		currentTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, err
		}

		if currentTime.After(expirationTime) {
			return "", fmt.Errorf("time_to_expiration %s cannot be before current time %s ", expirationTime, currentTime.Format(time.RFC3339))
		}
	}

	holdAsset, err := t.holding.BuildHoldAsset(token_id, operation_id)
	if err != nil {
		return "", fmt.Errorf("error in Building Hold Asset %v", holdAsset)
	}

	holdAsset.FromAccountId = from_account_id
	holdAsset.ToAccountId = to_account_id
	holdAsset.NotaryAccountId = notary_account_id
	holdAsset.TokenId = token_id
	holdAsset.Quantity = quantity
	holdAsset.TimeToExpiration = time_to_expiration

	_, err = t.holding.Save(&holdAsset)
	if err != nil {
		return nil, err
	}

	tokenDecimalAllowed, err := t.GetDecimals(token_id)
	if err != nil {
		return nil, err
	}
	if util.GetDecimals(quantity) > tokenDecimalAllowed {
		return nil, fmt.Errorf("quantity has greater number of decimal places than maximum decimal places: %v for token_id: %s", tokenDecimalAllowed, token_id)
	}
	if quantity > from_account.Balance {
		return nil, fmt.Errorf("not sufficient balance for creating hold for Amount %v, current balance for from_account %s (org_id: %s, user_id: %s) is %v", quantity, from_account.AccountId, from_account.OrgId, from_account.UserId, from_account.Balance)
	}

	updated_Balance, err := t.decrement(from_account.Balance, quantity, tokenDecimalAllowed)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new balance after hold %s", err)
	}
	from_account.Balance = updated_Balance

	updated_Balance, err = t.increment(from_account.BalanceOnHold, quantity, tokenDecimalAllowed)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new onhold balance after hold %s", err)
	}
	from_account.BalanceOnHold = updated_Balance

	_, err = t.account.Update(&from_account)
	if err != nil {
		return nil, err
	}

	_, err = t.transaction.CreateTransaction(token_id, from_account_id, to_account_id, "ONHOLD", quantity, 0, holdAsset.HoldingId)
	if err != nil {
		return nil, fmt.Errorf("error in creating transaction asset %s", err.Error())
	}
	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("AccountId %s (org_id : %s, user_id : %s) is successfully holding %v tokens", from_account_id, from_account.OrgId, from_account.UserId, quantity)
	return response, nil
}

func (t *TokenReciever) ExecuteHold(operation_id string, quantity float64, tokenAsset interface{}) (interface{}, error) {

	caller_org_id := t.model.GetCreatorMspId()
	caller_user_id, err := t.model.GetUserId()
	if err != nil {
		return nil, fmt.Errorf("error in retrieving user_id of the caller %s", err.Error())
	}

	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}

	caller_id, err := t.account.GenerateAccountId(token_id, caller_org_id, caller_user_id)
	if err != nil {
		return nil, err
	}

	isCallerNotary, err := t.checkNotary(caller_id, tokenAsset)
	if err != nil {
		return nil, err
	}

	if !isCallerNotary {
		return nil, fmt.Errorf("not authorized to perform execute hold, caller %s (org_id: %s, user_id %s) is not a notary", caller_id, caller_org_id, caller_user_id)
	}

	result, err := t.checkBehaviors(token_id, "holdable")
	if err != nil {
		return "", err
	}
	if !result {
		return "", fmt.Errorf("holdable behavior is not configured for token_id %s. Please add this behavior to continue", token_id)
	}

	hold, err := t.holding.GetOnHoldDetailsWithOperationID(token_id, operation_id)
	if err != nil {
		return nil, err
	}

	// Check if Caller is the same notary
	if caller_id != hold.NotaryAccountId {
		return "", fmt.Errorf("cannot perform execute hold, caller %s is not the notary for the hold with operation id %s and token id %s", caller_id, operation_id, token_id)
	}

	if hold.Quantity == 0 {
		return "", fmt.Errorf("opertion Id %s does not hold any tokens", operation_id)
	}

	if quantity > hold.Quantity {
		return "", fmt.Errorf("hold quantity is %v, cannot release more than the hold quantity", hold.Quantity)
	}

	tokenDecimalAllowed, err := t.GetDecimals(token_id)
	if err != nil {
		return nil, err
	}
	if util.GetDecimals(quantity) > tokenDecimalAllowed {
		return nil, fmt.Errorf("token quantity has greater number of decimal places than maximum decimal places: %v for token_id: %s", tokenDecimalAllowed, token_id)
	}
	if err != nil {
		return "", err
	}

	from_account, err := t.account.GetAccount(hold.FromAccountId)
	if err != nil {
		return "", err
	}

	toUser, err := t.account.GetAccount(hold.ToAccountId)
	if err != nil {
		return "", err
	}
	balance := hold.Quantity - quantity

	hold.Quantity = balance

	_, err = t.holding.Update(&hold)
	if err != nil {
		return "", err
	}

	if quantity > from_account.BalanceOnHold {
		return "", fmt.Errorf("quantity is more than the Holding Quantity held with Account Id %s (org_id : %s, user_id : %s)", from_account.AccountId, from_account.OrgId, from_account.UserId)
	}
	//from_account.BalanceOnHold = from_account.BalanceOnHold - quantity
	updated_Balance, err := t.decrement(from_account.BalanceOnHold, quantity, tokenDecimalAllowed)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new onhold balance after executing hold %s", err)
	}
	from_account.BalanceOnHold = updated_Balance

	t.account.Update(&from_account)

	//toUser.Balance = toUser.Balance + quantity
	updated_Balance, err = t.increment(toUser.Balance, quantity, tokenDecimalAllowed)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new balance for to_account after executing hold %s", err)
	}
	toUser.Balance = updated_Balance
	t.account.Update(&toUser)

	_, err = t.transaction.CreateTransaction(token_id, hold.FromAccountId, hold.ToAccountId, "EXECUTEHOLD", quantity, 0, hold.HoldingId)
	if err != nil {
		return nil, fmt.Errorf("error in creating transaction asset %s", err.Error())
	}

	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Account Id %s (org_id : %s, user_id : %s) has successfully executed '%v' tokens(%s) from the hold with Operation Id '%s'", hold.FromAccountId, from_account.OrgId, from_account.UserId, quantity, token_id, operation_id)

	return response, nil
}

func (t *TokenReciever) ReleaseHold(operation_id string, tokenAsset interface{}) (interface{}, error) {

	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}

	check, err := t.checkBehaviors(token_id, "holdable")
	if err != nil {
		return "", err
	}
	if !check {
		return "", fmt.Errorf("holdable Behavior is not configured for token id %s. Please add behavior to continue", token_id)
	}
	stub := t.model.GetNetworkStub()
	tx, err := stub.GetTxTimestamp()
	if err != nil {
		return nil, err
	}
	timestamp := time.Unix(tx.Seconds, int64(tx.Nanos)).Format(time.RFC3339)

	hold, err := t.holding.GetOnHoldDetailsWithOperationID(token_id, operation_id)
	if err != nil {
		return "", err
	}

	// check caller is one of the 3 parties
	caller_org_id := t.model.GetCreatorMspId()
	caller_user_id, err := t.model.GetUserId()
	if err != nil {
		return nil, fmt.Errorf("error in retrieving user_id of the caller %s", err.Error())
	}
	caller_id, err := t.account.GenerateAccountId(token_id, caller_org_id, caller_user_id)
	if err != nil {
		return nil, err
	}
	if caller_id != hold.FromAccountId && caller_id != hold.NotaryAccountId && caller_id != hold.ToAccountId {
		return nil, fmt.Errorf("the tokens for Operation Id %s can be released only by Payer/Payee/Notary of this hold", operation_id)
	}
	currentTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, err
	}
	expiryTime := hold.TimeToExpiration
	if expiryTime == "0" {
		return nil, fmt.Errorf("time to expiration is '0', cannot perform release on this hold")
	}
	expirationTime, err := time.Parse(time.RFC3339, hold.TimeToExpiration)
	if err != nil {
		return nil, err
	}
	if currentTime.Before(expirationTime) && caller_id != hold.NotaryAccountId {
		return "", fmt.Errorf("currentokens for Operation Id %s can be released only by the notary before expiry time", operation_id)
	}
	if hold.Quantity == 0 {
		return "", fmt.Errorf("operation Id %s does not hold any tokens", operation_id)
	}

	from_account, err := t.account.GetAccount(hold.FromAccountId)

	if err != nil {
		return "", err
	}

	_, err = t.account.GetAccount(hold.NotaryAccountId)

	if err != nil {
		return "", err
	}

	balance := hold.Quantity
	hold.Quantity = 0

	t.holding.Update(&hold)

	tokenDecimalAllowed, err := t.GetDecimals(token_id)
	if err != nil {
		return nil, err
	}

	//from_account.Balance = from_account.Balance + balance
	updated_Balance, err := t.increment(from_account.Balance, balance, tokenDecimalAllowed)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new balance for from_account after releasing hold %s", err)
	}
	from_account.Balance = updated_Balance

	//from_account.BalanceOnHold = updated_Balance
	updated_Balance, err = t.decrement(from_account.BalanceOnHold, balance, tokenDecimalAllowed)
	if err != nil {
		return nil, fmt.Errorf("error in calculating new onHoldBalance for from_account after releasing hold %s", err)
	}
	from_account.BalanceOnHold = updated_Balance

	_, err = t.account.Update(&from_account)

	if err != nil {
		return "", err
	}

	_, err = t.transaction.CreateTransaction(token_id, hold.FromAccountId, hold.FromAccountId, "RELEASEHOLD", balance, 0, hold.HoldingId)
	if err != nil {
		return nil, fmt.Errorf("error in creating transaction asset %s", err.Error())
	}
	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Successfully released '%v' tokens from Operation Id '%s' to Account Id %s (org_id : %s, user_id : %s)", balance, operation_id, hold.FromAccountId, from_account.OrgId, from_account.UserId)

	return response, nil
}

func (t *TokenReciever) verifyRoleNameForToken(role string, token_id string) (bool, error) {
	token, _ := t.Get(token_id)
	tokenMap := token.(map[string]interface{})
	var minterRoleName string
	var burnerRoleName string
	var notaryRoleName string

	roleValue, ok := tokenMap["Roles"]
	if !ok {
		return false, fmt.Errorf("roles behavior are not configured for the token %s", token_id)
	}

	if val, ok := roleValue.(map[string]interface{})["minter_role_name"]; ok {
		minterRoleName = val.(string)
	}
	if val, ok := roleValue.(map[string]interface{})["burner_role_name"]; ok {
		burnerRoleName = val.(string)
	}
	if val, ok := roleValue.(map[string]interface{})["notary_role_name"]; ok {
		notaryRoleName = val.(string)
	}
	if role == minterRoleName || role == burnerRoleName || role == notaryRoleName {
		return true, nil
	}
	return false, nil
}

func (t *TokenReciever) AddRoleMember(role string, account_id string, tokenAsset interface{}) (interface{}, error) {
	if role == "" {
		return nil, fmt.Errorf("user role cannot be empty")
	}
	if account_id == "" {
		return nil, fmt.Errorf("account id cannot be empty")
	}
	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}

	// chekRolesInBehavior
	check, err := t.verifyRoleNameForToken(role, token_id)
	if err != nil {
		return nil, fmt.Errorf("error in validating role, %s", err.Error())
	}
	if !check {
		return nil, fmt.Errorf("role_name %s does not exist for token %s", role, token_id)
	}

	account, err := t.account.GetAccount(account_id)
	if err != nil {
		return nil, fmt.Errorf("account does not exist for account_id %s. Please create account first", account_id)
	}

	roleAsset, _ := t.tokenRole.GetRole(token_id, role, account_id)
	// role_name has already been assigned to account_id
	if !reflect.DeepEqual(roleAsset, tokenRole.Role{}) {
		response := make(map[string]interface{})
		response["msg"] = `Account Id: ` + account_id + ` (Org-Id: ` + account.OrgId + `, User-Id: ` + account.UserId + ` ) is already added to the role ` + role
		return response, nil
	}

	// roleName is not present for this account id
	roleAsset, err = tokenRole.BuildRoleAsset(token_id, role, account_id)
	if err != nil {
		return nil, err
	}
	_, err = t.tokenRole.Save(&roleAsset)
	if err != nil {
		return nil, err
	}
	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Successfully added role %s to %s (org_id : %s, user_id : %s)", role, account_id, account.OrgId, account.UserId)
	return response, nil
}

func (t *TokenReciever) IsInRole(role string, account_id string, tokenAsset interface{}) (bool, error) {

	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return false, err
	}
	_, err = t.account.GetAccount(account_id)
	if err != nil {
		return false, fmt.Errorf("error in getting account with id %s %s", account_id, err.Error())
	}

	check, err := t.verifyRoleNameForToken(role, token_id)
	if err != nil {
		return false, fmt.Errorf("error in validating role, %s", err.Error())
	}
	if !check {
		return false, fmt.Errorf("role_name %s does not exist for token %s", role, token_id)
	}

	_, err = t.tokenRole.GetRole(token_id, role, account_id)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (t *TokenReciever) RemoveRoleMember(role string, account_id string, tokenAsset interface{}) (interface{}, error) {
	if role == "" {
		return nil, fmt.Errorf("user role cannot be empty")
	}

	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return false, err
	}

	userAccount, err := t.account.GetAccount(account_id)
	if err != nil {
		return nil, fmt.Errorf("no account exist for account_id %s", account_id)
	}

	check, err := t.verifyRoleNameForToken(role, token_id)
	if err != nil {
		return nil, fmt.Errorf("error in validating role, %s", err.Error())
	}
	if !check {
		return nil, fmt.Errorf("role_name %s does not exist for token %s", role, token_id)
	}

	roleAsset, err := t.tokenRole.GetRole(token_id, role, account_id)

	if err != nil {
		return nil, fmt.Errorf("error in getting Role %s", err.Error())
	}

	_, err = t.tokenRole.Delete(roleAsset)
	if err != nil {
		return nil, err
	}
	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("successfully removed member_id %s (org_id : %s, user_id : %s) from role %s", account_id, userAccount.OrgId, userAccount.UserId, role)

	return response, nil
}

func (t *TokenReciever) GetDecimals(token_id string) (int, error) {
	token, err := t.Get(token_id)
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

func (t *TokenReciever) GetMaxMintQuantity(token_id string) (float64, error) {
	token, err := t.Get(token_id)
	if err != nil {
		return 0, err
	}

	tokenMap := token.(map[string]interface{})
	BehaviorValue := tokenMap["Behavior"].([]interface{})
	var behaviors []string
	for i := range BehaviorValue {
		behaviors = append(behaviors, BehaviorValue[i].(string))
	}

	if util.FindInStringSlice(behaviors, "mintable") {
		val, ok := tokenMap["Mintable"]
		if ok && val != nil {
			mintable_properties := val.(map[string]interface{})
			val, ok = mintable_properties["Max_mint_quantity"]
			if !ok {
				return 0, nil
			}
			return val.(float64), nil
		}
		return 0, nil
	} else {
		return 0, nil
	}
}

func (t *TokenReciever) GetTotalMintedTokens(tokenAsset interface{}) (map[string]interface{}, error) {
	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}
	token_name, err := util.GetTokenName(tokenAsset)
	if err != nil {
		return nil, err
	}
	metadata, err := t.getTokenMetadata(token_id, token_name)
	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("total minted amount for token with id %s is %v", token_id, metadata.Total_minted_amount)
	response["quantity"] = metadata.Total_minted_amount
	return response, err
}

func (t *TokenReciever) GetNetTokens(tokenAsset interface{}) (map[string]interface{}, error) {
	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return nil, err
	}
	token_name, err := util.GetTokenName(tokenAsset)
	if err != nil {
		return nil, err
	}
	metadata, err := t.getTokenMetadata(token_id, token_name)
	if err != nil {
		return nil, err
	}

	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("net minted amount for token with id %s is %v", token_id, metadata.Total_supply)
	response["quantity"] = metadata.Total_supply

	return response, nil
}

func (t *TokenReciever) BulkTransfer(flow []map[string]interface{}, tokenAsset interface{}) (interface{}, error) {
	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return 0, err
	}
	user_id, err := t.model.GetUserId()
	if err != nil {
		return nil, fmt.Errorf("error in retrieving user_id of the caller %s", err.Error())
	}
	from_account_id, err := t.account.GenerateAccountId(token_id, t.model.GetCreatorMspId(), user_id)
	if err != nil {
		return nil, err
	}
	return t.bulkTransferFrom(from_account_id, flow, tokenAsset)
}

func (t *TokenReciever) bulkTransferFrom(from_account_id string, flow []map[string]interface{}, tokenAsset interface{}) (interface{}, error) {

	token_id, err := util.GetTokenId(tokenAsset)
	if err != nil {
		return 0, err
	}
	var internalFlow []map[string]interface{}

	for _, element := range flow {
		elem := make(map[string]interface{})
		if element["quantity"].(float64) <= 0 {
			return nil, fmt.Errorf("cannot initiate Bulk Transfer, quantity has to be greater than 0 for all transfers")
		}
		elem["quantity"] = element["quantity"]
		account_id, err := t.account.GenerateAccountId(token_id, element["to_org_id"].(string), element["to_user_id"].(string))
		if err != nil {
			return nil, fmt.Errorf("cannot generate account id for org_id %s and user_id %s", element["to_org_id"].(string), element["to_user_id"].(string))
		}
		elem["to"] = account_id
		internalFlow = append(internalFlow, elem)
	}

	userBalanceMap := make(map[string]interface{})
	fromUserData := make(map[string]interface{})

	fromAccount, err := t.account.GetAccount(from_account_id)
	if err != nil {
		return nil, err
	}
	from_account_balance := fromAccount.Balance
	fromUserData["balance"] = from_account_balance
	userBalanceMap[from_account_id] = fromUserData

	for _, element := range internalFlow {
		if userBalanceMap[element["to"].(string)] == nil {
			to_User, err := t.account.GetAccount(element["to"].(string))
			if err != nil {
				return nil, err
			}
			if fromAccount.TokenId != to_User.TokenId {
				return nil, fmt.Errorf("transferring tokens of different token_id is not allowed, from_account %s and to_acount %s hold different tokens", from_account_id, to_User.AccountId)
			}
			toUserBalance := to_User.Balance
			toUserData := make(map[string]interface{})
			toUserData["balance"] = toUserBalance
			userBalanceMap[element["to"].(string)] = toUserData
		}
	}
	count := 1
	var currentFromAmount float64
	var tokenDecimalAllowed int
	var sub_transactions_response []map[string]interface{}
	for _, element := range internalFlow {
		fromAccount := userBalanceMap[from_account_id]
		fromAccountTransactionData := fromAccount.(map[string]interface{})
		currentFromAmount = fromAccount.(map[string]interface{})["balance"].(float64)

		quantity := element["quantity"].(float64)

		tokenDecimalAllowed, err = t.GetDecimals(token_id)
		if err != nil {
			return nil, err
		}
		if util.GetDecimals(quantity) > tokenDecimalAllowed {
			return nil, fmt.Errorf("quantity has greater number of decimal places than maximum decimal places: %v for token_id: %s", tokenDecimalAllowed, token_id)
		}
		if quantity > currentFromAmount {
			return nil, fmt.Errorf("not Sufficient balance for Transaction to %s for Amount %v", element["to"], element["quantity"])
		}

		//currentFromAmount -= quantity
		updated_Balance, err := t.decrement(currentFromAmount, quantity, tokenDecimalAllowed)
		if err != nil {
			return nil, fmt.Errorf("error in calculating new balance for from_account after transfer %s", err)
		}
		currentFromAmount = updated_Balance
		fromAccountTransactionData["balance"] = currentFromAmount

		userBalanceMap[from_account_id] = fromAccountTransactionData

		toAccountTransactionData := userBalanceMap[element["to"].(string)].(map[string]interface{})
		currentToAmount := toAccountTransactionData["balance"].(float64)
		updated_Balance, err = t.increment(currentToAmount, quantity, tokenDecimalAllowed)
		if err != nil {
			return nil, fmt.Errorf("error in calculating new balance for to_account after transfer %s", err.Error())
		}
		currentToAmount = updated_Balance
		toAccountTransactionData["balance"] = currentToAmount
		_, err = t.transaction.CreateTransaction(token_id, from_account_id, element["to"].(string), "TRANSFER", quantity, float64(count), "")
		if err != nil {
			return nil, fmt.Errorf("error in creating transaction asset %s", err.Error())
		}
		userBalanceMap[element["to"].(string)] = toAccountTransactionData

		//Adding subtransactions to the final result
		resultMap := make(map[string]interface{})
		resultMap["to_account_id"] = element["to"].(string)
		resultMap["amount"] = quantity
		sub_transactions_response = append(sub_transactions_response, resultMap)
		count++
	}
	total, err := t.decrement(from_account_balance, currentFromAmount, tokenDecimalAllowed)
	if err != nil {
		return nil, fmt.Errorf("error in calculating total transfered amount")
	}

	for key, userDataInMap := range userBalanceMap {
		userAsset, err := t.account.GetAccount(key)
		if err != nil {
			return nil, err
		}
		data := userDataInMap.(map[string]interface{})
		userAsset.Balance = data["balance"].(float64)
		_, err = t.account.Update(&userAsset)

		if err != nil {
			return nil, err
		}

	}

	_, err = t.transaction.CreateTransaction(token_id, from_account_id, "", "BULKTRANSFER", total, float64(count-1), "")
	if err != nil {
		return nil, fmt.Errorf("error in creating transaction asset %s", err.Error())
	}
	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Successfully transferred %v tokens from Account Id %s (Org-Id: %s, User-Id: %s)", total, from_account_id, fromAccount.OrgId, fromAccount.UserId)
	response["sub_transactions"] = sub_transactions_response
	response["from_account_id"] = from_account_id
	return response, nil
}


func (t *TokenReciever) GetTokensByName(token_name string) (interface{}, error) {
	queryString := fmt.Sprintf("SELECT key, valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType') = 'otoken' AND json_extract(valueJson, '$.Token_name') = '%s'", token_name)
	return t.model.Query(queryString)
}
