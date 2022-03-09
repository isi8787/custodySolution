package date

import (
	"fmt"
	"strings"
	"time"
)

const (
	// CustomDateLayout is the custom layout specifier for time.Time object. It is used while parsing a given time string in this layout i.e. YYYY-MM-DD.
	CustomDateLayout = "2006-01-02"
)

//Date object. This holds time
type Date struct {
	time.Time
}

func (d Date) String() (string, error) {
	return d.Time.String(), nil
}

// UnmarshalJSON is the implementation of Unmarshaller interface for Date
func (d *Date) UnmarshalJSON(input []byte) error {
	strInput := string(input)
	length := len(strInput)
	if length < 2 || strInput[0] != '"' || strInput[length-1] != '"' {
		return fmt.Errorf("Date Unmarshal Error: missing double quotes (%s)", strInput)
	}
	strInput = strings.Trim(strInput, `"`)
	inputTime, customParsingError := time.Parse(CustomDateLayout, strInput)
	if customParsingError == nil {
		d.Time = inputTime
		return nil
	}
	inputTime, rfcParsingError := time.Parse(time.RFC3339, strInput)
	if rfcParsingError == nil {
		d.Time = inputTime
		return nil
	}
	return fmt.Errorf("Unmarshalling Error: Date is expected in YYYY-MM-DD/YYYY-MM-DDTHH:MM:SSZ format")
}

// MarshalJSON is the implementation of Marhaller interface for Date
func (d Date) MarshalJSON() ([]byte, error) {
	result, err := d.Time.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// After reports whether the date intance d is after compareDate.
func (d Date) After(compareDate Date) bool {
	return d.Time.After(compareDate.Time)
}

// Before reports whether the date instance d  is before compareDate.
func (d Date) Before(compareDate Date) bool {
	return d.Time.Before(compareDate.Time)
}
