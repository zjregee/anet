package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/zjregee/anet/ahttp"
)

const stackSize = 4 * 1024
const disableStackAll = false

func Recover() ahttp.MiddlewareFunc {
	return func(next ahttp.HandlerFunc) ahttp.HandlerFunc {
		return func(c *ahttp.Context) (returnErr error) {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, stackSize)
					len := runtime.Stack(stack, !disableStackAll)
					stack = stack[:len]
					fmt.Printf("[MIDDLEWARE PANIC RECOVER]: %v %s\n", err, stack)
					c.Response().WriteHeader(http.StatusInternalServerError)
					_, returnErr = c.Response().Write(nil)
				}
			}()
			return next(c)
		}
	}
}
