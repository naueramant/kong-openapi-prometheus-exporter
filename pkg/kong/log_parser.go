package kong

import (
	"io"
	"strconv"
	"strings"

	"github.com/goccy/go-reflect"
	jsoniter "github.com/json-iterator/go"
)

type Log struct {
	ClientIP  string    `json:"client_ip"`
	Request   Request   `json:"request"`
	Response  Response  `json:"response"`
	Latencies Latencies `json:"latencies"`
}

type Request struct {
	URI         string            `json:"uri"`
	Headers     map[string]string `json:"headers"`
	Method      string            `json:"method"`
	Size        int               `json:"size"`
	URL         string            `json:"url"`
	QueryString map[string]string `json:"querystring"`
}

type Response struct {
	Size   int `json:"size"`
	Status int `json:"status"`
}

type Latencies struct {
	Request int `json:"request"`
	Proxy   int `json:"proxy"`
	Kong    int `json:"kong"`
	Receive int `json:"receive"`
}

func ParseLog(body io.Reader) (*Log, error) {
	var log Log

	err := jsoniter.NewDecoder(body).Decode(&log)
	if err != nil {
		return nil, err
	}

	return &log, nil
}

// Get returns the value of the given path in any struct
func (log *Log) Get(path string) interface{} {
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil
	}

	return getValue(reflect.ValueOf(log), parts)
}

// splitPath splits a path into parts, handling both dot notation and bracket notation
func splitPath(path string) []string {
	var parts []string
	var current strings.Builder
	var inBracket bool

	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '.':
			if !inBracket && current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else if inBracket {
				current.WriteByte(path[i])
			}
		case '[':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inBracket = true
		case ']':
			if inBracket {
				if current.Len() > 0 {
					// Strip quotes if present
					s := current.String()
					if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') {
						s = s[1 : len(s)-1]
					}
					parts = append(parts, s)
				}
				current.Reset()
			}
			inBracket = false
		case '\'', '"':
			// Skip quotes in bracket notation
			if !inBracket {
				current.WriteByte(path[i])
			}
		default:
			current.WriteByte(path[i])
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// getValue recursively gets the value at the given path
func getValue(v reflect.Value, parts []string) interface{} {
	if len(parts) == 0 {
		return v.Interface()
	}

	// Dereference pointer if needed
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	part := parts[0]
	restPath := parts[1:]

	switch v.Kind() {
	case reflect.Struct:
		field := v.FieldByName(part)
		if !field.IsValid() {
			return nil
		}
		return getValue(field, restPath)

	case reflect.Map:
		key := reflect.ValueOf(part)
		if v.Type().Key().Kind() != key.Type().Kind() {
			// Try converting string to appropriate key type
			switch v.Type().Key().Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if i, err := strconv.ParseInt(part, 10, 64); err == nil {
					key = reflect.ValueOf(i).Convert(v.Type().Key())
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if i, err := strconv.ParseUint(part, 10, 64); err == nil {
					key = reflect.ValueOf(i).Convert(v.Type().Key())
				}
			case reflect.Float32, reflect.Float64:
				if f, err := strconv.ParseFloat(part, 64); err == nil {
					key = reflect.ValueOf(f).Convert(v.Type().Key())
				}
			default:
				key = reflect.ValueOf(part).Convert(v.Type().Key())
			}
		}

		mapValue := v.MapIndex(key)
		if !mapValue.IsValid() {
			return nil
		}
		return getValue(mapValue, restPath)

	case reflect.Slice, reflect.Array:
		index, err := strconv.Atoi(part)
		if err != nil || index < 0 || index >= v.Len() {
			return nil
		}
		return getValue(v.Index(index), restPath)

	default:
		return nil
	}
}
