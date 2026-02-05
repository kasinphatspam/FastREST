package middlewares

import (
	"fmt"
	"time"

	"fastrest/constant"
	"fastrest/context"
)

func RequestLogger() context.Middleware {
	return func(next context.Handler) context.Handler {
		return func(c *context.Ctx) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			status := c.Response.StatusCode()
			if status == 0 {
				status = 200
			}

			method := c.Method()
			path := c.Path()
			ip := c.IP()

			now := time.Now().Format("15:04:05")
			statusColor := getStatusColor(status)
			methodColor := getMethodColor(method)

			fmt.Printf("%s%s%s | %sREQ%s | %s%-7s%s | %s%3d%s | %12v | %s | %s\n",
				constant.ColorGray, now, constant.ColorReset,
				constant.ColorWhite, constant.ColorReset,
				methodColor, method, constant.ColorReset,
				statusColor, status, constant.ColorReset,
				duration,
				ip,
				path)

			return err
		}
	}
}

func getStatusColor(status int) string {
	switch {
	case status >= 500:
		return constant.ColorRed
	case status >= 400:
		return constant.ColorYellow
	case status >= 300:
		return constant.ColorCyan
	case status >= 200:
		return constant.ColorGreen
	default:
		return constant.ColorWhite
	}
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return constant.ColorBlue
	case "POST":
		return constant.ColorGreen
	case "PUT":
		return constant.ColorYellow
	case "DELETE":
		return constant.ColorRed
	case "PATCH":
		return constant.ColorCyan
	case "HEAD":
		return constant.ColorPurple
	case "OPTIONS":
		return constant.ColorGray
	default:
		return constant.ColorWhite
	}
}
