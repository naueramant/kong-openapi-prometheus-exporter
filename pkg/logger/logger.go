package logger

import (
	"api-usage/pkg/kong"
	"api-usage/pkg/logger/transformers"
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
	zap          *zap.Logger
	fields       []Field
	transformers map[string]transformers.Transformer
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

	transformers := map[string]transformers.Transformer{
		"jwt": transformers.NewJWT(),
		"url": transformers.NewURL(),
	}

	for _, field := range fields {
		if err := validateFieldConfig(field, transformers); err != nil {
			return nil, err
		}
	}

	return &Logger{
		zap:          zap,
		fields:       fields,
		transformers: transformers,
	}, nil
}

func (logger *Logger) Log(log kong.Log) {
	zapFields := make([]zap.Field, 0)

	for _, field := range logger.fields {
		value := log.Get(field.Property)
		if value == nil {
			continue
		}

		if transformer := field.Transformer; transformer != "" {
			transformer := logger.transformers[transformer]
			value = transformer.Transform(value, field.With)
		}

		if value == nil || value == "" {
			continue
		}

		zapFields = append(zapFields, zap.Any(field.Name, value))
	}

	zapFields = removeDuplicatedFields(zapFields)

	logger.zap.Info("", zapFields...)
}

var (
	ErrInvalidFieldName     = errors.New("invalid field name")
	ErrInvalidFieldProperty = errors.New("invalid field property")
	ErrInvalidTransformer   = errors.New("invalid transformer")
)

func validateFieldConfig(field Field, transformers map[string]transformers.Transformer) error {
	// Name should not be empty
	if field.Name == "" {
		return ErrInvalidFieldName
	}

	// Property should not be empty
	if field.Property == "" {
		return ErrInvalidFieldProperty
	}

	// Transformer should be valid
	if field.Transformer != "" {
		if _, ok := transformers[field.Transformer]; !ok {
			return ErrInvalidTransformer
		}

		if err := transformers[field.Transformer].ValidateWith(field.With); err != nil {
			return err
		}
	}

	return nil
}

// The last field with the same key will be kept
func removeDuplicatedFields(fields []zap.Field) []zap.Field {
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
