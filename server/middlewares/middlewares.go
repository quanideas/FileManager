package middlewares

import (
	"filemanager/common/constants"
	"filemanager/common/helpers"
	"filemanager/models/request"
	"filemanager/models/response"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
)

func CatchPanic() fiber.Handler {
	// Return new handler
	return func(c *fiber.Ctx) (err error) { //nolint:nonamedreturns // Uses recover() to overwrite the error
		// Catch panics
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("Panic: %v", r))
				helpers.InternalServerError(c, fmt.Sprintf("%v", r))
			}
		}()

		// Return err if exist, else move to next handler
		return c.Next()
	}
}

func ValidateJWT() fiber.Handler {
	// Return new handler
	return func(c *fiber.Ctx) (err error) { //nolint:nonamedreturns // Uses recover() to overwrite the error
		// Get token and refresh token from cookies
		token := c.Cookies("token")
		refreshToken := c.Cookies("refreshToken")

		// No token or refresh token, return unauthorized
		if token == "" || refreshToken == "" {
			// Clear cookies
			c.ClearCookie("token", "refreshToken")

			c.Status(fiber.StatusUnauthorized)
			c.JSON(response.ErrorResponse{
				ErrorCode: 401,
				Error:     "unauthorized",
			})
			return nil
		}

		// Get info of Token Validation microservice
		host := os.Getenv("TOKEN_SERVICE_HOST")
		port := os.Getenv("TOKEN_SERVICE_PORT")
		api := constants.TokenValidation
		url := fmt.Sprintf("%s:%s/%s", host, port, api)

		// Call token validation service to validate token
		agent := fiber.Post(url)
		agent.JSON(request.ValidationRequest{
			Token:        token,
			RefreshToken: refreshToken,
		})
		var data response.ValidationResponse
		errCode, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken)

		// If error then return bad request, or token is invalid then return unauthorize
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			c.JSON(response.ErrorResponse{
				ErrorCode: errCode,
				Error:     err.Error(),
			})
			return nil
		} else if !data.IsValid {
			// Clear cookies
			c.ClearCookie("token", "refreshToken")

			c.Status(fiber.StatusUnauthorized)
			c.JSON(response.ErrorResponse{
				ErrorCode: 401,
				Error:     "unauthorized",
			})
			return nil
		}

		// Set new Token to cookies
		c.Cookie(&fiber.Cookie{
			Name:  "token",
			Value: data.Token,
		})

		// Set token in local scope (this request's scope) to be able to parse when needed
		var localToken *jwt.Token
		localToken, _ = jwt.Parse(data.Token, nil)
		c.Locals("user", localToken)

		// Move to next handler
		return c.Next()
	}
}
