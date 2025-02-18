package transformers

import (
	"errors"
	"net/url"
)

var (
	ErrURLKeyNotProvided = errors.New("key not provided in URL transformer")
	ErrURLInvalidKey     = errors.New("invalid key provided in URL transformer")
)

type URLTransformer struct{}

func NewURL() *URLTransformer {
	return &URLTransformer{}
}

var allowedURLKeys = map[string]struct{}{
	"scheme":   {},
	"host":     {},
	"path":     {},
	"query":    {},
	"fragment": {},
}

func (t *URLTransformer) ValidateWith(with map[string]string) error {
	if _, ok := with["key"]; !ok {
		return ErrURLKeyNotProvided
	}

	if _, ok := allowedURLKeys[with["key"]]; !ok {
		return ErrURLInvalidKey
	}

	return nil
}

func (t *URLTransformer) Transform(value any, with map[string]string) any {
	val, ok := value.(string)
	if !ok {
		return ""
	}

	u, err := url.Parse(val)
	if err != nil {
		return ""
	}

	key, ok := with["key"]
	if !ok {
		return ""
	}

	switch key {
	case "scheme":
		return u.Scheme
	case "host":
		return u.Host
	case "path":
		return u.Path
	case "query":
		return u.RawQuery
	case "fragment":
		return u.Fragment
	default:
		return ""
	}
}
