package logger

import "net/url"

func URLTransformer(val string, key string) any {
	u, err := url.Parse(val)
	if err != nil {
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
