// @title         hr-service API
// @version       1.0
// @description   Сервис автоматизирующий процесс оценки соответствия кандидатов требованиям вакансии на основе анализа их резюме с применением NLP-технологий и LLM-модели.
// @BasePath      /api/v1
// @schemes       http
// @host          localhost:8080
package main

import (
	"context"
	"log"
	"time"

	_ "github.com/artem13815/hr/docs"
	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"

	// internal imports
	"github.com/artem13815/hr/api/http"
	"github.com/artem13815/hr/api/http/handlers"
	"github.com/artem13815/hr/pkg/auth"
	"github.com/artem13815/hr/pkg/config"
	"github.com/artem13815/hr/pkg/health"
	healthpg "github.com/artem13815/hr/pkg/health/checkers"
	"github.com/artem13815/hr/pkg/llm/openrouter"
	pgrepo "github.com/artem13815/hr/pkg/repository/postgres"
	"github.com/artem13815/hr/pkg/resume"
	"github.com/artem13815/hr/pkg/security/jwt"
	"github.com/artem13815/hr/pkg/storage/postgres"
)

func main() {
	app := fiber.New()

	// Load configuration from env/.env
	cfg := config.Load()

	// Connect to PostgreSQL
	dsn := cfg.DatabaseURL
	if dsn == "" {
		log.Fatal("DATABASE_URL не задан: например, postgres://user:pass@localhost:5432/db?sslmode=disable")
	}
	pool, err := postgres.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("postgres connect: %v", err)
	}
	defer pool.Close()

	// Wire dependencies (Clean Architecture)
	userRepo, err := pgrepo.NewUserRepository(pool)
	if err != nil {
		log.Fatalf("init user repo: %v", err)
	}
	// Token generator
	jwtGen := jwt.NewGenerator(cfg.JWTSecret, cfg.JWTIssuer, time.Duration(cfg.JWTTTLMinutes)*time.Minute)

	authUC := auth.NewAuthService(userRepo, jwtGen)
	authHandler := handlers.NewAuthHandler(authUC)

	// Health service: compose checkers
	readiness := health.NewService(healthpg.NewPostgresChecker(pool))
	healthHandler := handlers.NewHealthHandler(readiness)

	// OpenRouter client and resume handler
	llmClient := openrouter.New(
		cfg.OpenRouterAPIKey,
		cfg.OpenRouterBase,
		cfg.OpenRouterModel,
		cfg.OpenRouterAppTitle,
		cfg.OpenRouterReferer,
	)
	resumeSvc := resume.NewAnalysisService(llmClient)
	resumeHandler := handlers.NewResumeHandler(resumeSvc, llmClient)

	// Register routes
	http.Register(app, authHandler, healthHandler, resumeHandler)

	// Swagger UI
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Start server
	port := cfg.Port
	log.Printf("HTTP server listening on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
