package src

import (
	"example.com/TokenCC/lib/admin"
	"example.com/TokenCC/lib/account"
	"example.com/TokenCC/lib/token"
	"fmt"
	"reflect"

	"example.com/TokenCC/lib/trxcontext"
)

type Router struct {
	Ctx trxcontext.TrxContext
}



/**
 * Needed at start, pass admin list for all future operations, can be called multiple times 
 */
func (t *Router) Init(adminList []admin.TokenAdminAsset) (interface{}, error) {
	list, err := t.Ctx.Admin.InitAdmin(adminList)
	if err != nil {
		return nil, fmt.Errorf("initialising admin list failed %s", err.Error())
	}
	return list, nil
}

//-----------------------------------------------------------------------------
//BasicToken
//-----------------------------------------------------------------------------

func (t *Router) InitializeToken(asset token.BasicToken) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.Save", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Token.Save(&asset)
}
func (t *Router) UpdateToken(asset token.BasicToken) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.Update", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Token.Update(&asset)
}

//-----------------------------------------------------------------------------
//Token Setup
//-----------------------------------------------------------------------------

func (t *Router) IsTokenAdmin(org_id string, user_id string) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Auth.IsTokenAdmin", "TOKEN", org_id, user_id)
	if err != nil || !auth {
		return false, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}

	return t.Ctx.Auth.IsUserTokenAdmin(org_id, user_id)
}

func (t *Router) AddTokenAdmin(org_id string, user_id string) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Admin.AddAdmin", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Admin.AddAdmin(org_id, user_id)
}

func (t *Router) GetAllTokenAdmins() (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Admin.GetAllAdmins", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Admin.GetAllAdminUsers()
}

func (t *Router) RemoveTokenAdmin(org_id string, user_id string) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Admin.RemoveAdmin", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Admin.RemoveAdmin(org_id, user_id)
}

func (t *Router) getTokenObject(token_id string) (reflect.Value, error) {
	if token_id == "" {
		return reflect.Value{}, fmt.Errorf("error in retrieving token, token_id cannot be empty")
	}
	tokenAsset, err := t.Ctx.Token.Get(token_id)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("no token exists with id %s %s", token_id, err.Error())
	}
	token_name := tokenAsset.(map[string]interface{})["AssetType"].(string)
	switch token_name {
	case "otoken":
		var asset token.BasicToken
		_, err := t.Ctx.Token.Get(token_id, &asset)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(&asset), nil
	default:
		return reflect.Value{}, fmt.Errorf("no token exists with token name %s", token_name)
	}
}

func (t *Router) GetTokenById(token_id string) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.Get", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	tokenAsset, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	return tokenAsset.Interface(), err
}

func (t *Router) GetTokenDecimals(token_id string) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.GetTokenDecimals", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller %s", err.Error())
	}
	tokenDecimal, err := t.Ctx.Token.GetDecimals(token_id)
	if err != nil {
		return nil, fmt.Errorf("error in GetTokenDecimals  %s", err.Error())
	}
	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Token Id: %s has %d decimal places.", token_id, tokenDecimal)
	return response, nil
}

func (t *Router) GetTokenList() (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.Get", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	result, err := t.Ctx.Token.GetAllTokens()
    if err != nil {
        return nil, fmt.Errorf("error get token list: %s", err.Error())
    }
    return result, nil
}

//-----------------------------------------------------------------------------
//Account Setup
//-----------------------------------------------------------------------------

func (t *Router) CreateAccount(token_id string, org_id string, user_id string, alias string, ecParams string) (interface{}, error) {
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}

	auth, err := t.Ctx.Auth.CheckAuthorization("Account.CreateAccount", "TOKEN", account_id, user_id, token_id)
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Account.CreateAccount(token_id, org_id, user_id, alias, ecParams)
}

func (t *Router) GetAccountHistory(token_id string, org_id string, user_id string) (interface{}, error) {
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}
	auth, err := t.Ctx.Auth.CheckAuthorization("Account.History", "TOKEN", account_id)
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Account.History(account_id)
}

func (t *Router) GetAccountTransactionHistory(token_id string, org_id string, user_id string) (interface{}, error) {
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}
	auth, err := t.Ctx.Auth.CheckAuthorization("Account.GetAccountTransactionHistory", "TOKEN", account_id)
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	transactionArray, err := t.Ctx.Account.GetAccountTransactionHistory(account_id)
	return transactionArray, err
}

