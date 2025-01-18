package handlers

import (
	"errors"
	"filemanager/common/constants"
	"filemanager/common/helpers"
	"filemanager/models/request"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func callValidatePermission(projectID uuid.UUID, token, refreshToken string) (int, error) {
	// Get info of User microservice
	host := os.Getenv("USER_SERVICE_HOST")
	port := os.Getenv("USER_SERVICE_PORT")
	api := constants.PermissionValidate
	url := fmt.Sprintf("%s:%s/permission%s", host, port, api)

	// Check if user has permission view to this project
	agent := fiber.Post(url)
	agent.JSON(request.GetUserSpecificPermissionRequest{
		ProjectID:       &projectID,
		PermissionType:  constants.PERM_PROJECT,
		PermissionLevel: constants.PERM_LEVEL_VIEW,
	})

	// Get permission
	var data string
	if errCode, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken); err != nil {
		return errCode, err
	}

	// Permission denied then return denied
	if data != "Granted" {
		return constants.ERR_COMMON_PERMISSION_NOT_ALLOWED, errors.New("no permission")
	}

	return 0, nil
}
