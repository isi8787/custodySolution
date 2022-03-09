package admin

import (
	"encoding/json"
	"fmt"

	"example.com/TokenCC/lib/model"
	"example.com/TokenCC/lib/util"
)

const admin_asset_type = "oadmin"



type Admin struct {
	AssetType string `json:"AssetType" final:"oadmin"`
	Key       string `json:"Key" derived:"strategy=concat,format=%1~%2~%3,AssetType,OrgId,UserId" id:"true"`
	OrgId     string `json:"OrgId"`
	UserId    string `json:"UserId"`
}

type TokenAdminAsset struct {
	OrgId  string `json:"org_id"`
	UserId string `json:"user_id"`
}

type AdminReciever struct {
	model *model.Model
}

func GetNewAdminReciever(m *model.Model) *AdminReciever {
	var a AdminReciever
	a.model = m
	return &a
}

/**
 * Function initAdmin is called during Init of the chaincode if token asset is present.
 * It makes the user who instantiated the chaincode as Admin.
 */

 func (a *AdminReciever) InitAdmin(params []TokenAdminAsset) (interface{}, error) {
	
	var newAdminList []Admin
	for _, elements := range params {
		org_id := elements.OrgId
		user_id := elements.UserId
		newAdminList = append(newAdminList, Admin{
			OrgId:  org_id,
			UserId: user_id,
		})
	}

	if len(newAdminList) <= 0 {
		return nil, fmt.Errorf("List of admins not provided, it should be an array of objects with each object having 'user_id' & 'org_id'")
	}
	
	adminList, err := a.GetAllAdmins()
	if err != nil {
		for _, s := range newAdminList {
			_, err := a.model.Save(&s)
			if err != nil {
				return nil, fmt.Errorf("error in saving admin with org_id %s and user_id %s, %s", s.OrgId, s.UserId, err.Error())
			}
		}
	}
	for _, newAdmin := range newAdminList {
		found := false
		for _, admin := range adminList {
			if newAdmin.UserId == admin.UserId && newAdmin.OrgId == admin.OrgId {
				found = true
				break
			}
		}
		if !found {
			_, err := a.model.Save(&newAdmin)
			if err != nil {
				return nil, fmt.Errorf("error in saving admin with org_id %s and user_id %s, %s", newAdmin.OrgId, newAdmin.UserId, err.Error())
			}
		}
	}
	return newAdminList, nil
}

func (a *AdminReciever) AddAdmin(org_id string, user_id string) (interface{}, error) {
	err := util.ValidateOrgAndUser(org_id, user_id)
	if err != nil {
		return nil, err
	}
	adminList, err := a.GetAllAdmins()
	if err != nil {
		return nil, fmt.Errorf("error in getting admin list from ledger %s", err.Error())
	}

	var admin Admin
	admin.OrgId = org_id
	admin.UserId = user_id
	for _, s := range adminList {
		if s.OrgId == admin.OrgId && s.UserId == admin.UserId {
			return nil, fmt.Errorf("admin already exist with org_id %s and user_id %s", org_id, user_id)
		}
	}
	// adminList.UserList = append(adminList.UserList, admin)

	_, err = a.model.Save(&admin)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	result["msg"] = fmt.Sprintf("Successfully added Admin (Org_Id: %s, User_Id: %s)", org_id, user_id)
	return result, err
}

func (a *AdminReciever) RemoveAdmin(org_id string, user_id string) (interface{}, error) {
	err := util.ValidateOrgAndUser(org_id, user_id)
	if err != nil {
		return nil, err
	}
	caller_user_id, err := a.model.GetUserId()
	if err != nil {
		return nil, fmt.Errorf("error in retrieving user_id of the caller %s", err.Error())
	}

	if a.model.GetCreatorMspId() == org_id && caller_user_id == user_id {
		return nil, fmt.Errorf("user with org_id %s and user_id %s cannot revoke admin access of self", org_id, user_id)
	}


	var admin Admin
	admin.OrgId = org_id
	admin.UserId = user_id
	admin.AssetType = admin_asset_type

	adminAssetId, err := a.model.GetId(&admin)
	if err != nil {
		return "", fmt.Errorf("unable to generate ID for the admin asset. %s", err.Error())
	}

	_, err = a.model.Delete(adminAssetId)
	if err != nil {
		return nil, fmt.Errorf("error in deleting admin asset with id %s %s", adminAssetId, err.Error())
	}
	
	response := make(map[string]interface{})
	result := fmt.Sprintf("Successfuly removed Admin (Org_Id %s User_Id %s)", org_id, user_id)
	response["msg"] = result
	return response, err
}

func (a *AdminReciever) GetAllAdmins() ([]Admin, error) {
	adminStartId := util.TokenAdminRangeStartKey
	adminEndId := util.TokenAdminRangeStartKey + util.EndKeyChar

	adminListBuffer, err := a.model.GetByRangeFromLedger(adminStartId, adminEndId)
	if err != nil {
		return nil, fmt.Errorf("error in getting admin list from ledger %s", err.Error())
	}
	adminListMap, err := util.FilterRangeResultsByAssetType(admin_asset_type, adminListBuffer)
	if err != nil {
		return nil, fmt.Errorf("error in converting range results to admin list %s", err.Error())
	}
	mapBytes, err := json.Marshal(adminListMap)
	if err != nil {
		return nil, fmt.Errorf("error in marshalling map %s", err.Error())
	}
	var adminList []Admin
	err = json.Unmarshal(mapBytes, &adminList)
	if err != nil {
		return nil, fmt.Errorf("error in unmarshalling map %s", err.Error())
	}
	return adminList, nil
}

func (a *AdminReciever) GetAllAdminUsers() (interface{}, error) {
	adminList, err := a.GetAllAdmins()
	if err != nil {
		return nil, err
	}

	response := make([]map[string]interface{}, 0)
	for _, s := range adminList {
		adminElem := make(map[string]interface{})
		adminElem["OrgId"] = s.OrgId
		adminElem["UserId"] = s.UserId
		
		response = append(response, adminElem)
	}

	result := make(map[string]interface{})
	result["admins"] = response
	return result, nil
}