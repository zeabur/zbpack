package php

import (
	"strconv"

	"github.com/zeabur/zbpack/pkg/types"
)

// PropertyToString serializes a PHPProperty.
func PropertyToString(p types.PHPProperty) string {
	return strconv.FormatUint(uint64(p), 16)
}

// PropertyFromString deserializes a property.
// It must be the serialized result from PropertyToString.
func PropertyFromString(s string) types.PHPProperty {
	i, _ := strconv.ParseUint(s, 16, 64)
	return types.PHPProperty(i)
}
