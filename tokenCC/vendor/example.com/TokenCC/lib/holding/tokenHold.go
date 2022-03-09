package holding

import (
	"encoding/json"
	"fmt"
	"example.com/TokenCC/lib/model"
	"example.com/TokenCC/lib/util"
)

type Hold struct {
	AssetType        string  `json:"AssetType" final:"ohold"`
	HoldingId        string  `json:"HoldingId" id:"true" derived:"strategy=concat,format=%1~%2~%3~%4,AssetType,TokenName,TokenId,OperationId"`
	OperationId      string  `json:"OperationId"`
	TokenName      string  `json:"TokenName"`
	FromAccountId    string  `json:"FromAccountId"`
	ToAccountId      string  `json:"ToAccountId"`
	NotaryAccountId  string  `json:"NotaryAccountId"`
	TokenId          string  `json:"TokenId"`
	Quantity         float64 `json:"Quantity"`
	TimeToExpiration string  `json:"TimeToExpiration"`
}

type HoldReciever struct {
	model *model.Model
}

func GetNewHoldReciever(m *model.Model) *HoldReciever {
	var h HoldReciever
	h.model = m
	return &h
}

// replica of the function in tokenAccount.go
func (h *HoldReciever) getTokenName(token_id string) (string, error) {
	if token_id == "" {
		return "", fmt.Errorf("unable to generate account_id since token_id is empty")
	}

	tokenAsset, err := h.model.Get(token_id)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve the token object for token-id: %s", token_id)
	}

	if tokenAsset.(map[string]interface{})["AssetType"].(string) == "otoken" {
		return tokenAsset.(map[string]interface{})["Token_name"].(string), nil
	}

	return "", fmt.Errorf("No token asset exists with token-id: %s", token_id)
}

const Hold_asset_type = "ohold"

func (h *HoldReciever) BuildHoldAsset(token_id string, operation_id string) (Hold, error) {
	var holdAsset Hold
	holdAsset.TokenId = token_id
	holdAsset.OperationId = operation_id

	tokenName, err := h.getTokenName(token_id)
	if err != nil {
		return Hold{}, err
	}
	holdAsset.TokenName = tokenName

	_, err = util.SetFinaTagInput(&holdAsset)
	if err != nil {
		return Hold{}, err
	}

	return holdAsset, nil
}

func (h *HoldReciever) validateHoldProperties(asset interface{}) error {

	assetAsBytes, errMarshal := json.Marshal(asset)
	if errMarshal != nil {
		return fmt.Errorf("marshal error %s", errMarshal.Error())
	}
	var holdObject map[string]interface{}
	unmarshalError := json.Unmarshal(assetAsBytes, &holdObject)
	if unmarshalError != nil {
		return fmt.Errorf("unmarshalling error %s", unmarshalError.Error())
	}
	val, ok := holdObject["OperationId"]
	if !ok {
		return fmt.Errorf("invalid hold operation id is missing")
	}
	if val.(string) == "" {
		return fmt.Errorf("operation_id cannot be an empty string")
	}
	if len(val.(string)) > 16 {
		return fmt.Errorf("length of operation_id cannot be greater than 16 characters. Operation Id is %s", val.(string))
	}
	return nil
}

func (h *HoldReciever) Save(asset interface{}) (interface{}, error) {
	err := h.validateHoldProperties(asset)
	if err != nil {
		return nil, fmt.Errorf("error in creating hold. Validation failed with error : %s", err.Error())
	}
	return h.model.Save(asset)
}

func (h *HoldReciever) Get(Id string) (Hold, error) {
	stub := h.model.GetNetworkStub()

	assetAsBytes, err := stub.GetState(Id)
	if err != nil {
		return Hold{}, fmt.Errorf("error in getting Hold with Id %s %s", Id, err.Error())
	}
	if assetAsBytes == nil {
		return Hold{}, fmt.Errorf("Hold with Id %s does not exist", Id)
	}

	var asset Hold
	unmarshalError := json.Unmarshal(assetAsBytes, &asset)
	if unmarshalError != nil {
		return Hold{}, fmt.Errorf("error in getting: marshalling error %s", unmarshalError.Error())
	}
	if asset.AssetType != Hold_asset_type {
		return Hold{}, fmt.Errorf("asset of type Hold with Id %s does not exist", Id)
	}
	return asset, nil
}

func (h *HoldReciever) getHoldingId(token_id string, operation_id string) (string, error) {
	holdAsset, err := h.BuildHoldAsset(token_id, operation_id)
	if err != nil {
		return "", err
	}
	id, err := h.model.GetId(&holdAsset)
	if err != nil {
		return "", fmt.Errorf("unable to generate ID for the user asset. %s", err.Error())
	}
	return id, nil
}

func (h *HoldReciever) Update(asset interface{}) (interface{}, error) {
	err := util.SetField(asset, "HoldingId", "")
	if err != nil {
		return nil, err
	}

	err = h.validateHoldProperties(asset)
	if err != nil {
		return nil, fmt.Errorf("error in updating the hold. Validation failed with error %s", err.Error())
	}
	return h.model.Update(asset)
}

func (h *HoldReciever) GetOnHoldBalanceWithOperationID(token_id string, operation_id string) (map[string]interface{}, error) {

	holdAsset, err := h.BuildHoldAsset(token_id, operation_id)

	if err != nil {
		return nil, fmt.Errorf("error in building hold asset %s", err.Error())
	}

	holdId, err := h.model.GetId(&holdAsset)

	if err != nil {
		return nil, err
	}
	holdData, err := h.Get(holdId)

	if err != nil {
		return nil, err
	}

	response := make(map[string]interface{})
	response["msg"] = fmt.Sprintf("Current Holding Balance of OperationId %s for token %s is : %v", operation_id, token_id, holdData.Quantity)
	response["holding_balance"] = holdData.Quantity

	return response, nil
}

func (h *HoldReciever) GetOnHoldDetailsWithOperationID(token_id string, operation_id string) (Hold, error) {
	holdingId, err := h.getHoldingId(token_id, operation_id)
	if err != nil {
		return Hold{}, err
	}
	holdAsset, err := h.Get(holdingId)
	if err != nil {
		return Hold{}, fmt.Errorf("error in getting Hold with Operation Id %s and Token Id %s %s", operation_id, token_id, err.Error())
	}
	return holdAsset, nil
}

func (h *HoldReciever) GetHistoryById(id string) (interface{}, error) {
	_, err := h.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error in getting hold %s", err.Error())
	}
	return h.model.GetHistoryById(id)
}
