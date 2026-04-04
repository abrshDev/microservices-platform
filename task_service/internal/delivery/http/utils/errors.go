package utils

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

func MapError(c *fiber.Ctx, err error) error {
	// If it's a validation error (User not found)
	if errors.Is(err, errors.New("user not found")) { // You'll want to use a custom error type here later
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":  "User Validation Failed",
			"detail": "The provided user_id does not exist in our system.",
		})
	}

	// Default: Internal Server Error
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":  "Internal Service Error",
		"detail": "An unexpected error occurred. Please try again later.",
	})
}
