package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func parseLimitOffset(c *fiber.Ctx, defLimit int) (limit, offset int) {
	limit = defLimit
	offset = 0
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}


