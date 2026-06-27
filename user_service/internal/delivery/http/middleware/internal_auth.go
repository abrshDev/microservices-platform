// internal/middleware/internal_auth.go
package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

// InternalAuth protects routes that should only be called by other services.
// It reads the expected key from the INTERNAL_API_KEY environment variable.
func InternalAuth() fiber.Handler {
	expectedKey := os.Getenv("INTERNAL_API_KEY")

	return func(c *fiber.Ctx) error {
		key := c.Get("X-INTERNAL-API-KEY") // read the key from the request header

		if key == "" || key != expectedKey {
			return c.Status(401).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		return c.Next() // key is valid — let the request through to the handler
	}
}
