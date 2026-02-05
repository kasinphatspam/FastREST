# FastREST

A high-performance REST API framework for Go, built on top of [fasthttp](https://github.com/valyala/fasthttp).

## Installation

```bash
go get github.com/yourusername/fastrest
```

## Quick Start

```go
package main

import (
    "fastrest"
)

func main() {
    app := fastrest.New(&fastrest.Config{
        Addr:        ":8080",
        Banner:      true,
        HealthCheck: true,
        Metrics:     true,
    })

    app.GET("/", func(c *fastrest.Ctx) error {
        return c.OK(map[string]string{"message": "Hello, World!"})
    })

    app.Listen()
}
```

## Configuration

```go
app := fastrest.New(&fastrest.Config{
    Addr:               ":8080",          // Server address
    Env:                "development",    // Environment name
    Banner:             true,             // Show startup banner
    HealthCheck:        true,             // Enable health endpoints
    HealthPath:         "/health",        // Health check path
    Metrics:            true,             // Enable metrics
    RequestLogger:      true,             // Log all requests
    ReadTimeout:        30 * time.Second, // Read timeout
    WriteTimeout:       30 * time.Second, // Write timeout
    IdleTimeout:        60 * time.Second, // Idle timeout
    GracefulTimeout:    10 * time.Second, // Graceful shutdown timeout
    MaxConnsPerIP:      0,                // Max connections per IP
    MaxRequestsPerConn: 0,                // Max requests per connection
})
```

## Routing

### Basic Routes

```go
app.GET("/users", getUsers)
app.POST("/users", createUser)
app.PUT("/users/:id", updateUser)
app.PATCH("/users/:id", patchUser)
app.DELETE("/users/:id", deleteUser)
app.HEAD("/users", headUsers)
app.OPTIONS("/users", optionsUsers)
```

### Route Parameters

```go
app.GET("/users/:id", func(c *fastrest.Ctx) error {
    id := c.Param("id")
    return c.OK(map[string]string{"id": id})
})
```

### Query Parameters

```go
app.GET("/search", func(c *fastrest.Ctx) error {
    query := c.Query("q")
    page := c.QueryDefault("page", "1")
    return c.OK(map[string]string{"query": query, "page": page})
})
```

### Route Groups

```go
api := app.Group("/api/v1")
api.GET("/users", getUsers)
api.POST("/users", createUser)

admin := app.Group("/admin")
admin.Use(authMiddleware)
admin.GET("/stats", getStats)
```

## Context Methods

### Request

```go
c.Param("id")                    // Get route parameter
c.Body()                         // Get raw body as []byte
c.BodyParser(&user)              // Parse JSON body into struct
c.Get("Content-Type")            // Get request header
c.Method()                       // Get HTTP method
c.Path()                         // Get request path
c.IP()                           // Get client IP
```

### Query Parameters

```go
// String
c.Query("name")                              // Returns string
c.QueryDefault("name", "default")            // With default value

// Integer
c.QueryInt("page")                           // Returns (int, error)
c.QueryIntDefault("page", 1)                 // With default value
c.QueryInt64("id")                           // Returns (int64, error)
c.QueryInt64Default("id", 0)                 // With default value

// Float
c.QueryFloat32("price")                      // Returns (float32, error)
c.QueryFloat32Default("price", 0.0)          // With default value
c.QueryFloat64("amount")                     // Returns (float64, error)
c.QueryFloat64Default("amount", 0.0)         // With default value

// Boolean
c.QueryBool("active")                        // Returns (bool, error)
c.QueryBoolDefault("active", false)          // With default value

// Time
c.QueryTime("date", time.RFC3339)            // Returns (time.Time, error)
c.QueryTimeDefault("date", layout, default)  // With default value

// Duration
c.QueryDuration("timeout")                   // Returns (time.Duration, error)
c.QueryDurationDefault("timeout", 30*time.Second)

// Slices (for ?ids=1,2,3)
c.QuerySlice("tags", ",")                    // Returns []string
c.QueryIntSlice("ids", ",")                  // Returns ([]int, error)
```

### Response

```go
c.JSON(200, data)                // Send JSON response
c.String(200, "text")            // Send text response
c.Status(201)                    // Set status code (chainable)
c.Set("X-Custom", "value")       // Set response header
c.Redirect("/new-path", 302)     // Redirect
c.SendFile("/path/to/file")      // Send file
c.NoContent()                    // 204 No Content
```

### Convenience Methods

```go
c.OK(data)                       // 200 with JSON
c.Created(data)                  // 201 with JSON
c.BadRequest("message")          // 400 with error JSON
c.Unauthorized("message")        // 401 with error JSON
c.Forbidden("message")           // 403 with error JSON
c.NotFound("message")            // 404 with error JSON
c.InternalServerError("message") // 500 with error JSON
```

### Locals (Request-scoped data)

```go
c.SetLocal("user", user)
user := c.GetLocal("user")
```

## Middleware

### Global Middleware

```go
app.Use(loggingMiddleware)
app.Use(corsMiddleware)
```

### Route Group Middleware

```go
api := app.Group("/api")
api.Use(authMiddleware)
```

### Custom Middleware

```go
func myMiddleware(next fastrest.Handler) fastrest.Handler {
    return func(c *fastrest.Ctx) error {
        // Before handler
        c.SetLocal("start", time.Now())

        err := next(c)

        // After handler
        duration := time.Since(c.GetLocal("start").(time.Time))
        c.GetLogger().Info("request completed", "duration", duration)

        return err
    }
}
```

## Authentication

All auth middlewares use validator functions, giving you full control over validation logic.

### Basic Auth

```go
app.Use(fastrest.BasicAuth(func(username, password string) bool {
    // Your validation logic (database, config, etc.)
    return username == "admin" && password == "secret"
}))
```

### Bearer Token Auth

```go
app.Use(fastrest.BearerAuth(func(token string) bool {
    // Validate JWT, check database, etc.
    return token == "valid-token"
}))
```

### API Key Auth

```go
app.Use(fastrest.APIKeyAuth(func(key string) bool {
    // Check against database, config, etc.
    return key == "my-api-key"
}, "X-API-Key"))
```

### Combined Auth (Multiple Methods)

```go
authConfig := fastrest.NewAuthConfig().
    SetBasicValidator(func(username, password string) bool {
        return username == "admin" && password == "secret"
    }).
    SetBearerValidator(func(token string) bool {
        return token == "valid-token"
    }).
    SetAPIKeyValidator(func(key string) bool {
        return key == "my-api-key"
    }).
    SetAPIKeyName("X-API-Key")

app.Use(fastrest.Auth(authConfig))
```

### Database Example

```go
app.Use(fastrest.BasicAuth(func(username, password string) bool {
    user, err := db.FindUserByUsername(username)
    if err != nil {
        return false
    }
    return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil
}))
```

### JWT Example

```go
app.Use(fastrest.BearerAuth(func(token string) bool {
    claims, err := jwt.Parse(token, secretKey)
    if err != nil {
        return false
    }
    return claims.Valid()
}))
```

### Accessing Auth Info

```go
app.GET("/profile", func(c *fastrest.Ctx) error {
    auth := c.GetAuth()
    if auth != nil && auth.Valid {
        return c.OK(map[string]string{
            "type":     auth.Type,     // "basic", "bearer", or "apikey"
            "username": auth.Username, // For basic auth
            "value":    auth.Value,    // Token or API key
        })
    }
    return c.Unauthorized("not authenticated")
})
```

## Built-in Features

### Health Checks

When `HealthCheck: true`:

```
GET /health        - Full health status with system info
GET /health/live   - Liveness probe (Kubernetes)
GET /health/ready  - Readiness probe (Kubernetes)
```

### Metrics

When `Metrics: true`:

```
GET /metrics       - Prometheus format
GET /metrics/json  - JSON format
```

### Request Logger

When `RequestLogger: true`, all requests are logged with method, path, status, and duration.

## Logging

```go
// Use built-in logger
logger := fastrest.NewLogger()
logger.Info("message", "key", "value")
logger.Debug("debug message")
logger.Warn("warning")
logger.Error("error", "err", err.Error())

// In handlers
c.GetLogger().Info("processing request")

// Custom logger
app := fastrest.New(&fastrest.Config{
    Logger: myCustomLogger,
})
```

## HTTP Status Constants

```go
fastrest.StatusOK                    // 200
fastrest.StatusCreated               // 201
fastrest.StatusNoContent             // 204
fastrest.StatusBadRequest            // 400
fastrest.StatusUnauthorized          // 401
fastrest.StatusForbidden             // 403
fastrest.StatusNotFound              // 404
fastrest.StatusInternalServerError   // 500
// ... and more
```

## Example

See full example in [examples/server/main.go](examples/server/main.go)

```bash
cd examples/server
go run main.go
```

### Test Endpoints

```bash
# Public
curl http://localhost:8080/
curl http://localhost:8080/api/v1/users
curl http://localhost:8080/api/v1/users/1

# Basic Auth
curl -u admin:secret http://localhost:8080/basic/profile

# Bearer Token
curl -H "Authorization: Bearer my-secret-token" http://localhost:8080/bearer/data

# API Key
curl -H "X-API-Key: my-api-key" http://localhost:8080/apikey/data

# Health & Metrics
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

## License

MIT
