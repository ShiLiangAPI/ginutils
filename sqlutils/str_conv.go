package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// parseQueryValue 把字符串 "false,bool"  转换为 false
func parseStringValue(val string) (any, bool) {
	if strings.Index(val, ",") > -1 {
		valTrList := strings.SplitN(val, ",", 2)
		valTr := valTrList[0]
		valTrType := valTrList[1]
		newVal, err := queryTranslateType(valTr, valTrType)
		if err != nil {
			return nil, false
		}
		return newVal, true
	} else {
		return val, true
	}
}

// 把字符串；例如 "true" "bool" 转换成 true
func queryTranslateType(value, valueType string) (any, error) {
	switch strings.ToLower(valueType) {
	case "int":
		return strconv.Atoi(value)
	case "int64":
		value, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		} else {
			return int64(value), err
		}
	case "bool":
		return strconv.ParseBool(value)
	case "slice", "list":
		return strings.Split(value, ","), nil
	case "like":
		return fmt.Sprintf("%%%s%%", value), nil
	case "llike":
		return fmt.Sprintf("%%%s", value), nil
	case "rlike":
		return fmt.Sprintf("%s%%", value), nil
	default:
		return value, nil
	}
}
