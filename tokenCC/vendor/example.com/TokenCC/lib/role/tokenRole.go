package role

import (
	"encoding/json"
	"fmt"

	"example.com/TokenCC/lib/account"
	"example.com/TokenCC/lib/model"
	"example.com/TokenCC/lib/util"
)

const role_asset_type = "orole"

type Role struct {
	AssetType string `json:"AssetType" final:"orole"`
	Key       string `json:"Key" derived:"strategy=concat,format=%1~%2~%3~%4,AssetType,TokenId,RoleName,AccountId" id:"true"`
	RoleName  string `json:"RoleName"`
	TokenId   string `json:"TokenId"`
	AccountId string `json:"AccountID"`
}

type RoleReciever struct {
	model   *model.Model
	account *account.AccountReciever
}

func GetNewRoleReciever(m *model.Model, a *account.AccountReciever) *RoleReciever {
	var r RoleReciever
	r.model = m
	r.account = a
	return &r
}

func (r *RoleReciever) Save(asset interface{}) (interface{}, error) {

	return r.model.Save(asset)
}

func BuildRoleAsset(token_id string, role_name string, account_id string) (Role, error) {
	var roleAsset Role
	roleAsset.RoleName = role_name
	roleAsset.TokenId = token_id
	roleAsset.AccountId = account_id
	_, err := util.SetFinaTagInput(&roleAsset)
	if err != nil {
		return Role{}, fmt.Errorf("unable to build role asset %s", err.Error())
	}

	return roleAsset, nil
}
func (r *RoleReciever) GetRoleId(token_id string, role_name string, account_id string) (string, error) {
	asset, err := BuildRoleAsset(token_id, role_name, account_id)
	if err != nil {
		return "", err
	}
	id, err := r.model.GetId(&asset)
	if err != nil {
		return "", fmt.Errorf("unable to generate ID for the user asset. %s", err.Error())
	}
	return id, nil
}

func (r *RoleReciever) GetUsersByRole(token_id string, role_name string) (interface{}, error) {
	getAccountResponse, err := r.GetAccountsByRole(token_id, role_name)
	if err != nil {
		return nil, err
	}
	accountIds := getAccountResponse.(map[string]interface{})["accounts"].([]string)
	// accoundIdsInterface := getAccountResponse.(map[string]interface{})["accounts"].([]interface{})
	// accountIds:= make([]string, len(accoundIdsInterface))
	// for i, v := range accoundIdsInterface {
    // 	accountIds[i] = v.(string)
	// }
	var result []map[string]interface{}
	for _, element := range accountIds {
		a, err := r.account.GetAccount(element)
		if err == nil {
			element := make(map[string]interface{})
			element["user_id"] = a.UserId
			element["org_id"] = a.OrgId
			element["token_id"] = a.TokenId
			result = append(result, element)
		}
	}
	response := make(map[string]interface{})
	response["Users"] = result
	return response, nil
}

func (r *RoleReciever) GetAccountsByRole(token_id string, role_name string) (interface{}, error) {
	if token_id == "" {
		return nil, fmt.Errorf("error in generating role_id, token_id cannot be empty")
	}
	if role_name == "" {
		return nil, fmt.Errorf("error in generating role_id, role cannot be empty")
	}

	var roles []Role

	rolesStartString := fmt.Sprintf("%s~%s~%s~", role_asset_type, token_id, role_name)
	rolesEndString := rolesStartString + util.EndKeyChar

	result, err := r.model.GetByRangeFromLedger(rolesStartString, rolesEndString)
	if err != nil {
		return nil, fmt.Errorf("error in getting roles from ledger %s", err.Error())
	}
	rolesList, err := util.FilterRangeResultsByAssetType("orole", result)
	if err != nil {
		return nil, fmt.Errorf("error in converting range results to roles list %s", err.Error())
	}
	mapBytes, err := json.Marshal(rolesList)
	if err != nil {
		return nil, fmt.Errorf("error in marshalling map %s", err.Error())
	}
	err = json.Unmarshal(mapBytes, &roles)
	if err != nil {
		return nil, fmt.Errorf("error in unmarshalling map %s", err.Error())
	}

	if err != nil {
		return nil, err
	}

	var accounts []string
	response := make(map[string]interface{})

	for _, elem := range roles {
		accounts = append(accounts, elem.AccountId)
	}
	response["accounts"] = accounts
	return response, nil
}

func (r *RoleReciever) get(Id string) (Role, error) {
	stub := r.model.GetNetworkStub()
	assetAsBytes, err := stub.GetState(Id)
	if err != nil {
		return Role{}, fmt.Errorf("error in getting role with Id %s %s", Id, err.Error())
	}
	if assetAsBytes == nil {
		return Role{}, fmt.Errorf("Role with Id %s does not exist", Id)
	}

	var asset Role
	unmarshalError := json.Unmarshal(assetAsBytes, &asset)
	if unmarshalError != nil {
		return Role{}, fmt.Errorf("error in getting: marshalling error %s", unmarshalError.Error())
	}
	if asset.AssetType != role_asset_type {
		return Role{}, fmt.Errorf("asset of type Role with Id %s does not exist", Id)
	}
	return asset, nil
}

func (r *RoleReciever) GetRole(token_id string, role string, account_id string) (Role, error) {

	id, err := r.GetRoleId(token_id, role, account_id)
	if err != nil {
		return Role{}, err
	}
	return r.get(id)
}

func (r *RoleReciever) Update(asset interface{}) (interface{}, error) {
	err := util.SetField(asset, "Key", "")
	if err != nil {
		return nil, err
	}
	return r.model.Update(asset)
}

func (r *RoleReciever) GetHistoryById(Id string) (interface{}, error) {
	_, err := r.get(Id)
	if err != nil {
		return nil, fmt.Errorf("error in getting role with id %s %s", Id, err.Error())
	}
	return r.model.GetHistoryById(Id)
}

 func (r *RoleReciever) Delete(asset interface{}) (interface{}, error) {
	roleAsset := asset.(Role)
	id, err := r.model.GetId(&roleAsset)
	if err != nil {
		return "", fmt.Errorf("unable to generate ID for the user asset. %s", err.Error())
	}
	return r.model.Delete(id)
}
 