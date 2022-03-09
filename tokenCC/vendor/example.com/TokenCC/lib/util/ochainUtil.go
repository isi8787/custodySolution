package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"crypto/md5"

	"example.com/TokenCC/lib/util/validators"
	"github.com/creasty/defaults"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type TokenAccess struct {
	Save                 []string
	GetAllTokens         []string
	Get                  []string
	Update               []string
	GetTokenDecimals     []string
	GetTokensByName      []string
	AddRoleMember        []string
	RemoveRoleMember     []string
	IsInRole             []string
	Mint                 []string
	GetTotalMintedTokens []string
	GetNetTokens         []string
	Transfer             []string
	BulkTransfer         []string
	Hold                 []string
	ExecuteHold          []string
	ReleaseHold          []string
	Burn                 []string
}

type AdminAccess struct {
	AddAdmin     []string
	RemoveAdmin  []string
	GetAllAdmins []string
}

type RoleAccess struct {
	GetAccountsByRole []string
	GetUsersByRole    []string
}

type TransactionAccess struct {
	DeleteHistoricalTransactions []string
}

type AccountAccess struct {
	CreateAccount                []string
	GetAllAccounts               []string
	GetUserByAccountID           []string
	GetAccount                   []string
	History                      []string
	GetAccountTransactionHistory []string
	GetAccountBalance            []string
	GetAccountOnHoldBalance      []string
	GetOnHoldIds                 []string
	GetAccountsByUser            []string	
}

type HoldAccess struct {
	GetOnHoldDetailsWithOperationId []string
	GetOnHoldBalanceWithOperationId []string
}

type AuthAccess struct {
	IsTokenAdmin []string
}

type TokenAccessControl struct {
	Token   TokenAccess
	Role    RoleAccess
	Account AccountAccess
	Hold    HoldAccess
	Admin   AdminAccess
	Transaction TransactionAccess
	Auth AuthAccess
}

const RegexForTokenId = "^[A-Za-z0-9][A-Za-z0-9_-]*$"
const RegexForUserAndOrgId = "^[A-Za-z0-9][A-Za-z0-9_@.-]*$"
const TokenIdPrefix = "oaccount"
const TokenAdminRangeStartKey = "oadmin~"
const TokenHoldRangeStartKey = "ohold~"
const TokenRoleRangeStartKey = "orole~"
const EndKeyChar = "�"

var ChaincodeName string

type ErrorMap map[string]ErrorArray

func (err ErrorMap) Error() string {
	var b bytes.Buffer
	for k, errs := range err {
		if len(errs) > 0 {
			b.WriteString(fmt.Sprintf("%s: %s, ", k, errs.Error()))
		}
	}
	return strings.TrimSuffix(b.String(), ", ")
}

type ErrorArray []error

func GetTokenAccessMap() TokenAccessControl {
	var t TokenAccess
	var r RoleAccess
	var a AccountAccess
	var h HoldAccess
	var ad AdminAccess
	var trx TransactionAccess
	var auth AuthAccess
	auth.IsTokenAdmin = []string{"Admin", "MultipleAccountOwner"}

	trx.DeleteHistoricalTransactions = []string{"Admin"}
	ad.AddAdmin = []string{"Admin"}
	ad.RemoveAdmin = []string{"Admin"}
	ad.GetAllAdmins = []string{"Admin"}
	t.Save = []string{"Admin"}
	t.GetAllTokens = []string{"Admin"}
	t.Update = []string{"Admin"}
	t.GetTokenDecimals = []string{"Admin"}
	t.GetTokensByName = []string{"Admin"}
	t.AddRoleMember = []string{"Admin"}
	t.RemoveRoleMember = []string{"Admin"}
	t.IsInRole = []string{"Admin", "AccountOwner"}
	t.GetTotalMintedTokens = []string{"Admin"}
	t.GetNetTokens = []string{"Admin"}
	t.Get = []string{"Admin"}

	a.CreateAccount = []string{"Admin", "AccountOwner"}
	a.GetAllAccounts = []string{"Admin"}
	a.GetAccount = []string{"Admin", "AccountOwner"}
	a.History = []string{"Admin", "AccountOwner"}
	a.GetAccountTransactionHistory = []string{"Admin", "AccountOwner"}
	a.GetAccountBalance = []string{"Admin", "AccountOwner"}
	a.GetAccountOnHoldBalance = []string{"Admin", "AccountOwner"}
	a.GetOnHoldIds = []string{"Admin", "AccountOwner"}
	a.GetAccountsByUser = []string{"Admin", "MultipleAccountOwner"}

	r.GetAccountsByRole = []string{"Admin"}
	r.GetUsersByRole = []string{"Admin"}
	var accessMap TokenAccessControl
	accessMap.Token = t
	accessMap.Account = a
	accessMap.Hold = h
	accessMap.Role = r
	accessMap.Admin = ad
	accessMap.Auth = auth

	return accessMap

}

