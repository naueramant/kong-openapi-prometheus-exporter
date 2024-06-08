package swagger

import (
	"fmt"
	"regexp"
	"strings"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func makeRegexFromPath(part string, parameters []*v3.Parameter) (*regexp.Regexp, error) {
	part = strings.Trim(part, "{}")

	for _, param := range parameters {
		if param.In != "path" {
			continue
		}

		if part == param.Name {
			return paramToRegex(param), nil
		}
	}

	// Not found
	return nil, fmt.Errorf("parameter %s not found in parameters list", part)
}

func paramToRegex(param *v3.Parameter) *regexp.Regexp {
	types := param.Schema.Schema().Type

	// If the parameter has no type, it can be anything
	if len(types) == 0 {
		return regexp.MustCompile(`.*`)
	}

	var re *regexp.Regexp

	for _, t := range types {
		switch t {
		case "string":
			re = combineRegexesWithOr(re, regexp.MustCompile(`.*`))
		case "number":
			re = combineRegexesWithOr(re, regexp.MustCompile(`\d+`))
		case "integer":
			re = combineRegexesWithOr(re, regexp.MustCompile(`\d+`))
		case "boolean":
			re = combineRegexesWithOr(re, regexp.MustCompile(`(true|false)`))
		default:
			//  If the type is not supported, it can be anything
			return regexp.MustCompile(`.*`)
		}
	}

	return re
}

func combineRegexesWithOr(re1 *regexp.Regexp, re2 *regexp.Regexp) *regexp.Regexp {
	if re1 == nil {
		return re2
	}

	combined := fmt.Sprintf("(%s|%s)", re1.String(), re2.String())

	return regexp.MustCompile(combined)
}
