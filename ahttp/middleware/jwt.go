package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/zjregee/anet/ahttp"
)

var ErrJWTMissing = ahttp.NewHTTPError(http.StatusBadRequest, "missing or malformed jwt")
var ErrJWTInvalid = ahttp.NewHTTPError(http.StatusUnauthorized, "invalid or expired jwt")

type JWTConfig struct {
	SigningKey    interface{}
	SigningMethod string
	ContextKey    string
	Claims        jwt.Claims
	TokenLookup   string
	AuthScheme    string
}

var DefaultJWTConfig = JWTConfig{
	SigningMethod: "HS256",
	ContextKey:    "user",
	Claims:        jwt.MapClaims{},
	TokenLookup:   "header:Authorization",
	AuthScheme:    "Bearer",
}

func JWT(key interface{}) ahttp.MiddlewareFunc {
	if key == nil {
		panic("shouldn't failed here")
	}
	config := DefaultJWTConfig
	config.SigningKey = key
	return func(next ahttp.HandlerFunc) ahttp.HandlerFunc {
		return func(c *ahttp.Context) error {
			extractor := valuesFromHeader("Authorization", config.AuthScheme)
			auths, err := extractor(c)
			if err != nil {
				return ErrJWTMissing.SetInternal(err)
			}
			auth := auths[0]
			token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
				return config.SigningKey, nil
			})
			if err != nil {
				return ErrJWTInvalid.SetInternal(err)
			}
			if !token.Valid {
				return ErrJWTInvalid
			}
			c.Set(config.ContextKey, token)
			return next(c)
		}
	}
}