func (err ErrorArray) Error() string {
	var b bytes.Buffer

	for _, errs := range err {
		b.WriteString(fmt.Sprintf("%s, ", errs.Error()))
	}

	errs := b.String()
	return strings.TrimSuffix(errs, ", ")
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

		if assetTypeString == assetType {
			resultAssets = append(resultAssets, mapAsset)
		}
	}
	return resultAssets, nil
}

// CreateModel constructs the struct object from the given jsonString
func CreateModel(obj interface{}, inputString string) error {
	err := json.Unmarshal([]byte(inputString), &obj)
	if err != nil {
		return fmt.Errorf("error in creating asset %s", err.Error())
	}
	if err = defaults.Set(obj); err != nil {
		return fmt.Errorf("failure in default setting %s ", err.Error())
	}
	err = validators.ValidateStruct(obj)
	if err != nil {
		return err
	}
	return nil
}

func FindInStringSlice(source []string, value string) bool {
	for _, item := range source {
		if item == value {
			return true
		}
	}
	return false
}

func StringSliceSplice(a []string, i int) []string {
	// Copy last element to index i.
	a[i] = a[len(a)-1]
	// Erase last element (write zero value).
	a[len(a)-1] = ""
	a = a[:len(a)-1]
	return a
}

func FindIndexInStringSlice(source []string, value string) int {
	for index, item := range source {
		if item == value {
			return index
		}
	}
	return -1
}

// SetField sets the structField of a struct object with the given value
func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)
	if !structFieldValue.IsValid() {
		return fmt.Errorf("Error in setting field: No such field: %s in obj", name)
	}
	if !structFieldValue.CanSet() {
		return fmt.Errorf("Error in setting field: Cannot set %s field value", name)
	}
	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New(fmt.Sprintln("Error in setting field: Provided value type didn't match obj field type. Field-", structFieldValue, "Value-", val))
	}
	structFieldValue.Set(val)
	return nil
}

func DeepValidate(f reflect.Value, m ErrorMap, fnameFn func() string, userInput interface{}) {
	switch f.Kind() {
	case reflect.Interface, reflect.Ptr:
		if f.IsNil() {
			return
		}
		DeepValidate(f.Elem(), m, fnameFn, userInput)
	case reflect.Struct:
		//Skipping recursive validation for date
		if f.Type().String() == "date.Date" {

			return
		}
		subm := make(ErrorMap)
		err := structValidationHandler(f, subm, userInput.(map[string]interface{}))
		parentName := fnameFn()
		if err != nil {
			m[parentName] = ErrorArray{err}
		}
		for j, k := range subm {
			keyName := j
			if parentName != "" {
				keyName = parentName + "." + keyName
			}
			m[keyName] = k
		}
	case reflect.Array, reflect.Slice:
		// we don't need to loop over every byte in a byte slice so we only end up
		// looping when the kind is something we care about
		switch f.Type().Elem().Kind() {
		case reflect.Struct, reflect.Interface, reflect.Ptr, reflect.Map, reflect.Array, reflect.Slice:
			for i := 0; i < f.Len(); i++ {
				DeepValidate(f.Index(i), m, func() string {
					return fmt.Sprintf("%s[%d]", fnameFn(), i)
				}, userInput.([]interface{})[i])
			}
		}
	case reflect.Map:
		for _, key := range f.MapKeys() {
			DeepValidate(key, m, func() string {
				return fmt.Sprintf("%s[%+v](key)", fnameFn(), key.Interface())
			}, key.Interface()) // validate the map key
			value := f.MapIndex(key)
			DeepValidate(value, m, func() string {
				return fmt.Sprintf("%s[%+v](value)", fnameFn(), key.Interface())
			}, userInput.(map[string]interface{})[key.Interface().(string)])
		}
	}
}

