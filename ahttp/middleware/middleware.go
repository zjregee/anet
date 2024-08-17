package middleware

import (
	"errors"
	"net/textproto"
	"strings"

	"github.com/zjregee/anet/ahttp"
)

type valuesExtractor func(c *ahttp.Context) ([]string, error)

func valuesFromHeader(header string, valuePrefix string) valuesExtractor {
	prefixLen := len(valuePrefix)
	header = textproto.CanonicalMIMEHeaderKey(header)
	return func(c *ahttp.Context) ([]string, error) {
		values := c.Request().Header.Values(header)
		if len(values) == 0 {
			return nil, errors.New("header not found")
		}
		result := make([]string, 0)
		for _, value := range values {
			if prefixLen == 0 {
				result = append(result, value)
				continue
			}
			if len(value) > prefixLen && strings.EqualFold(value[:prefixLen], valuePrefix) {
				result = append(result, value[prefixLen:])
			}
		}
		if len(result) == 0 {
			return nil, errors.New("value not found")
		}
		return result, nil
	}
}
