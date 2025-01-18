package server

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func RunServer() {
	env := os.Getenv("ENVIRONMENT")

	var app *fiber.App

	if env == "development" {
		app = fiber.New(fiber.Config{
			BodyLimit:               1024 * 1024 * 1024, // 1024 MB = 1 GB
			EnableTrustedProxyCheck: false,
		})
	} else {
		requestLimit, _ := strconv.Atoi(os.Getenv("REQUEST_LIMIT"))
		app = fiber.New(fiber.Config{
			BodyLimit:               requestLimit * 1024 * 1024, // requestLimit * 1 MB
			EnableTrustedProxyCheck: true,
		})
	}

	SetupRoutes(app)

	port := fmt.Sprintf(":%v", os.Getenv("SERVER_IN_PORT"))
	app.Listen(port)
}