func (t *Router) GetAccount(token_id string, org_id string, user_id string) (interface{}, error) {
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}
	auth, err := t.Ctx.Auth.CheckAuthorization("Account.GetAccount", "TOKEN", account_id)
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	accountAsset, err := t.Ctx.Account.GetAccount(account_id)
	return accountAsset, err
}

func (t *Router) GetAllAccounts() (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Account.GetAllAccounts", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Account.GetAllAccounts()
}

func (t *Router) GetUserByAccountId(account_id string) (interface{}, error) {
	return t.Ctx.Account.GetUserByAccountById(account_id)
}

func (t *Router) GetAccountBalance(token_id string, org_id string, user_id string) (interface{}, error) {
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}
	auth, err := t.Ctx.Auth.CheckAuthorization("Account.GetAccountBalance", "TOKEN", account_id)
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	response, err := t.Ctx.Account.GetAccountBalance(account_id)
	return response, err
}

func (t *Router) GetAccountsByUser(org_id string, user_id string) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Account.GetAccountsByUser", "TOKEN", org_id, user_id)
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Account.GetAccountsByUser(org_id, user_id)
}

func (t *Router) GetUserPubKey() (account.ECParameters, error) {

	//auth, err := t.Ctx.Auth.CheckAuthorization("Account.GetAccountsByUser", "TOKEN", org_id, user_id)  // To be defined
	//if err != nil && !auth {
	//	return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	//}
	return t.Ctx.Account.GetUserPubKey()
}

//-----------------------------------------------------------------------------
//Roles Setup
//-----------------------------------------------------------------------------

func (t *Router) AddRole(token_id string, user_role string, org_id string, user_id string) (interface{}, error) {
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}
	tokenAssetValue, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.AddRoleMember", "TOKEN", account_id)
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Token.AddRoleMember(user_role, account_id, tokenAssetValue.Interface())
}

func (t *Router) RemoveRole(token_id string, user_role string, org_id string, user_id string) (interface{}, error) {
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}
	tokenAssetValue, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.RemoveRoleMember", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Token.RemoveRoleMember(user_role, account_id, tokenAssetValue.Interface())
}

func (t *Router) GetAccountsByRole(token_id string, user_role string) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Role.GetAccountsByRole", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Role.GetAccountsByRole(token_id, user_role)
}

func (t *Router) GetUsersByRole(token_id string, user_role string) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Role.GetUsersByRole", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Role.GetUsersByRole(token_id, user_role)
}

func (t *Router) IsInRole(token_id string, org_id string, user_id string, user_role string) (interface{}, error) {
	tokenAssetValue, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.IsInRole", "TOKEN", account_id)
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	result, err := t.Ctx.Token.IsInRole(user_role, account_id, tokenAssetValue.Interface())
	if err != nil {
		return nil, fmt.Errorf("error in IsInRole  %s", err.Error())
	}
	response := make(map[string]interface{})
	response["result"] = result
	return response, nil
}

//-----------------------------------------------------------------------------
//Mintable Behavior
//-----------------------------------------------------------------------------

func (t *Router) IssueTokens(token_id string, quantity float64) (interface{}, error) {
	tokenAssetValue, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	return t.Ctx.Token.Mint(quantity, tokenAssetValue.Interface())
}

func (t *Router) GetTotalMintedTokens(token_id string) (interface{}, error) {
	tokenAssetValue, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.GetTotalMintedTokens", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Token.GetTotalMintedTokens(tokenAssetValue.Interface())
}