func structValidationHandler(sv reflect.Value, m ErrorMap, userInput map[string]interface{}) error {
	kind := sv.Kind()
	if (kind == reflect.Ptr || kind == reflect.Interface) && !sv.IsNil() {
		return structValidationHandler(sv.Elem(), m, userInput)
	}
	if kind != reflect.Struct && kind != reflect.Interface {
		return fmt.Errorf("type is unsupported")
	}
	st := sv.Type()
	nfields := st.NumField()
	for i := 0; i < nfields; i++ {
		userInputValue := userInput[st.Field(i).Name]
		mandatoryTagValue := st.Field(i).Tag.Get("mandatory")
		derivedtagValue := st.Field(i).Tag.Get("derived")
		defaulttagValue := st.Field(i).Tag.Get("default")
		if userInputValue != nil {
			if err := validateStructField(st.Field(i), sv.Field(i), m, userInputValue); err != nil {
				return err
			}
		} else if mandatoryTagValue == "true" && derivedtagValue == "" {
			return fmt.Errorf("field %s is a mandatory field. It is missing from the input", st.Field(i).Name)
		}
		if userInputValue == nil && defaulttagValue != "" {
			if sv.Field(i).Kind() != reflect.Struct {
				fieldValue, err := convert(sv.Field(i).Kind(), defaulttagValue, sv.Field(i).Type())
				if err != nil {
					return fmt.Errorf("could not set default value for field %s, Error: %s", st.Field(i).Name, err.Error())
				}
				sv.Field(i).Set(fieldValue)
			}
		}
	}
	return nil
}

func validateStructField(fieldDef reflect.StructField, fieldVal reflect.Value, m ErrorMap, userInputValue interface{}) error {
	tag := fieldDef.Tag.Get("validate")
	if tag == "-" {
		return nil
	}
	// deal with pointers
	for (fieldVal.Kind() == reflect.Ptr || fieldVal.Kind() == reflect.Interface) && !fieldVal.IsNil() {
		fieldVal = fieldVal.Elem()
	}

	// ignore private structs unless Anonymous
	if !fieldDef.Anonymous && fieldDef.PkgPath != "" {
		return nil
	}

	var errs ErrorArray
	if tag != "" {
		var err error
		if fieldDef.PkgPath != "" {
			err = errors.New("cannot validate unexported struct")
		} else {
			err = validators.Validate(fieldVal.Interface(), tag)
		}
		if errarr, ok := err.(ErrorArray); ok {
			errs = errarr
		} else if err != nil {
			errs = ErrorArray{err}
		}
	}

	// no-op if field is not a struct, interface, array, slice or map
	DeepValidate(fieldVal, m, func() string {
		return fieldDef.Name
	}, userInputValue)

	if len(errs) > 0 {
		n := fieldDef.Name
		m[n] = errs
	}
	return nil
}

