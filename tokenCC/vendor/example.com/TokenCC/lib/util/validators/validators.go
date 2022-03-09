package validators

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"example.com/TokenCC/lib/util/date"

	"gopkg.in/validator.v2"
)

var validatorsInitialised bool = false

// ValidatorMapping conatains mapping of validation string with the required functions
// If a user should want to add his own validation functions then
// he should define his own validation functions and mention in the map
var ValidatorMapping = map[string]validator.ValidationFunc{
	BooleanTag:   checkBoolean,
	IntegerTag:   checkInteger,
	FloatTag: 	  checkFLoat64,
	StringTag:    checkString,
	NumericTag:   checkNumeric,
	PositiveTag:  checkPositive,
	DateTag:      checkDate,
	MaxDateTag:   checkMaxDate,
	MinDateTag:   checkMinDate,
	URLTag:       checkURL,
	EmailTag:     checkEmail,
	MandatoryTag: checkMandatory,
	ArrayTag:     checkArray,
	RangeTag:     checkRange,
}

func initializeValidators() {
	for key, value := range ValidatorMapping {
		validator.SetValidationFunc(key, value)
	}
}

var sepPattern *regexp.Regexp = regexp.MustCompile(`((?:^|[^\\])(?:\\\\)*),`)

func splitUnescapedComma(str string) []string {
	ret := []string{}
	indexes := sepPattern.FindAllStringIndex(str, -1)
	last := 0
	for _, is := range indexes {
		ret = append(ret, str[last:is[1]-1])
		last = is[1]
	}
	ret = append(ret, str[last:])
	return ret
}

// Validate validates an independent value against the struct tags passed as string
func Validate(input interface{}, param string) error {
	if validatorsInitialised != true {
		initializeValidators()
		validatorsInitialised = true
	}
	tags := splitUnescapedComma(param)
	var regexString string
	for i := 0; i < len(tags); i++ {
		if strings.Contains(tags[i], "regex") {
			regexString = tags[i]
		}
	}
	errs := validator.Valid(input, param)
	if errs != nil {
		if strings.Contains(errs.Error(), "regular expression mismatch") {
			str2 := fmt.Sprintf("Doesn't match the regular expression %s", regexString)
			newErrorString := strings.Replace(errs.Error(), "regular expression mismatch", str2, 1)
			return fmt.Errorf("Validation failed for the value %v with error %s", input, newErrorString)
		}
		return fmt.Errorf("Validation failed for the value %v with error: %s", input, errs.Error())
	}
	return nil
}

// ValidateStruct validates a struct which has defined valid tags mentioned on its fields
func ValidateStruct(nur interface{}) error {
	if validatorsInitialised != true {
		initializeValidators()
		validatorsInitialised = true
	}
	if errs := validator.Validate(nur); errs != nil {
		// fmt.Println(errs)
		return fmt.Errorf("Validation has failed with following errors %s", errs.Error())
	}
	return nil
}

func checkBoolean(input interface{}, param string) error {
	inputValue := reflect.TypeOf(input)
	if inputValue.Kind() != reflect.Bool {
		return fmt.Errorf("Boolean Validation Error: input is not boolean %v", input)
	}
	return nil
}

func checkMandatory(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	errs := validator.Valid(inputValue, "nonzero")
	if errs != nil {
		return fmt.Errorf("Mandatory Validation Failed: this field is mandatory cannot be empty")
	}
	return nil
}

func checkNumeric(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.String {
		return fmt.Errorf("Numeric Validator: input is not a string %v", input)
	}
	s := inputValue.String()
	re := regexp.MustCompile("^[0-9]+$")
	if re.MatchString(s) != true {
		return fmt.Errorf("Numeric Validation: input is not numeric %s", input)
	}
	return nil
}

func checkPositive(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.String {
		return fmt.Errorf("Positive Validation: input is not string %v ", input)
	}
	s := inputValue.String()
	i1, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("Positive Validation: input is not numeric %s", input)
	}
	if i1 < 0 {
		return fmt.Errorf("Positive Validation: input is not positive %s", input)
	}
	return nil
}

func checkString(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.String {
		return fmt.Errorf("String Validation Fail: input is not a string %v", input)
	}
	return nil
}

func checkInteger(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.Int {
		return fmt.Errorf("Integer Validation Failed: input is not integer %v", input)
	}
	return nil
}

func checkFLoat64(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.Float64 {
		return fmt.Errorf("Integer Validation Failed: input is not float64 %v", input)
	}
	return nil
}

