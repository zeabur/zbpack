package utils

import (
	"encoding/json"
	"fmt"
)

// IntOrBool is a type that can be either an int or a bool
type IntOrBool struct {
	IntValue  int
	BoolValue bool
	IsInt     bool
}

// UnmarshalJSON implements json.Unmarshaler
func (iob *IntOrBool) UnmarshalJSON(data []byte) error {
	var intValue int
	if err := json.Unmarshal(data, &intValue); err == nil {
		iob.IntValue = intValue
		iob.IsInt = true
		return nil
	}

	var boolValue bool
	if err := json.Unmarshal(data, &boolValue); err == nil {
		iob.BoolValue = boolValue
		iob.IsInt = false
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into IntOrBool", data)
}
