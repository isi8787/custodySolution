package auth

import (
	"encoding/json"
	"fmt"
	"strings"

	"example.com/TokenCC/lib/model"
	"example.com/TokenCC/lib/util"
	"example.com/TokenCC/lib/admin"
	"example.com/TokenCC/lib/account"
)

type AuthReciever struct {
	model *model.Model
	admin *admin.AdminReciever
	account *account.AccountReciever
}

func GetNewAuthReciever(m *model.Model, a *admin.AdminReciever, acc *account.AccountReciever) *AuthReciever {
	var authReciever AuthReciever
	authReciever.model = m
	authReciever.admin = a
	authReciever.account = acc
	return &authReciever
}

func (a *AuthReciever) CheckAuthorization(funcName string, args ...string) (bool, error) {
	if funcName == "" {
		return false, fmt.Errorf("function name is empty")
	}
	if len(args) > 0 {
		switch args[0] {
		case "TOKEN":
			return a.checkTokenAuthorization(funcName, args[1:])
		default:
			return false, fmt.Errorf("not a valid authorization check")
		}
	}
	return false, fmt.Errorf("improper arguments")
}

func (a *AuthReciever) isCallerAccountOwner(account_id string) bool {
	org_id := a.model.GetCreatorMspId()
	user_id, err := a.model.GetUserId()
	if err != nil {
		return false
	}
	asset, _ := a.model.Get(account_id)
	token_id := asset.(map[string]interface{})["TokenId"].(string)
	caller_account_id, _ := a.account.GenerateAccountId(token_id, org_id, user_id)
	return caller_account_id == account_id
}

func (a *AuthReciever) isCallerAccountOwner2(account_id string, userId string, tokenId string) bool {
	org_id := a.model.GetCreatorMspId()
	user_id, err := a.model.GetUserId()
	if err != nil {
		return false
	}
	caller_account_id, _ := a.account.GenerateAccountId(tokenId, org_id, user_id)
	return caller_account_id == account_id
}

func (a *AuthReciever) IsUserTokenAdmin(org_id string, user_id string) (interface{}, error) {
	if org_id == "" || user_id == "" {
		return false, fmt.Errorf("org_id or user_id cannot be empty")
	}

	adminListData, err := a.admin.GetAllAdmins()
	if err != nil {
		return false, fmt.Errorf("error in retrieving adminList %s", err.Error())
	}

	var foundAdmin bool
	foundAdmin = false
	for _, s := range adminListData {
		if org_id == s.OrgId && user_id == s.UserId {
			foundAdmin = true
			break
		}
	}
	response := make(map[string]interface{})
	response["result"] = foundAdmin
	return response, nil
}

func (a *AuthReciever) IsCallerAdmin() (bool, error) {
	adminListData, err := a.admin.GetAllAdmins()
	if err != nil {
		return false, fmt.Errorf("error in retrieving adminList %s", err.Error())
	}

	org_id := a.model.GetCreatorMspId()
	user_id, err := a.model.GetUserId()
	if err != nil {
		return false, fmt.Errorf("error in retrieving user id %s", err.Error())
	}

	var foundAdmin bool
	foundAdmin = false
	for _, s := range adminListData {
		adminData := s
		if org_id == adminData.OrgId && user_id == adminData.UserId {
			foundAdmin = true
			break
		}
	}
	return foundAdmin, nil
}

func (a * AuthReciever) isCallerOwnerOfMultipleAccounts (org_id string, user_id string) (bool, error) {
	if org_id == "" || user_id == "" {
		return false, fmt.Errorf("org_id or user_id cannot be empty")
	}
	caller_org_id := a.model.GetCreatorMspId()
	caller_user_id, err := a.model.GetUserId()
	if err != nil {
		return false, fmt.Errorf("error in retrieving user id %s", err.Error())
	}
	if caller_org_id == org_id && caller_user_id == user_id {
		return true, nil
	}
	return false, nil
}

func (a *AuthReciever) checkTokenAuthorization(funcName string, args []string) (bool, error) {
	split := strings.Split(funcName, ".")
	tokenMap := util.GetTokenAccessMap()
	mapBytes, err := json.Marshal(tokenMap)
	if err != nil {
		return false, fmt.Errorf("error in marshalling map %s", err.Error())
	}
	var accessMap map[string]interface{}
	unmarshalError := json.Unmarshal(mapBytes, &accessMap)
	if unmarshalError != nil {
		return false, fmt.Errorf("error in unmarshalling token map %s", unmarshalError.Error())
	}
	caller_org_id := a.model.GetCreatorMspId()
	caller_user_id, err := a.model.GetUserId()
	if err != nil {
		return false, fmt.Errorf("error in getting caller id")
	}

	if len(split) == 2 {
		className := split[0]
		funcName := split[1]

		if val, ok := accessMap[className]; ok {
			accessValues := val.(map[string]interface{})
			accessArray := accessValues[funcName]
			mandatoryChecks, ok := accessArray.([]interface{})
			if !ok {
				return true, nil
			}
			var errMsg []string
			for _, check := range mandatoryChecks {
				switch check.(string) {
				case "Admin":
					resp, err := a.IsCallerAdmin()
					if err != nil {
						return false, fmt.Errorf("error in validating as an admin %s", err.Error())
					}
					if resp {
						return true, nil
					} else {
						errMsg = append(errMsg, fmt.Sprintf("Caller (org_id: %s, user_id: %s) is not an Admin ", caller_org_id, caller_user_id))
					}
				case "AccountOwner":
					if len(args) == 0 {
						return false, fmt.Errorf("account id is not passed, require account id to authorize the caller as an Account Owner")
					}
					if args[0] == "" {
						return false, fmt.Errorf("account id is empty, cannot authorize the caller as an Account Owner")
					}

					if (funcName != "CreateAccount"){
						_, err := a.model.Get(args[0])
						if err != nil {
							return false, fmt.Errorf("account_id %s does not exist. Please create account first. Can not perform account owner check", args[0])
						}

						if a.isCallerAccountOwner(args[0]) {
							return true, nil
						} else {
							errMsg = append(errMsg, fmt.Sprintf("Caller (org_id: %s, user_id: %s) does not have the authorization to access the account", caller_org_id, caller_user_id))
						}
					} else {
						if a.isCallerAccountOwner2(args[0], args[1], args[2]) {
							return true, nil
						} else {
							errMsg = append(errMsg, fmt.Sprintf("Caller (org_id: %s, user_id: %s) does not have the authorization to access the account", caller_org_id, caller_user_id))
						}
					}
					
					
				case "MultipleAccountOwner":
					if len(args) < 2 {
						return false, fmt.Errorf("org_id and user_id is not passed for account owner check for multiple accounts")
					}
					resp, err := a.isCallerOwnerOfMultipleAccounts(args[0], args[1]) 
					if err != nil{
						return false , fmt.Errorf("error in performing multiple account owner check %s", err.Error())
					} 
					if resp {
						return true, nil
					} else {
						errMsg = append(errMsg, fmt.Sprintf("Caller does not have authorization to access the account for Org_id: %s & User_id: %s", args[0], args[1]))
					}
				default:
					return false, fmt.Errorf("not a valid authorization flag")
				}
			}
			if len(errMsg) > 0 {
				return false, fmt.Errorf("authorization failed with %s", errMsg)
			} else {
				return true, nil
			}
		}
		return false, fmt.Errorf("authorization for class does not exist %s", className)
	}
	return false, fmt.Errorf("invalid function Name for authorization")
}
