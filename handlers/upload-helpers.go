package handlers

import (
	"archive/zip"
	"bytes"
	"errors"
	"filemanager/common/constants"
	"filemanager/common/helpers"
	"filemanager/models/request"
	"filemanager/models/response"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func callGetProjectIteration(iterationID string, token, refreshToken string) (response.IterationResponse, int, error) {
	// Get info of Project microservice
	host := os.Getenv("PROJECT_SERVICE_HOST")
	port := os.Getenv("PROJECT_SERVICE_PORT")
	api := constants.ProjectIterationGet
	url := fmt.Sprintf("%s:%s/project/%s", host, port, api)

	// Call to delete an iteration
	agent := fiber.Post(url)
	agent.JSON(request.GetByIDRequest{
		ID: iterationID,
	})

	var data response.IterationResponse
	errCode, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken)
	return data, errCode, err
}

func callCreateProjectIteration(projectID uuid.UUID, revision, token, refreshToken string) (response.IterationResponse, int, error) {
	// Get info of Project microservice
	host := os.Getenv("PROJECT_SERVICE_HOST")
	port := os.Getenv("PROJECT_SERVICE_PORT")
	api := constants.ProjectIterationCreate
	url := fmt.Sprintf("%s:%s/project/%s", host, port, api)

	// Call to create an iteration
	agent := fiber.Post(url)
	agent.JSON(request.CreateIterationRequest{
		ProjectID: uuid.UUID(projectID),
		Revision:  &revision,
	})

	var data response.IterationResponse
	errCode, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken)
	return data, errCode, err
}

func callUpdateProjectIteration(request request.UpdateIterationRequest, token, refreshToken string) (response.IterationResponse, int, error) {
	// Get info of Project microservice
	host := os.Getenv("PROJECT_SERVICE_HOST")
	port := os.Getenv("PROJECT_SERVICE_PORT")
	api := constants.ProjectIterationUpdate
	url := fmt.Sprintf("%s:%s/project/%s", host, port, api)

	// Call to update an iteration
	agent := fiber.Post(url)
	agent.JSON(request)

	var data response.IterationResponse
	errCode, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken)
	return data, errCode, err
}

func callDeleteProjectIteration(iterationID uuid.UUID, token, refreshToken string) (int, error) {
	// Get info of Project microservice
	host := os.Getenv("PROJECT_SERVICE_HOST")
	port := os.Getenv("PROJECT_SERVICE_PORT")
	api := constants.ProjectIterationDelete
	url := fmt.Sprintf("%s:%s/project/%s", host, port, api)

	// Call to delete an iteration
	agent := fiber.Post(url)
	agent.JSON(request.DeleteByIDRequest{
		ID: iterationID,
	})

	var data string
	errCode, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken)
	return errCode, err
}

func callGetCompanyIDFromProjectID(projectID uuid.UUID, token, refreshToken string) (string, int, error) {
	// Get info of Project microservice
	host := os.Getenv("PROJECT_SERVICE_HOST")
	port := os.Getenv("PROJECT_SERVICE_PORT")
	api := constants.ProjectGetCompanyIDByProjectID
	url := fmt.Sprintf("%s:%s/project/%s", host, port, api)

	// Call to delete an iteration
	agent := fiber.Post(url)
	agent.JSON(request.GetByIDRequest{
		ID: projectID.String(),
	})

	var data string
	errCode, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken)
	return data, errCode, err
}

func companySavedFileSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func saveAndUnzipFile(saveDirectory string, file *multipart.FileHeader, errChannel chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	if file == nil {
		return
	}

	// Create new unzipper from mime multipart file
	fileOpened, err := file.Open()
	if err != nil {
		errChannel <- err
		return
	}
	unzipper, err := zip.NewReader(fileOpened, file.Size)
	if err != nil {
		errChannel <- err
		return
	}

	// Unzip files inside of the zipped file
	for _, f := range unzipper.File {
		err := unzipFile(f, saveDirectory)
		if err != nil {
			errChannel <- err
			return
		}
	}

}

func unzipFile(f *zip.File, destination string) error {
	// 4. Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// 5. Create directory tree
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// 6. Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// 7. Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}

func allowFileTypeCheck(files []*multipart.FileHeader) error {
	allowedExtensionsList := []string{".zip", ".rar", ".7z"}
	var notAllowed []string

	for _, file := range files {
		if file != nil {
			extension := filepath.Ext(file.Filename)

			if !slices.Contains(allowedExtensionsList, extension) &&
				!slices.Contains(notAllowed, extension) {

				notAllowed = append(notAllowed, extension)
			}
		}
	}

	if len(notAllowed) != 0 {
		var buffer bytes.Buffer

		buffer.WriteString("File extensions: ")

		for i := 0; i < len(notAllowed); i++ {
			if i == len(notAllowed)-1 {
				buffer.WriteString(fmt.Sprintf("%s ", notAllowed[i]))
			} else {
				buffer.WriteString(fmt.Sprintf("%s, ", notAllowed[i]))
			}
		}

		buffer.WriteString("not allowed.")

		return errors.New(buffer.String())
	}

	return nil
}
