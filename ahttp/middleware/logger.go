package middleware

import (
	"fmt"

	"github.com/zjregee/anet/ahttp"
)

func Logger() ahttp.MiddlewareFunc {
	return func(next ahttp.HandlerFunc) ahttp.HandlerFunc {
		return func(c *ahttp.Context) (returnErr error) {
			returnErr = next(c)
			fmt.Printf("[MIDDLEWARE LOGGER]: %s %s %v\n", c.Request().Method, c.Request().URL.Path, returnErr)
			return
		}
	}
}
