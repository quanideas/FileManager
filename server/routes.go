package server

import (
	"filemanager/handlers"
	healthcheck "filemanager/handlers/health-check"
	"filemanager/server/middlewares"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupRoutes(app *fiber.App) {
	allowedDevOrigins := os.Getenv("ALLOWED_DEV_ORIGINS")
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	// Apply CORS
	if os.Getenv("ENVIRONMENT") == "development" {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     allowedDevOrigins,
			AllowCredentials: true,
		}))
	} else {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     allowedOrigins,
			AllowCredentials: true,
		}))
	}

	// Recover from panics
	app.Use(middlewares.CatchPanic())

	// Unauthenticated
	app.Get("/health-check", healthcheck.HealthCheck)
	app.Get("/connection-check", healthcheck.ConnectionCheck)

	// JWT Middleware
	app.Use(middlewares.ValidateJWT())

	// Authenticated
	app.Get("/project/:companyID/:projectID/:iterationID/*", handlers.GetProjectFile)
	app.Post("/project/upload-iteration", handlers.CreateProjectIteration)
	app.Post("/project/edit-iteration", handlers.UpdateProjectIteration)
	app.Post("/project/remove-iteration", handlers.DeleteProjectIteration)
}
