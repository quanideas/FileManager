package handlers

import (
	"filemanager/common/constants"
	"filemanager/common/helpers"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/google/uuid"
)

func GetProjectFile(c *fiber.Ctx) error {
	// Parse url
	companyID := c.Params("companyID")
	projectIDString := c.Params("projectID")
	iterationID := c.Params("iterationID")
	file, _ := url.PathUnescape(c.Params("*"))
	token := c.Cookies("token")
	refreshToken := c.Cookies("refreshToken")

	// Parse ProjectID
	projectID, err := uuid.Parse(projectIDString)
	if err != nil {
		helpers.BadRequest(c, "invalid project id", constants.ERR_PROJECT_NOT_FOUND)
		return nil
	}

	// Validate permission
	if errCode, err := callValidatePermission(projectID, token, refreshToken); err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Path to file
	fileSystemRoot := helpers.GetFileSystemRootLocation()
	fileLocation := fmt.Sprintf("%s/%s/%s/%s",
		companyID,
		projectID,
		iterationID,
		file)

	// Serve the file using SendFile function
	err = filesystem.SendFile(c, http.Dir(fileSystemRoot), fileLocation)
	if err != nil {
		// Handle the error, e.g., return a 404 Not Found response
		return c.Status(fiber.StatusNotFound).SendString("File not found")
	}

	return nil
}
