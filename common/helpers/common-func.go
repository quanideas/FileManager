package helpers

import (
	"encoding/json"
	"errors"
	"filemanager/models/response"
	"os"
	"path/filepath"
	"reflect"

	"github.com/gofiber/fiber/v2"
)

func GetFileSystemRootLocation() string {
	fileSystemRoot := os.Getenv("UPLOAD_DIRECTORY")
	if len(fileSystemRoot) == 0 {
		fileSystemRoot, _ = os.Executable()
		fileSystemRoot = filepath.Dir(fileSystemRoot)
		fileSystemRoot += "/uploads"
	}

	return fileSystemRoot
}

func SendAndParseResponseData(agent *fiber.Agent, object any, token, refreshToken string) (int, error) {
	// Validate object to be a pointer
	reflectObject := reflect.ValueOf(object)
	if reflectObject.Kind() != reflect.Pointer || reflectObject.IsNil() {
		return 500, errors.New("object is not a pointer")
	}

	// Set tokens
	agent.Cookie("token", token)
	agent.Cookie("refreshToken", refreshToken)

	// Send request
	statusCode, body, errs := agent.Bytes()

	// Handle failed request
	if len(errs) > 0 || statusCode < 200 || statusCode >= 300 {
		var errMsg response.ErrorResponse
		json.Unmarshal(body, &errMsg)

		return errMsg.ErrorCode, errors.New(errMsg.Error)
	}

	// Success, decode json to response DTO
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return 500, err
	}

	// Encode Data to JSON bytes, then decode to get data as object
	data := response["Data"]
	if dataEncoded, err := json.Marshal(data); err != nil {
		return 500, err
	} else if err = json.Unmarshal(dataEncoded, object); err != nil {
		return 500, err
	}

	return 0, nil
}
