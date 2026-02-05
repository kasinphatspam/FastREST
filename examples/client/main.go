//go:build ignore

package main

import (
	"fmt"
	"log"

	"fastrest/client"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           FastREST Client Test Suite                          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	baseURL := "http://localhost:8080"
	c := client.New(baseURL)

	testPublicEndpoints(c)
	testUsersAPI(c)
	testBasicAuth(baseURL)
	testBearerAuth(baseURL)
	testAPIKeyAuth(baseURL)
	testCombinedAuth(baseURL)
	testMetrics(c)
	testErrorCases(c, baseURL)

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           All Tests Completed                                 ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
}

func testPublicEndpoints(c *client.Client) {
	fmt.Println("┌───────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Testing Public Endpoints                                      │")
	fmt.Println("└───────────────────────────────────────────────────────────────┘")

	log.Println("[TEST] GET / - Root endpoint")
	resp, err := c.Get("/")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /ping - Ping endpoint")
	resp, err = c.Get("/ping")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /health - Health check")
	resp, err = c.Get("/health")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, IsSuccess: %v\n", resp.StatusCode, resp.IsSuccess())
	}

	log.Println("[TEST] GET /health/live - Liveness probe")
	resp, err = c.Get("/health/live")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /health/ready - Readiness probe")
	resp, err = c.Get("/health/ready")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /external/info - External info")
	resp, err = c.Get("/external/info")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	fmt.Println()
}

func testUsersAPI(c *client.Client) {
	fmt.Println("┌───────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Testing Users API (CRUD)                                      │")
	fmt.Println("└───────────────────────────────────────────────────────────────┘")

	log.Println("[TEST] GET /users - List all users")
	resp, err := c.Get("/users")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /users/1 - Get user by ID")
	resp, err = c.Get("/users/1")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] POST /users - Create new user")
	resp, err = c.Post("/users", map[string]interface{}{
		"name":  "New User",
		"email": "newuser@example.com",
	})
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] PUT /users/1 - Update user")
	resp, err = c.Put("/users/1", map[string]interface{}{
		"name":  "Updated User",
		"email": "updated@example.com",
	})
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] PATCH /users/1 - Partial update user")
	resp, err = c.Patch("/users/1", map[string]interface{}{
		"email": "patched@example.com",
	})
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] DELETE /users/1 - Delete user")
	resp, err = c.Delete("/users/1")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, IsSuccess: %v\n", resp.StatusCode, resp.IsSuccess())
	}

	fmt.Println()
}

func testBasicAuth(baseURL string) {
	fmt.Println("┌───────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Testing Basic Authentication                                  │")
	fmt.Println("└───────────────────────────────────────────────────────────────┘")

	log.Println("[TEST] GET /auth/basic/profile - Valid credentials (admin)")
	c := client.New(baseURL, client.WithBasicAuth("admin", "password123"))
	resp, err := c.Get("/auth/basic/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/basic/profile - Valid credentials (user)")
	c = client.New(baseURL, client.WithBasicAuth("user", "user123"))
	resp, err = c.Get("/auth/basic/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/basic/profile - Invalid credentials")
	c = client.New(baseURL, client.WithBasicAuth("wrong", "wrong"))
	resp, err = c.Get("/auth/basic/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d (expected 401), Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/basic/profile - No credentials")
	c = client.New(baseURL)
	resp, err = c.Get("/auth/basic/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d (expected 401), Body: %s\n", resp.StatusCode, resp.String())
	}

	fmt.Println()
}

func testBearerAuth(baseURL string) {
	fmt.Println("┌───────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Testing Bearer Token Authentication                           │")
	fmt.Println("└───────────────────────────────────────────────────────────────┘")

	log.Println("[TEST] GET /auth/bearer/profile - Valid token")
	c := client.New(baseURL, client.WithBearerToken("secret-token-123"))
	resp, err := c.Get("/auth/bearer/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/bearer/profile - Another valid token")
	c = client.New(baseURL, client.WithBearerToken("another-token"))
	resp, err = c.Get("/auth/bearer/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/bearer/profile - Invalid token")
	c = client.New(baseURL, client.WithBearerToken("invalid-token"))
	resp, err = c.Get("/auth/bearer/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d (expected 401), Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/bearer/profile - No token")
	c = client.New(baseURL)
	resp, err = c.Get("/auth/bearer/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d (expected 401), Body: %s\n", resp.StatusCode, resp.String())
	}

	fmt.Println()
}

func testAPIKeyAuth(baseURL string) {
	fmt.Println("┌───────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Testing API Key Authentication                                │")
	fmt.Println("└───────────────────────────────────────────────────────────────┘")

	log.Println("[TEST] GET /auth/apikey/profile - Valid API key")
	c := client.New(baseURL, client.WithAPIKey("api-key-xyz"))
	resp, err := c.Get("/auth/apikey/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/apikey/profile - Another valid API key")
	c = client.New(baseURL, client.WithAPIKey("api-key-testing"))
	resp, err = c.Get("/auth/apikey/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/apikey/profile - Invalid API key")
	c = client.New(baseURL, client.WithAPIKey("wrong-key"))
	resp, err = c.Get("/auth/apikey/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d (expected 401), Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/apikey/profile - No API key")
	c = client.New(baseURL)
	resp, err = c.Get("/auth/apikey/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d (expected 401), Body: %s\n", resp.StatusCode, resp.String())
	}

	fmt.Println()
}

func testCombinedAuth(baseURL string) {
	fmt.Println("┌───────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Testing Combined Authentication (any method)                  │")
	fmt.Println("└───────────────────────────────────────────────────────────────┘")

	log.Println("[TEST] GET /auth/any/profile - With Basic Auth")
	c := client.New(baseURL, client.WithBasicAuth("admin", "password123"))
	resp, err := c.Get("/auth/any/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/any/profile - With Bearer Token")
	c = client.New(baseURL, client.WithBearerToken("secret-token-123"))
	resp, err = c.Get("/auth/any/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	log.Println("[TEST] GET /auth/any/profile - With API Key")
	c = client.New(baseURL, client.WithAPIKey("api-key-xyz"))
	resp, err = c.Get("/auth/any/profile")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	fmt.Println()
}

func testMetrics(c *client.Client) {
	fmt.Println("┌───────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Testing Metrics Endpoints                                     │")
	fmt.Println("└───────────────────────────────────────────────────────────────┘")

	log.Println("[TEST] GET /metrics - Prometheus format")
	resp, err := c.Get("/metrics")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Length: %d bytes\n", resp.StatusCode, len(resp.Body))
	}

	log.Println("[TEST] GET /metrics/json - JSON format")
	resp, err = c.Get("/metrics/json")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d, Body: %s\n", resp.StatusCode, resp.String())
	}

	fmt.Println()
}

func testErrorCases(c *client.Client, baseURL string) {
	fmt.Println("┌───────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Testing Error Cases                                           │")
	fmt.Println("└───────────────────────────────────────────────────────────────┘")

	log.Println("[TEST] GET /notfound - 404 Not Found")
	resp, err := c.Get("/notfound")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d (expected 404), IsError: %v, Body: %s\n", resp.StatusCode, resp.IsError(), resp.String())
	}

	log.Println("[TEST] POST /users - Invalid JSON body")
	invalidClient := client.New(baseURL, client.WithHeader("Content-Type", "application/json"))
	resp, err = invalidClient.Post("/users", "invalid json")
	if err != nil {
		log.Printf("[FAIL] Error: %v\n", err)
	} else {
		log.Printf("[PASS] Status: %d (expected 400), Body: %s\n", resp.StatusCode, resp.String())
	}

	fmt.Println()
}