func checkURL(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.String {
		return fmt.Errorf("Url Validation Error: input url is not a string %v", input)
	}
	inputString := inputValue.String()
	parseURL, err := url.Parse(inputString)
	if err != nil {
		return fmt.Errorf("Url Validation Error: not a valid url %s parse error %s", inputString, err.Error())
	}
	if parseURL.Scheme == "" || parseURL.Host == "" || parseURL.Path == "" {
		return fmt.Errorf("Url Validation Error: not a valid url %s", inputString)
	}
	return nil
}

func checkEmail(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.String {
		return fmt.Errorf("Invalid Email: input in not string %v", input)
	}

	s := inputValue.String()
	re := regexp.MustCompile(RegexEmail)

	if re.MatchString(s) != true {
		return fmt.Errorf("Email Validation Error: invalid email %s", inputValue.String())
	}
	return nil
}

func checkArray(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.Slice {
		return fmt.Errorf("Array Validation Error: value is not a slice type %s", inputValue)
	}
	return nil
}

func checkRange(input interface{}, param string) error {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.Slice {
		return fmt.Errorf("Range Validation Error: value is not a slice type %s", inputValue)
	}
	split := strings.Split(param, "-")
	var max, min int
	var err error
	if split[0] != "" {
		min, err = strconv.Atoi(split[0])
		if err != nil {
			return fmt.Errorf("Range Validation Error: failed conversion to int for lowerbound argument given as %d expecting a number", min)
		}
	} else {
		min = -1
	}
	if split[1] != "" {
		max, err = strconv.Atoi(split[1])
		if err != nil {
			return fmt.Errorf("Range Validation Error: failed conversion to int for upperbound argument given as %d expecting a number", max)
		}
	} else {
		max = -1
	}

	length := inputValue.Len()
	// fmt.Println("Length of the Array", length)
	if min != -1 && max != -1 {
		if length < min || length > max {
			return fmt.Errorf("Range Validation Error: the size of array is out of range, min-%d max-%d, but size is-%d", min, max, length)
		}
		return nil
	} else if min != -1 {
		//Compare with min only
		if length < min {
			return fmt.Errorf("Range Validation Error: the size of array is smaller than min, min: %d but size is: %d", min, length)
		}
		return nil
	} else if max != -1 {
		//Compare with max only
		if length > max {
			return fmt.Errorf("Range Validation Error: the size of array is greater than max, max: %d but size is: %d", max, length)
		}
		return nil
	} else {
		fmt.Println("The range is not explicitly mentioned")
		return nil
	}
}

func parseTime(param string) (time.Time, error) {
	inputTime, errCustomParsing := time.Parse(date.CustomDateLayout, param)
	if errCustomParsing == nil {
		return inputTime, nil
	}
	inputTime, errRfcParsing := time.Parse(time.RFC3339, param)
	if errRfcParsing == nil {
		return inputTime, nil
	}
	return time.Time{}, fmt.Errorf("Invalid date %s", param)
}

func checkDate(input interface{}, param string) error {
	inputValue := reflect.TypeOf(input)
	if inputValue.String() != "date.Date" {
		return fmt.Errorf("Date Validation Error %v", input)
	}
	return nil
}

func checkMinDate(input interface{}, param string) error {
	inputValue := reflect.TypeOf(input)
	if inputValue.String() != "date.Date" {
		return fmt.Errorf("Date Validation Error : input is not a valid date %v", input)
	}
	minTime, err := parseTime(param)
	if err != nil {
		return fmt.Errorf("Date Validation Error: min date %s", err.Error())
	}
	givenDate := input.(date.Date)
	givenTime := givenDate.Time
	dateString, err := givenDate.String()
	if err != nil {
		return fmt.Errorf("Date Validation Error: %s", err.Error())
	}
	if givenTime.After(minTime) != true {
		return fmt.Errorf("Date Validation Error: date is not greater than min date %s, and given date is %s", minTime.String(), dateString)
	}
	return nil
}

func checkMaxDate(input interface{}, param string) error {
	inputValue := reflect.TypeOf(input)
	if inputValue.String() != "date.Date" {
		return fmt.Errorf("Date Validation Error : not a valid date %v", input)
	}
	maxTime, err := parseTime(param)
	if err != nil {
		return fmt.Errorf("Date Validation Error: max date %s", err.Error())
	}
	givenDate := input.(date.Date)
	givenTime := givenDate.Time
	dateString, err := givenDate.String()
	if err != nil {
		return fmt.Errorf("Date Validation Error: %s", err.Error())
	}
	if givenTime.Before(maxTime) != true {
		return fmt.Errorf("Date Validation Error: date is not lesser than max date is %s, given time is %s", maxTime.String(), dateString)
	}
	return nil
}
