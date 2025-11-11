package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/artem13815/hr/api/http/handlers"
)

// Register wires all HTTP routes onto given Fiber app.
func Register(app *fiber.App, auth *handlers.AuthHandler, health *handlers.HealthHandler, resume *handlers.ResumeHandler) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Health and readiness endpoints for probes/monitoring
	v1.Get("/health", health.Health)
	v1.Get("/ready", health.Ready)

	a := v1.Group("/auth")
	a.Post("/register", auth.Register)
	a.Post("/login", auth.Login)

	// Resume analysis
	rg := v1.Group("/resume")
	rg.Post("/analyze", resume.Analyze)
}
