package logger

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
)

func JWTTransformer(val string, key string, typ string) any {
	if len(val) > 7 && val[:7] == "Bearer " {
		val = val[7:]
	}

	parts := strings.Split(val, ".")

	var v *fastjson.Value

	var partIndex int
	if strings.HasPrefix(key, "header") {
		partIndex = 0
		key = strings.TrimPrefix(key, "header.")
	} else if strings.HasPrefix(key, "payload") {
		partIndex = 1
		key = strings.TrimPrefix(key, "payload.")
	} else if strings.HasPrefix(key, "signature") {
		partIndex = 2
		key = strings.TrimPrefix(key, "signature.")
	} else {
		return ""
	}

	part := parts[partIndex]

	decoded, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return ""
	}

	v, err = fastjson.Parse(string(decoded))
	if err != nil {
		return ""
	}

	if v == nil {
		return ""
	}

	strVal := string(v.GetStringBytes(key))
	if strVal == "" {
		return nil
	}

	var castedType any
	switch typ {
	case "string":
		castedType = strVal
	case "number":
		castedType = convertStringToNumber(strVal)
	case "bool":
		castedType = convertStringToBool(strVal)
	default:
		castedType = strVal
	}

	return castedType
}

func convertStringToNumber(str string) any {
	if str == "" {
		return 0
	}

	if strings.Contains(str, ".") {
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0
		}

		return f
	}

	i, err := strconv.Atoi(str)
	if err != nil {
		return nil
	}

	return i
}

func convertStringToBool(str string) any {
	return strings.ToLower(str) == "true"
}