func (t *Router) GetNetTokens(token_id string) (interface{}, error) {
	tokenAssetValue, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	auth, err := t.Ctx.Auth.CheckAuthorization("Token.GetNetTokens", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Token.GetNetTokens(tokenAssetValue.Interface())
}

//-----------------------------------------------------------------------------
//Transferable Behavior
//-----------------------------------------------------------------------------

func (t *Router) TransferTokens(token_id string, to_org_id string, to_user_id string, quantity float64) (interface{}, error) {
	tokenAssetValue, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	to_account_id, err := t.Ctx.Account.GenerateAccountId(token_id, to_org_id, to_user_id)
	if err != nil {
		return nil, err
	}
	return t.Ctx.Token.Transfer(to_account_id, quantity, tokenAssetValue.Interface())
}

func (t *Router) BulkTransferTokens(token_id string, flow []map[string]interface{}) (interface{}, error) {
	tokenAssetValue, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	return t.Ctx.Token.BulkTransfer(flow, tokenAssetValue.Interface())
}

//-----------------------------------------------------------------------------
//Transactions
//-----------------------------------------------------------------------------

func (t *Router) GetTransactionsHistory(transaction_id string) (interface{}, error) {
	return t.Ctx.Transaction.GetTransactionsHistory(transaction_id)
}

func (t *Router) DeleteHistoricalTransactions(timestamp string) (interface{}, error) {
	auth, err := t.Ctx.Auth.CheckAuthorization("Transaction.DeleteHistoricalTransactions", "TOKEN")
	if err != nil && !auth {
		return nil, fmt.Errorf("error in authorizing the caller  %s", err.Error())
	}
	return t.Ctx.Transaction.DeleteHistoricalTransactions(timestamp)
}

//-----------------------------------------------------------------------------
//Burnable Behavior
//-----------------------------------------------------------------------------

func (t *Router) BurnTokens(orgId string, userId string, token_id string, quantity float64) (interface{}, error) {
	tokenAssetValue, err := t.getTokenObject(token_id)
	if err != nil {
		return nil, err
	}
	return t.Ctx.Token.Burn(orgId, userId, quantity, tokenAssetValue.Interface())
}


//-----------------------------------------------------------------------------
//Transaction Signature Behavior
//-----------------------------------------------------------------------------

func (t *Router) QueryTransaction(txHash string) (interface{}, error) {

	return t.Ctx.TxSignature.QueryTransaction(txHash)
}

func (t *Router) QueryExternalTransaction(txHash string) (interface{}, error) {

	return t.Ctx.TxSignature.QueryExternalTransaction(txHash)
}

func (t *Router) QueryThresholdTransaction(txHash string) (interface{}, error) {

	return t.Ctx.TxSignature.QueryThresholdTransaction(txHash)
}

func (t *Router) PostTransaction(txJSON string) (interface{}, error) {
	return t.Ctx.TxSignature.PostTransaction(txJSON)
}

func (t *Router) PostSignature(token_id string, org_id string, user_id, txHash string, signedMsg string, signedDate string, keyAlias string) (interface{}, error) {
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}

	return t.Ctx.TxSignature.PostSignature(txHash, signedMsg, signedDate, keyAlias, account_id)

}

func (t *Router) PostExternalTransaction(txJSON string) (interface{}, error) {
	return t.Ctx.TxSignature.PostExternalTransaction(txJSON)
}

func (t *Router) PostExternalTransactionSignature(token_id string, org_id string, user_id, identifier string, signedMsg string, signedDate string, keyAlias string) (interface{}, error) {
	account_id, err := t.Ctx.Account.GenerateAccountId(token_id, org_id, user_id)
	if err != nil {
		return nil, err
	}

	return t.Ctx.TxSignature.PostExternalTransactionSignature(identifier, signedMsg, signedDate, keyAlias, account_id)

}

func (t *Router) GetPrivateOrg(key string, orgId string) (interface{}, error){
	return t.Ctx.TxSignature.GetPrivateOrg(key, orgId)
}

func (t *Router) GenerateWalletSharedKey() (interface{}, error){
	return t.Ctx.TxSignature.GenerateWalletSharedKey()
}

func (t *Router) PostDealerKeyFragments() (interface{}, error){
	return t.Ctx.TxSignature.PostDealerKeyFragments()
}

func (t *Router) PostInitialThresholdTransaction(transaction string) (interface{}, error){
	return t.Ctx.TxSignature.PostInitialThresholdTransaction(transaction)
}

func (t *Router) GenerateSharedNonce(msg string, orgId string) (interface{}, error){
	return t.Ctx.TxSignature.GenerateSharedNonce(msg, orgId)
}


func (t *Router) PostTransactionSharedNonce(txHash string, orgId string, nonceShare string, noncePubShare string) (interface{}, error){
	return t.Ctx.TxSignature.PostTransactionSharedNonce(txHash, orgId, nonceShare, noncePubShare)
}

func (t *Router) TSignOrgSharedWallet(key string, orgId string) (interface{}, error){
	return t.Ctx.TxSignature.TSignOrgSharedWallet(key, orgId)
}

func (t *Router) PostSignShareWallet(key string, sigShareJSON string, orgId string) (interface{}, error){
	return t.Ctx.TxSignature.PostSignShareWallet(key, sigShareJSON, orgId)
}

