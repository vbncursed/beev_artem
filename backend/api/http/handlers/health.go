package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/artem13815/hr/pkg/health"
)

// HealthHandler serves liveness and readiness probes.
type HealthHandler struct{ svc health.ReadinessUseCase }

func NewHealthHandler(svc health.ReadinessUseCase) *HealthHandler { return &HealthHandler{svc: svc} }

// Health: basic liveness check.
// @Summary Liveness probe
// @Tags    health
// @Produce json
// @Success 200 {object} map[string]string
// @Router  /health [get]
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

// Ready: readiness check with DB ping.
// @Summary Readiness probe
// @Tags    health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router  /ready [get]
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 1*time.Second)
	defer cancel()
	if err := h.svc.Ready(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "not_ready",
			"details": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ready"})
}