// ConvertMapToStructBasic is an utility function to construct a struct from a given map[string]interface{}.
// This does not work for complex types.
func ConvertMapToStructBasic(inputMap map[string](interface{}), resultStruct interface{}) error {
	for key, value := range inputMap {
		err := SetField(resultStruct, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// ConvertMapToStruct is an another utility function to construct a struct from a given map[string]interface{}.
// This can handle complex types and custom types.
func ConvertMapToStruct(inputMap map[string](interface{}), resultStruct interface{}) error {
	mapBytes, err := json.Marshal(inputMap)
	if err != nil {
		return fmt.Errorf("Error in marshalling map %s", err.Error())
	}
	err = json.Unmarshal(mapBytes, resultStruct)
	if err != nil {
		return fmt.Errorf("Error in unmarshalling map %s", err.Error())
	}
	return nil
}

func makeFirstLetterLowerCaps(input string) string {
	runes := []rune(input)
	if len(runes) > 0 {
		runes[0] = unicode.ToLower(runes[0])
	}
	return string(runes)
}

func convert(argKind reflect.Kind, arg string, argType reflect.Type) (reflect.Value, error) {
	switch argKind {
	case reflect.Bool:
		val, err := strconv.ParseBool(arg)
		if err == nil {
			return reflect.ValueOf(val).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Int:
		val, err := strconv.ParseInt(arg, 10, 64)
		if err == nil {
			return reflect.ValueOf(int(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Int8:
		val, err := strconv.ParseInt(arg, 10, 8)
		if err == nil {
			return reflect.ValueOf(int8(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Int16:
		val, err := strconv.ParseInt(arg, 10, 16)
		if err == nil {
			return reflect.ValueOf(int16(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Int32:
		val, err := strconv.ParseInt(arg, 10, 32)
		if err == nil {
			return reflect.ValueOf(int32(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Int64:
		val, err := time.ParseDuration(arg)
		if err == nil {
			return reflect.ValueOf(val).Convert(argType), err
		} else if val, err := strconv.ParseInt(arg, 10, 64); err == nil {
			return reflect.ValueOf(val).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Uint:
		val, err := strconv.ParseUint(arg, 10, 64)
		if err == nil {
			return reflect.ValueOf(uint(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Uint8:
		val, err := strconv.ParseUint(arg, 10, 8)
		if err == nil {
			return reflect.ValueOf(uint8(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Uint16:
		val, err := strconv.ParseUint(arg, 10, 16)
		if err == nil {
			return reflect.ValueOf(uint16(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Uint32:
		val, err := strconv.ParseUint(arg, 10, 32)
		if err == nil {
			return reflect.ValueOf(uint32(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Uint64:
		val, err := strconv.ParseUint(arg, 10, 64)
		if err == nil {
			return reflect.ValueOf(val).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Uintptr:
		val, err := strconv.ParseUint(arg, 10, 64)
		if err == nil {
			return reflect.ValueOf(uintptr(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Float32:
		val, err := strconv.ParseFloat(arg, 32)
		if err == nil {
			return reflect.ValueOf(float32(val)).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.Float64:
		val, err := strconv.ParseFloat(arg, 64)
		if err == nil {
			return reflect.ValueOf(val).Convert(argType), err
		}
		return reflect.ValueOf((interface{})(nil)), err
	case reflect.String:
		return reflect.ValueOf(arg).Convert(argType), nil
	case reflect.Slice:
		ref := reflect.New(argType)
		ref.Elem().Set(reflect.MakeSlice(argType, 0, 0))
		if err := json.Unmarshal([]byte(arg), ref.Interface()); err != nil {
			return reflect.ValueOf((interface{})(nil)), err
		}
		return ref.Elem().Convert(argType), nil
	case reflect.Map:
		ref := reflect.New(argType)
		ref.Elem().Set(reflect.MakeMap(argType))
		if err := json.Unmarshal([]byte(arg), ref.Interface()); err != nil {
			return reflect.ValueOf((interface{})(nil)), err
		}
		return ref.Elem().Convert(argType), nil
	case reflect.Struct:
		var obj interface{}
		if err := json.Unmarshal([]byte(arg), &obj); err != nil {
			return reflect.ValueOf((interface{})(nil)), err
		}
		inputArgMap := obj.(map[string]interface{})
		for i := 0; i < argType.NumField(); i++ {
			mandatoryTagValue := argType.Field(i).Tag.Get("mandatory")
			derivedtagValue := argType.Field(i).Tag.Get("derived")
			idTagValue := argType.Field(i).Tag.Get("id")
			finalTagValue := argType.Field(i).Tag.Get("final")
			if mandatoryTagValue == "true" && finalTagValue == "" {
				if derivedtagValue == "" {
					_, ok := inputArgMap[argType.Field(i).Name]
					if !ok {
						_, ok2 := inputArgMap[makeFirstLetterLowerCaps(argType.Field(i).Name)]
						if !ok2 {
							return reflect.ValueOf((interface{})(nil)), fmt.Errorf("mandatory field %s in asset %s is not present in the input", argType.Field(i).Name, strings.Split(argType.String(), ".")[1])
						}
					}
				} else {
					if idTagValue != "true" {
						return reflect.ValueOf((interface{})(nil)), fmt.Errorf("field %s is a derived field but it is not an id field. Derived key is only supported on id field", argType.Field(i).Name)
					}
				}
			}
		}
		ref := reflect.New(argType)
		if err := defaults.Set(ref.Interface()); err != nil {
			return reflect.ValueOf((interface{})(nil)), err
		}
		if err := json.Unmarshal([]byte(arg), ref.Interface()); err != nil {
			return reflect.ValueOf((interface{})(nil)), err
		}
		if _, err := SetFinaTagInput(ref.Interface()); err != nil {
			return reflect.ValueOf((interface{})(nil)), err
		}
		m := make(ErrorMap)
		DeepValidate(ref.Elem(), m, func() string {
			return ""
		}, reflect.ValueOf(inputArgMap).Interface())

		if len(m) > 0 {
			return reflect.ValueOf((interface{})(nil)), reflect.ValueOf(m).Interface().(error)
		}
		val := ref.Elem().Convert(argType)
		return val, nil
	case reflect.Ptr:
		ref := reflect.New(argType.Elem())
		return ref.Elem().Convert(argType), nil
	case reflect.Interface:
		return reflect.ValueOf(arg), nil
	default:
		return reflect.ValueOf((interface{})(nil)), fmt.Errorf(("argument Parsing/Validation failed: argument kind does not match supported kinds"))
	}
}

func SetFinaTagInput(inputStruct interface{}) (interface{}, error) {
	st := reflect.TypeOf(inputStruct).Elem()
	sv := reflect.ValueOf(inputStruct).Elem()

	nfields := st.NumField()
	for i := 0; i < nfields; i++ {
		defaultInput := st.Field(i).Tag.Get("final")
		if defaultInput != "" {
			//defaultInputs := splitUnescapedComma(defaultInput)
			fieldName := st.Field(i).Name
			fieldType := st.Field(i).Type.String()

			data, err := convert(sv.Field(i).Kind(), defaultInput, st.Field(i).Type)
			if err != nil {
				return nil, fmt.Errorf("could not parse finalTagData %s to desired type %s for field %s err %s", defaultInput, fieldType, fieldName, err.Error())
			}
			sv.Field(i).Set(data)
		}
	}
	return nil, nil
}

func processArgs(inputArgTypes reflect.Type, args []string, functionName string) ([]reflect.Value, error) {
	result := make([]reflect.Value, inputArgTypes.NumIn())

	if inputArgTypes.NumIn() != len(args) {
		if functionName == "Init" && len(args) == 1 && args[0] == "" {
			dummyresult := make([]reflect.Value, 0)
			return dummyresult, nil
		}
		return nil, fmt.Errorf("number of input arguments required by the function %s are %d, which did not match the number arguments passed i.e %d", functionName, inputArgTypes.NumIn(), len(args))
	}

	for i := 0; i < inputArgTypes.NumIn(); i++ {
		response, err := convert(inputArgTypes.In(i).Kind(), args[i], inputArgTypes.In(i))
		if err == nil {
			result[i] = response
		} else {
			return nil, err
		}
	}
	return result, nil
}

// ExecuteMethod calls a method with the given name on the provided reciever
func ExecuteMethod(obj interface{}, function string, stub shim.ChaincodeStubInterface, args []string) peer.Response {
	methodValue := reflect.ValueOf(obj).MethodByName(function)
	if !methodValue.IsValid() {
		return shim.Error(fmt.Sprintf("ExecuteMethod: No method found by given name - %s", function))
	}
	convertedArgs, err := processArgs(methodValue.Type(), args, function)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error in argument parsing and validation Detailed Error : %s", err.Error()))
	}
	result := methodValue.Call(convertedArgs)
	resultError := result[1].Interface()
	if resultError != nil {
		return shim.Error(fmt.Sprintf("ExecuteMethod: Error: %s", resultError.(error).Error()))
	}
	returnObj := result[0].Interface()
	returnBytes, errMarshal := json.Marshal(returnObj)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("ExecuteMethod: Marshalling response Error: %s", errMarshal.Error()))
	}
	return shim.Success(returnBytes)
}

func ValidateOrgAndUser(org_id string, user_id string) error {
	if org_id == "" {
		return fmt.Errorf("org_id cannot be empty")
	}
	if user_id == "" {
		return fmt.Errorf("user_id cannot be empty")
	}
	expression, err := regexp.Compile(RegexForUserAndOrgId)
	if err != nil {
		return fmt.Errorf("validating ids failed, regex could not be compiled %s", err.Error())
	}
	//Regex match
	if !expression.MatchString(org_id) {
		return fmt.Errorf("org id can not be empty and it should start with alphanumeric and can include these symbols '-', '_', '.' and '@' %s", org_id)
	}
	if !expression.MatchString(user_id) {
		return fmt.Errorf("user id can not be empty and it should start with alphanumeric and can include these symbols '-', '_', '.' and '@' %s", user_id)
	}
	return nil
}

func GetTokenId(tokenAsset interface{}) (string, error) {
	typ := reflect.TypeOf(tokenAsset).Elem()
	name := "Token_id"
	structValue := reflect.ValueOf(tokenAsset).Elem()
	_, ok := typ.FieldByName(name)
	if !ok {
		return "", fmt.Errorf("tokenId field is missing from the %s asset", typ.Name())
	}
	val := structValue.FieldByName(name)
	token_id := val.String()

	return token_id, nil
}

func GetTokenName(tokenAsset interface{}) (string, error) {
	typ := reflect.TypeOf(tokenAsset).Elem()
	name := "Token_name"
	structValue := reflect.ValueOf(tokenAsset).Elem()
	_, ok := typ.FieldByName(name)
	if !ok {
		return "", fmt.Errorf("Token_name field is missing from the %s asset", typ.Name())
	}
	val := structValue.FieldByName(name)
	token_name := val.String()

	return token_name, nil
}

func Getmd5Hash(input string) string {
	h := md5.New()
	h.Write([]byte(input))
	k := h.Sum(nil)
	return fmt.Sprintf("%x", k)
}

// this is a replica from tokenModel.go.
func GetDecimals(value float64) int {
	s := strconv.FormatFloat(value, 'f', -1, 64)

	stringSplitted := strings.Split(s, ".")
	if len(stringSplitted) == 1 {
		return 0
	}

	return len(stringSplitted[1])
}
