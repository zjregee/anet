package middleware

import (
	"net/http"

	"github.com/zjregee/anet/ahttp"
)

func CORS() ahttp.MiddlewareFunc {
	return func(next ahttp.HandlerFunc) ahttp.HandlerFunc {
		return func(c *ahttp.Context) (returnErr error) {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if c.Request().Method == "OPTIONS" {
				c.Response().WriteHeader(http.StatusNoContent)
				return
			}
			returnErr = next(c)
			return
		}
	}
}
