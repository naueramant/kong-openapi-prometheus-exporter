package logger

import (
	"api-usage/pkg/kong"
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field struct {
	Name        string
	Property    string
	Transformer string
	With        map[string]string
}

type Logger struct {
	zap    *zap.Logger
	fields []Field
}

func New(fields []Field) (*Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.MessageKey = zapcore.OmitKey
	cfg.EncoderConfig.LevelKey = zapcore.OmitKey
	cfg.EncoderConfig.TimeKey = zapcore.OmitKey
	cfg.EncoderConfig.CallerKey = zapcore.OmitKey

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zap, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		if err := validateFieldConfig(field); err != nil {
			return nil, err
		}
	}

	return &Logger{
		zap:    zap,
		fields: fields,
	}, nil
}

func (logger *Logger) Log(log kong.Log) {
	zapFields := make([]zap.Field, 0)

	for _, field := range logger.fields {
		value := log.Get(field.Property)
		if value == nil {
			continue
		}

		if field.Transformer == "jwt" {
			key, _ := field.With["key"]
			typ, _ := field.With["type"]

			value = JWTTransformer(value.(string), key, typ)
		}

		if field.Transformer == "url" {
			key, _ := field.With["key"]

			value = URLTransformer(value.(string), key)
		}

		if value == nil || value == "" {
			continue
		}

		zapFields = append(zapFields, zap.Any(field.Name, value))
	}

	zapFields = removeDuplicates(zapFields)

	logger.zap.Info("", zapFields...)
}

var ErrInvalidFieldName = errors.New("invalid field name")
var ErrInvalidFieldProperty = errors.New("invalid field property")
var ErrInvalidFieldTransformer = errors.New("invalid field transformer")
var ErrMissingWithKey = errors.New("missing key in With field")

var allowedTransformers = []string{"", "jwt", "url"}
var allowedURLKeys = []string{"scheme", "host", "path", "query", "fragment"}

func validateFieldConfig(field Field) error {
	// Name should not be empty
	if field.Name == "" {
		return ErrInvalidFieldName
	}

	// Property should not be empty
	if field.Property == "" {
		return ErrInvalidFieldProperty
	}

	// Transformer should be one of the allowed transformers
	if !contains(allowedTransformers, field.Transformer) {
		return ErrInvalidFieldTransformer
	}

	// If transformer is jwt, there should be no "With" field containing a non-empty value "key"
	if transformer := field.Transformer; transformer == "jwt" {
		if field.With != nil {
			if key, ok := field.With["key"]; !ok || key == "" {
				return ErrMissingWithKey
			}
		}
	}

	// If transformer is url, there should be no "With" field containing a non-empty value "key" with a allowed value
	if transformer := field.Transformer; transformer == "url" {
		if field.With != nil {
			if key, ok := field.With["key"]; ok && key != "" {
				if !contains(allowedURLKeys, key) {
					return ErrMissingWithKey
				}
			}
		}
	}

	return nil
}

func contains(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}

	return false
}

// Remove duplicate fields from the slice
// The last field with the same key will be kept
func removeDuplicates(fields []zap.Field) []zap.Field {
	seen := make(map[string]struct{})
	result := make([]zap.Field, 0)

	for i := len(fields) - 1; i >= 0; i-- {
		field := fields[i]

		if _, ok := seen[field.Key]; !ok {
			seen[field.Key] = struct{}{}
			result = append([]zap.Field{field}, result...)
		}
	}

	return result
}
