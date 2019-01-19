package rollbar

import (
	"fmt"
	"runtime"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/stvp/roll"
)

func Recover() echo.MiddlewareFunc {
	return RecoverWithConfig(middleware.DefaultRecoverConfig)
}

// Recovery middleware for rollbar error monitoring
func RecoverWithConfig(config middleware.RecoverConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultRecoverConfig.Skipper
	}
	if config.StackSize == 0 {
		config.StackSize = middleware.DefaultRecoverConfig.StackSize
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					if !config.DisablePrintStack {
						c.Logger().Printf("[PANIC RECOVER] %v %s\n", err, stack[:length])
					}

					roll.CriticalStack(err, getCallers(3), map[string]string{
						"endpoint": c.Request().RequestURI,
					})

					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}

func getCallers(skip int) (pc []uintptr) {
	pc = make([]uintptr, 1000)
	i := runtime.Callers(skip+1, pc)
	return pc[0:i]
}