func (t *Router) GetAllPendingTxs() (interface{}, error) {
	return t.Ctx.TxSignature.GetAllPendingTxs()
}

func (t *Router) GetAllThresholdTxs() (interface{}, error) {
	return t.Ctx.TxSignature.GetAllThresholdTxs()
}



func (t *Router) GenerateECDSAWalletSharedKey(userId string, tokenId string) (interface{}, error) {
	return t.Ctx.TxSignature.GenerateECDSAWalletSharedKey(userId, tokenId)
}

func (t *Router) PostECDSAWalletSharedKey() (interface{}, error) {
	return t.Ctx.TxSignature.PostECDSAWalletSharedKey()
}

func (t *Router) PostInitialECDSAThresholdTransaction(tx string) (interface{}, error) {
	return t.Ctx.TxSignature.PostInitialECDSAThresholdTransaction(tx)
}

func (t *Router) PrepareApproveECDSATx() (interface{}, error) {
	return t.Ctx.TxSignature.PrepareApproveECDSATx()
}

func (t *Router) ApproveECDSATx() (interface{}, error) {
	return t.Ctx.TxSignature.ApproveECDSATx()
}

func (t *Router) PerformECDSARounds(key string) (interface{}, error) {
	return t.Ctx.TxSignature.PerformECDSARounds(key)
}

func (t *Router) PostECDSASignature(key string, signature string) (interface{}, error) {
	return t.Ctx.TxSignature.PostECDSASignature(key, signature)
}

func (t *Router) QueryECDSAThresholdTransaction(txHash string) (interface{}, error) {

	return t.Ctx.TxSignature.QueryECDSAThresholdTransaction(txHash)
}

func (t *Router) GetAllPendingECDSATxs() (interface{}, error) {
	return t.Ctx.TxSignature.GetAllPendingECDSATxs()
}

func (t *Router) GetAllThresholdECDSATxs() (interface{}, error) {
	return t.Ctx.TxSignature.GetAllThresholdECDSATxs()
}

func (t *Router) GetPublicInfo(userId string, tokenId string) (interface{}, error) {
	return t.Ctx.TxSignature.GetPublicInfo(userId, tokenId)
}

func (t *Router) GetWalletId(userId string, tokenId string) (interface{}, error) {
	return t.Ctx.TxSignature.GetWalletId(userId, tokenId)
}

func (t *Router) PrepareEthTx(userId string, tokenId string, destination string, value string) (interface{}, error) {
	return t.Ctx.TxSignature.PrepareEthTx(userId, tokenId, destination, value)
}

func (t *Router) SubmitEthTx(txhash string) (interface{}, error) {
	return t.Ctx.TxSignature.SubmitEthTx(txhash)
}

func (t *Router) UpdateTxReceipt(txHash string, txreceipt string) (interface{}, error) {
	return t.Ctx.TxSignature.UpdateTxReceipt(txHash, txreceipt)
}

//func (t *Router) PostECDSARound1(txhash string, orgId string, round1_1b string) (interface{}, error) {
//	return t.Ctx.TxSignature.PostECDSARound1(txhash, orgId, round1_1b)
//}

//-----------------------------------------------------------------------------
//User Behavior
//-----------------------------------------------------------------------------

func (t *Router) RegisterUser(user_profile string) (interface{}, error) {
	return t.Ctx.User.RegisterUser(user_profile)
}

func (t *Router) GetUser(user_id string) (interface{}, error) {
	return t.Ctx.User.GetUser(user_id)
}

func (t *Router) GetUserByEmail(email string) (interface{}, error) {
	return t.Ctx.User.GetUserByEmail(email)
}

func (t *Router) GetUserByPhone(phone string) (interface{}, error) {
	return t.Ctx.User.GetUserByPhone(phone)
}

func (t *Router) UpdateUser(user_profile string) (interface{}, error) {
	return t.Ctx.User.UpdateUser(user_profile)
}

func (t *Router) AddConnector(user_id string, connectorId string, connectorJSON string) (interface{}, error) {
	return t.Ctx.User.AddConnector(user_id, connectorId, connectorJSON)
}

func (t *Router) RemoveConnectorAPI(user_id string, connectorId string, apiName string) (interface{}, error) {
	return t.Ctx.User.RemoveConnectorAPI(user_id, connectorId, apiName)
}
