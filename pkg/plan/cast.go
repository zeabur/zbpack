package plan

import "github.com/spf13/cast"

// ToWeakBoolE enhances ToBoolE to support more string values
// like "TRUE", "1", "FALSE", "0".
func ToWeakBoolE(i interface{}) (bool, error) {
	switch i {
	case "true", "True", "TRUE", "1":
		return true, nil
	case "false", "False", "FALSE", "0":
		return false, nil
	}
	return cast.ToBoolE(i)
}
