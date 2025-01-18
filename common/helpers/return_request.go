package helpers

import (
	"filemanager/models/response"

	"github.com/gofiber/fiber/v2"
)

func BadRequest(c *fiber.Ctx, err string, code_optional ...int) {
	errorCode := 400
	if len(code_optional) > 0 {
		errorCode = code_optional[0]
	}

	c.Status(fiber.StatusBadRequest)
	c.JSON(response.ErrorResponse{
		ErrorCode: errorCode,
		Error:     err,
	})
}

func InternalServerError(c *fiber.Ctx, err string, code_optional ...int) {
	errorCode := 500
	if len(code_optional) > 0 {
		errorCode = code_optional[0]
	}

	c.Status(fiber.StatusInternalServerError)
	c.JSON(response.ErrorResponse{
		ErrorCode: errorCode,
		Error:     err,
	})
}
