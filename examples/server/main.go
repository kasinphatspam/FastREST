//go:build ignore

package main

import (
	"log"

	"fastrest"
)

func main() {
	app := fastrest.New(&fastrest.Config{
		Addr:          ":8080",
		Banner:        true,
		HealthCheck:   true,
		Metrics:       true,
		RequestLogger: true,
	})

	app.GET("/", func(c *fastrest.Ctx) error {
		c.GetLogger().Info("handling root request")
		return c.JSON(fastrest.StatusOK, map[string]string{
			"message": "Welcome to FastREST!",
		})
	})

	app.GET("/ping", func(c *fastrest.Ctx) error {
		c.GetLogger().Debug("ping endpoint called")
		return c.JSON(fastrest.StatusOK, map[string]string{
			"message": "pong",
		})
	})

	users := app.Group("/users")
	users.GET("", func(c *fastrest.Ctx) error {
		c.GetLogger().Info("fetching all users")
		return c.JSON(fastrest.StatusOK, []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
		})
	})

	users.GET("/:id", func(c *fastrest.Ctx) error {
		id := c.Param("id")
		c.GetLogger().Info("fetching user", "id", id)
		return c.JSON(fastrest.StatusOK, map[string]interface{}{
			"id":    id,
			"name":  "John Doe",
			"email": "john@example.com",
		})
	})

	users.POST("", func(c *fastrest.Ctx) error {
		var user map[string]interface{}
		if err := c.BodyParser(&user); err != nil {
			c.GetLogger().Error("failed to parse user body", "error", err.Error())
			return c.BadRequest("invalid JSON")
		}
		user["id"] = 3
		c.GetLogger().Info("created new user", "name", user["name"])
		return c.Created(user)
	})

	users.PUT("/:id", func(c *fastrest.Ctx) error {
		id := c.Param("id")
		var user map[string]interface{}
		if err := c.BodyParser(&user); err != nil {
			c.GetLogger().Error("failed to parse user body", "error", err.Error())
			return c.BadRequest("invalid JSON")
		}
		user["id"] = id
		c.GetLogger().Info("updated user", "id", id)
		return c.OK(user)
	})

	users.PATCH("/:id", func(c *fastrest.Ctx) error {
		id := c.Param("id")
		var updates map[string]interface{}
		if err := c.BodyParser(&updates); err != nil {
			c.GetLogger().Error("failed to parse patch body", "error", err.Error())
			return c.BadRequest("invalid JSON")
		}
		c.GetLogger().Info("patched user", "id", id)
		return c.OK(map[string]interface{}{
			"id":      id,
			"updated": updates,
		})
	})

	users.DELETE("/:id", func(c *fastrest.Ctx) error {
		id := c.Param("id")
		c.GetLogger().Info("deleted user", "id", id)
		return c.NoContent()
	})

	// Basic Auth - using validator function
	basicAuth := app.Group("/auth/basic")
	basicAuth.Use(fastrest.BasicAuth(func(username, password string) bool {
		users := map[string]string{
			"admin": "password123",
			"user":  "user123",
		}
		if pass, ok := users[username]; ok {
			return pass == password
		}
		return false
	}))
	basicAuth.GET("/profile", func(c *fastrest.Ctx) error {
		auth := c.GetAuth()
		c.GetLogger().Info("basic auth success", "username", auth.Username)
		return c.JSON(fastrest.StatusOK, map[string]interface{}{
			"message":  "authenticated with basic auth",
			"username": auth.Username,
		})
	})

	// Bearer Auth - using validator function
	bearerAuth := app.Group("/auth/bearer")
	bearerAuth.Use(fastrest.BearerAuth(func(token string) bool {
		validTokens := map[string]bool{
			"secret-token-123": true,
			"another-token":    true,
		}
		return validTokens[token]
	}))
	bearerAuth.GET("/profile", func(c *fastrest.Ctx) error {
		auth := c.GetAuth()
		c.GetLogger().Info("bearer auth success", "token", auth.Value)
		return c.JSON(fastrest.StatusOK, map[string]interface{}{
			"message": "authenticated with bearer token",
			"token":   auth.Value,
		})
	})

	// API Key Auth - using validator function
	apiKeyAuth := app.Group("/auth/apikey")
	apiKeyAuth.Use(fastrest.APIKeyAuth(func(key string) bool {
		validKeys := map[string]bool{
			"api-key-xyz":     true,
			"api-key-testing": true,
		}
		return validKeys[key]
	}, "X-API-Key"))
	apiKeyAuth.GET("/profile", func(c *fastrest.Ctx) error {
		auth := c.GetAuth()
		c.GetLogger().Info("api key auth success", "key", auth.Value)
		return c.JSON(fastrest.StatusOK, map[string]interface{}{
			"message": "authenticated with API key",
			"key":     auth.Value,
		})
	})

	// Combined Auth - supports multiple auth methods
	authConfig := fastrest.NewAuthConfig().
		SetBasicValidator(func(username, password string) bool {
			return username == "admin" && password == "password123"
		}).
		SetBearerValidator(func(token string) bool {
			return token == "secret-token-123"
		}).
		SetAPIKeyValidator(func(key string) bool {
			return key == "api-key-xyz"
		})

	combinedAuth := app.Group("/auth/any")
	combinedAuth.Use(fastrest.Auth(authConfig))
	combinedAuth.GET("/profile", func(c *fastrest.Ctx) error {
		auth := c.GetAuth()
		c.GetLogger().Info("combined auth success", "type", auth.Type)
		return c.JSON(fastrest.StatusOK, map[string]interface{}{
			"message":   "authenticated",
			"auth_type": auth.Type,
		})
	})

	external := app.Group("/external")
	external.GET("/info", func(c *fastrest.Ctx) error {
		c.GetLogger().Debug("external info requested")
		return c.JSON(fastrest.StatusOK, map[string]interface{}{
			"service": "FastREST Demo",
			"version": "1.0.0",
			"uptime":  app.Uptime().String(),
		})
	})

	if err := app.Listen(); err != nil {
		log.Fatal("Server error:", err)
	}
}
