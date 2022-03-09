package user

import (
	"encoding/json"
	"fmt"

	"example.com/TokenCC/lib/model"
)

type User struct {
	UserId           string  `json:"id" id:"true" mandatory:"true"`
	Email            string  `json:"email"`
	Verified_Email   bool    `json:"verified_email"`
	GivenName        string  `json:"given_name"`
	FamilyName       string  `json:"family_name"`
	Phone            string  `json:"phone"`
	OrgId            string  `json:"orgId"`
	Connectors       map[string][]Connector  `json:"connectors"`
}

type Connector struct{
    APIName    string  `json: "apiName"`
    APIKey     string  `json:"apiKey"`
    APISecret  string  `json:"apiSecret"`
    Passphrase string  `json: "passphrase,omitempty"`
    ClientId   string  `json: "clientId,omitempty"`
}

type UserReciever struct {
	model       *model.Model
}

func GetNewUserReciever(m *model.Model) *UserReciever {
	var u UserReciever
	u.model = m
	return &u
}

func (u *UserReciever) RegisterUser(userJson string) (User, error) {
	if userJson == "" {
		return User{}, fmt.Errorf("JSON with User can not be empty")
	}

	var user User
    unmarshalError := json.Unmarshal([]byte(userJson), &user)
    if unmarshalError != nil {
        return User{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
    }

    _, err := u.get(user.UserId)
	if err == nil {
		return User{}, fmt.Errorf("User Already Exist with Id %s", user.UserId)
	}

	if(user.Phone != ""){
        up, _ := u.GetUserByPhone(user.Phone)
        userListBytes, _ := json.Marshal(up)
        var userList []User
        json.Unmarshal(userListBytes, &userList)
        if (len(userList)>0) {
            return User{}, fmt.Errorf("User Already Exist with Phone %s", user.Phone)
        }
	}

	if(user.Email != ""){
        ue, _ := u.GetUserByEmail(user.Email)
        userListBytes, _ := json.Marshal(ue)
        var userList []User
        json.Unmarshal(userListBytes, &userList)
        if (len(userList)>0) {
            return User{}, fmt.Errorf("User Already Exist with Phone %s", user.Email)
        }
	}

	_, err = u.model.Save(&user)
	if err != nil {
		return User{}, fmt.Errorf("error in saving user with User_Id %s , Error: %s", user.UserId, err.Error())
	}
	return user, nil
}

func (u *UserReciever) UpdateUser(userJson string) (User, error) {
	if userJson == "" {
		return User{}, fmt.Errorf("JSON with User can not be empty")
	}

	var user User
    unmarshalError := json.Unmarshal([]byte(userJson), &user)
    if unmarshalError != nil {
        return User{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
    }

    existingUser, err := u.get(user.UserId)
	if err != nil {
		return User{}, fmt.Errorf("Error Finding User with Id %s", user.UserId)
	}

    if(user.Phone != existingUser.Phone){
        up, _ := u.GetUserByPhone(user.Phone)
        userListBytes, _ := json.Marshal(up)
        var userList []User
        json.Unmarshal(userListBytes, &userList)
        if (len(userList)>0) {
            return User{}, fmt.Errorf("User Already Exist with Phone %s", user.Phone)
        }
    }

    if(user.Email != existingUser.Email){
        ue, _ := u.GetUserByEmail(user.Email)
        userListBytes, _ := json.Marshal(ue)
        var userList []User
        json.Unmarshal(userListBytes, &userList)
        if (len(userList)>0) {
            return User{}, fmt.Errorf("User Already Exist with Phone %s", user.Phone)
        }
    }

	_, err = u.model.Update(&user)
	if err != nil {
		return User{}, fmt.Errorf("Error in updating user with User_Id %s , Error: %s", user.UserId, err.Error())
	}
	return user, nil
}

func (u *UserReciever) GetUser(user_id string) (User, error) {
	if user_id == "" {
		return User{}, fmt.Errorf("error in retrieving account, account id is empty")
	}
	user, err := u.get(user_id)
	if err != nil {
		return User{}, fmt.Errorf("error in getting user %s", err.Error())
	}
	return user, nil
}

func (u *UserReciever) GetUserAccount(user_id string) (User, error) {
	if user_id == "" {
		return User{}, fmt.Errorf("error in retrieving account, account id is empty")
	}
	user, err := u.get(user_id)
	if err != nil {
		return User{}, fmt.Errorf("error in getting user %s", err.Error())
	}
	return user, nil
}

func (u *UserReciever) get(Id string) (User, error) {
	stub := u.model.GetNetworkStub()
	userAsBytes, err := stub.GetState(Id)
	if err != nil {
		return User{}, fmt.Errorf("error in getting User with Id %s %s", Id, err.Error())
	}
	if userAsBytes == nil {
		return User{}, fmt.Errorf("User with Id %s does not exist", Id)
	}

	var user User
	unmarshalError := json.Unmarshal(userAsBytes, &user)
	if unmarshalError != nil {
		return User{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
	}

	return user, nil
}

func (u *UserReciever) GetUserByEmail(email string) (interface{}, error) {
    query := fmt.Sprintf(`{"selector": {"email":"%s"}}`, email)
	return u.model.Query(query)
}

func (u *UserReciever) GetUserByPhone(phone string) (interface{}, error) {
    query := fmt.Sprintf(`{"selector": {"phone":"%s"}}`, phone)
	return u.model.Query(query)
}

func (u *UserReciever) AddConnector(userId string, connectorId string, connectorJSON string) (User, error) {
	if userId == "" {
		return User{}, fmt.Errorf("JSON with User can not be empty")
	}


	var connector Connector
    unmarshalError := json.Unmarshal([]byte(connectorJSON), &connector)
    if unmarshalError != nil {
        return User{}, fmt.Errorf("marshalling error %s", unmarshalError.Error())
    }

    existingUser, err := u.get(userId)
	if err != nil {
		return User{}, fmt.Errorf("Error Finding User with Id %s", userId)
	}

    if len(existingUser.Connectors) == 0 {
        var connectorList = make(map[string][]Connector)
        connectorList[connectorId] = []Connector{connector}
        existingUser.Connectors = connectorList
    } else {
        if exchangeAPIs, ok := existingUser.Connectors[connectorId]; ok {
            exist := false
            for _, api := range exchangeAPIs {
                if (api.APIName == connector.APIName){
                exist = true
                break
                }
            }
            if (exist) {
                return User{}, fmt.Errorf("Connector already has api with name %s", connector.APIName)
            } else {
                exchangeAPIs = append(exchangeAPIs, connector)
                existingUser.Connectors[connectorId] = exchangeAPIs
            }
        } else {
                existingUser.Connectors[connectorId] = []Connector{connector}
        }
    }

	_, err = u.model.Update(&existingUser)
	if err != nil {
		return User{}, fmt.Errorf("Error in updating user with User_Id %s , Error: %s", userId, err.Error())
	}
	return existingUser, nil
}

func (u *UserReciever) RemoveConnectorAPI(userId string, connectorId string, apiName string) (User, error) {
	if userId == "" {
		return User{}, fmt.Errorf("JSON with User can not be empty")
	}

    existingUser, err := u.get(userId)
	if err != nil {
		return User{}, fmt.Errorf("Error Finding User with Id %s", userId)
	}

	connectorAPIs := existingUser.Connectors[connectorId]
	var updatedList []Connector
	for i, val := range connectorAPIs {
	    if(val.APIName == apiName){
	        connectorAPIs[i] = connectorAPIs[len(connectorAPIs)-1]
	        updatedList = connectorAPIs[:len(connectorAPIs)-1]
	        break
	    }
	}
	existingUser.Connectors[connectorId] = updatedList
	_, err = u.model.Update(&existingUser)
	if err != nil {
		return User{}, fmt.Errorf("Error in updating user with User_Id %s , Error: %s", userId, err.Error())
	}
	return existingUser, nil
}