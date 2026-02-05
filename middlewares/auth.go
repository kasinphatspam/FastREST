package middlewares

import (
	"encoding/base64"
	"strings"

	"fastrest/context"
)

type BasicAuthValidator func(username, password string) bool
type BearerAuthValidator func(token string) bool
type APIKeyValidator func(key string) bool

type AuthConfig struct {
	BasicValidator  BasicAuthValidator
	BearerValidator BearerAuthValidator
	APIKeyValidator APIKeyValidator
	APIKeyName      string
}

func NewAuthConfig() *AuthConfig {
	return &AuthConfig{
		APIKeyName: "X-API-Key",
	}
}

func (c *AuthConfig) SetBasicValidator(v BasicAuthValidator) *AuthConfig {
	c.BasicValidator = v
	return c
}

func (c *AuthConfig) SetBearerValidator(v BearerAuthValidator) *AuthConfig {
	c.BearerValidator = v
	return c
}

func (c *AuthConfig) SetAPIKeyValidator(v APIKeyValidator) *AuthConfig {
	c.APIKeyValidator = v
	return c
}

func (c *AuthConfig) SetAPIKeyName(name string) *AuthConfig {
	c.APIKeyName = name
	return c
}

func BasicAuth(validator BasicAuthValidator) context.Middleware {
	return func(next context.Handler) context.Handler {
		return func(c *context.Ctx) error {
			auth := c.Get("Authorization")
			if auth == "" {
				c.Set("WWW-Authenticate", `Basic realm="Restricted"`)
				return c.Unauthorized("missing authorization header")
			}

			if !strings.HasPrefix(auth, "Basic ") {
				return c.Unauthorized("invalid authorization type")
			}

			decoded, err := base64.StdEncoding.DecodeString(auth[6:])
			if err != nil {
				return c.Unauthorized("invalid base64 encoding")
			}

			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) != 2 {
				return c.Unauthorized("invalid credentials format")
			}

			username, password := parts[0], parts[1]
			if !validator(username, password) {
				return c.Unauthorized("invalid credentials")
			}

			c.SetAuth(&context.AuthInfo{
				Type:     "basic",
				Username: username,
				Password: password,
				Valid:    true,
			})

			return next(c)
		}
	}
}

func BearerAuth(validator BearerAuthValidator) context.Middleware {
	return func(next context.Handler) context.Handler {
		return func(c *context.Ctx) error {
			auth := c.Get("Authorization")
			if auth == "" {
				return c.Unauthorized("missing authorization header")
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				return c.Unauthorized("invalid authorization type")
			}

			token := auth[7:]
			if !validator(token) {
				return c.Unauthorized("invalid token")
			}

			c.SetAuth(&context.AuthInfo{
				Type:  "bearer",
				Value: token,
				Valid: true,
			})

			return next(c)
		}
	}
}

func APIKeyAuth(validator APIKeyValidator, headerName string) context.Middleware {
	if headerName == "" {
		headerName = "X-API-Key"
	}
	return func(next context.Handler) context.Handler {
		return func(c *context.Ctx) error {
			key := c.Get(headerName)
			if key == "" {
				return c.Unauthorized("missing API key")
			}

			if !validator(key) {
				return c.Unauthorized("invalid API key")
			}

			c.SetAuth(&context.AuthInfo{
				Type:  "apikey",
				Value: key,
				Valid: true,
			})

			return next(c)
		}
	}
}

func Auth(config *AuthConfig) context.Middleware {
	return func(next context.Handler) context.Handler {
		return func(c *context.Ctx) error {
			auth := c.Get("Authorization")
			apiKey := c.Get(config.APIKeyName)

			if apiKey != "" && config.APIKeyValidator != nil {
				if config.APIKeyValidator(apiKey) {
					c.SetAuth(&context.AuthInfo{
						Type:  "apikey",
						Value: apiKey,
						Valid: true,
					})
					return next(c)
				}
				return c.Unauthorized("invalid API key")
			}

			if auth == "" {
				c.Set("WWW-Authenticate", `Basic realm="Restricted"`)
				return c.Unauthorized("missing authorization")
			}

			if strings.HasPrefix(auth, "Bearer ") && config.BearerValidator != nil {
				token := auth[7:]
				if config.BearerValidator(token) {
					c.SetAuth(&context.AuthInfo{
						Type:  "bearer",
						Value: token,
						Valid: true,
					})
					return next(c)
				}
				return c.Unauthorized("invalid token")
			}

			if strings.HasPrefix(auth, "Basic ") && config.BasicValidator != nil {
				decoded, err := base64.StdEncoding.DecodeString(auth[6:])
				if err != nil {
					return c.Unauthorized("invalid base64 encoding")
				}

				parts := strings.SplitN(string(decoded), ":", 2)
				if len(parts) != 2 {
					return c.Unauthorized("invalid credentials format")
				}

				username, password := parts[0], parts[1]
				if config.BasicValidator(username, password) {
					c.SetAuth(&context.AuthInfo{
						Type:     "basic",
						Username: username,
						Password: password,
						Valid:    true,
					})
					return next(c)
				}
				return c.Unauthorized("invalid credentials")
			}

			return c.Unauthorized("invalid authorization")
		}
	}
}
