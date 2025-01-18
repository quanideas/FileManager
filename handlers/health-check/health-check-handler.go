package healthcheck

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

func HealthCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func ConnectionCheck(c *fiber.Ctx) error {
	healthCheck := struct {
		UserService    bool `json:"user_service"`
		ProjectService bool `json:"project_service"`
		TokenService   bool `json:"token_service"`
	}{}

	// Check User service
	host := os.Getenv("USER_SERVICE_HOST")
	port := os.Getenv("USER_SERVICE_PORT")
	api := "/health-check"
	url := fmt.Sprintf("%s:%s/%s", host, port, api)

	agent := fiber.Get(url)
	userStatusCode, _, _ := agent.Bytes()
	healthCheck.UserService = userStatusCode == fiber.StatusOK

	// Check Project service
	host = os.Getenv("PROJECT_SERVICE_HOST")
	port = os.Getenv("PROJECT_SERVICE_PORT")
	api = "/health-check"
	url = fmt.Sprintf("%s:%s/%s", host, port, api)

	agent = fiber.Get(url)
	projectStatusCode, _, _ := agent.Bytes()
	healthCheck.ProjectService = projectStatusCode == fiber.StatusOK

	// Check Token service
	host = os.Getenv("TOKEN_SERVICE_HOST")
	port = os.Getenv("TOKEN_SERVICE_PORT")
	api = "/health-check"
	url = fmt.Sprintf("%s:%s/%s", host, port, api)

	agent = fiber.Get(url)
	tokenStatusCode, _, _ := agent.Bytes()
	healthCheck.TokenService = tokenStatusCode == fiber.StatusOK

	c.JSON(healthCheck)
	return nil
}
