package handlers

import (
	"errors"
	"filemanager/common/constants"
	"filemanager/common/helpers"
	"filemanager/models/request"
	"filemanager/models/response"
	"fmt"
	"mime/multipart"
	"os"
	"sync"

	"github.com/devfeel/mapper"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CreateProjectIteration calls project microservice to create an iteration on db
// then saves all its files to the file system root location.
// Params
// project_id: ID of the project this iteration belongs to
// geojson: geojson files as zip
// tile_3d: 3DTile files as zip
// ortho_photo: ortho photo files as zip
func CreateProjectIteration(c *fiber.Ctx) error {
	var updatedProjectIteration response.IterationResponse

	// Get info from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	isRoot := claims["is_root"].(bool)
	token := c.Cookies("token")
	refreshToken := c.Cookies("refreshToken")

	// Only allow root to create
	if !isRoot {
		helpers.BadRequest(c, "field not found", constants.ERR_PROJECT_ITERATION_UPLOAD_NOT_ALLOWED)
		return nil
	}

	// => *multipart.Form
	form, err := c.MultipartForm()
	if err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get and validate projectID
	projectIDString := form.Value["project_id"][0]
	projectID, err := uuid.Parse(projectIDString)
	if err != nil {
		helpers.BadRequest(c, "invalid project id", constants.ERR_PROJECT_NOT_FOUND)
		return nil
	}

	// Get company ID from project's ID
	companyID, errCode, err := callGetCompanyIDFromProjectID(projectID, token, refreshToken)
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Get files
	var geoJSONFile, tile3DFile, orthoPhotoFile *multipart.FileHeader
	if len(form.File["geojson"]) != 0 {
		geoJSONFile = form.File["geojson"][0]
	}
	if len(form.File["tile_3d"]) != 0 {
		tile3DFile = form.File["tile_3d"][0]
	}
	if len(form.File["ortho_photo"]) != 0 {
		orthoPhotoFile = form.File["ortho_photo"][0]
	}

	// Check for allowed file types
	if fileCheckErr := allowFileTypeCheck([]*multipart.FileHeader{geoJSONFile, tile3DFile, orthoPhotoFile}); fileCheckErr != nil {
		helpers.BadRequest(c, fileCheckErr.Error(), constants.ERR_FILE_TYPE_NOT_ALLOWED)
		return nil
	}

	// Call project service to create a project iteration first
	revision := "" // Get revision
	if len(form.Value["revision"]) > 0 {
		revision = form.Value["revision"][0]
	}
	projectIteration, errCode, err := callCreateProjectIteration(projectID, revision, token, refreshToken)
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Get file save location
	fileSystemRoot := helpers.GetFileSystemRootLocation()
	saveDirectory := fmt.Sprintf("%s/%s/%s/%s", fileSystemRoot, companyID, projectID, projectIteration.ID.String())
	geoJSONDirectory := fmt.Sprintf("%s/%s", saveDirectory, "geojson")
	tile3DDirectory := fmt.Sprintf("%s/%s", saveDirectory, "tile_3d")
	orthoPhotoDirectory := fmt.Sprintf("%s/%s", saveDirectory, "ortho_photo")

	// Create save locations
	if _, err := os.Stat(geoJSONDirectory); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(geoJSONDirectory, os.ModePerm); err != nil {
			helpers.InternalServerError(c, err.Error())
			return nil
		}
		if err := os.MkdirAll(tile3DDirectory, os.ModePerm); err != nil {
			helpers.InternalServerError(c, err.Error())
			return nil
		}
		if err := os.MkdirAll(orthoPhotoDirectory, os.ModePerm); err != nil {
			helpers.InternalServerError(c, err.Error())
			return nil
		}
	}

	// Spawn go processes to save files cocurrenly
	var wg sync.WaitGroup
	wg.Add(3)
	errChannel := make(chan error)

	go saveAndUnzipFile(geoJSONDirectory, geoJSONFile, errChannel, &wg)
	go saveAndUnzipFile(tile3DDirectory, tile3DFile, errChannel, &wg)
	go saveAndUnzipFile(orthoPhotoDirectory, orthoPhotoFile, errChannel, &wg)

	// here we wait in other goroutine to all jobs done and close the channels
	go func() {
		wg.Wait()
		close(errChannel)
	}()

	// Check for error, then revert by deleting project iteration
	for err := range errChannel {
		// Delete files
		os.RemoveAll(saveDirectory)

		// Delete project iteration db record
		errCode, deleteErr := callDeleteProjectIteration(projectIteration.ID, token, refreshToken)
		if deleteErr != nil {
			helpers.InternalServerError(c, deleteErr.Error(), errCode)
			return nil
		}

		// Return save error
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// No error, update project iteration on db with project's url and file names
	updateIterationRequest := request.UpdateIterationRequest{
		ID:       projectIteration.ID,
		Revision: &revision,
	}
	baseURL := fmt.Sprintf("/%s/%s/%s", companyID, projectID, updateIterationRequest.ID.String())
	if geoJSONFile != nil {
		geoJSONURL := fmt.Sprintf("%s/%s", baseURL, "geojson")
		updateIterationRequest.GeoJSONURL = &geoJSONURL
		updateIterationRequest.GeoJSONFileName = &geoJSONFile.Filename
	}
	if tile3DFile != nil {
		tile3DURL := fmt.Sprintf("%s/%s", baseURL, "tile_3d")
		updateIterationRequest.Tile3DURL = &tile3DURL
		updateIterationRequest.Tile3DFileName = &tile3DFile.Filename
	}
	if orthoPhotoFile != nil {
		orthoPhotoURL := fmt.Sprintf("%s/%s", baseURL, "ortho_photo")
		updateIterationRequest.OrthoPhotoURL = &orthoPhotoURL
		updateIterationRequest.OrthoPhotoFileName = &orthoPhotoFile.Filename
	}
	updatedProjectIteration, errCode, err = callUpdateProjectIteration(updateIterationRequest, token, refreshToken)

	if err != nil {
		// Delete files
		os.RemoveAll(saveDirectory)

		// Delete project iteration db record
		deleteErrCode, deleteErr := callDeleteProjectIteration(projectIteration.ID, token, refreshToken)
		if deleteErr != nil {
			helpers.InternalServerError(c, deleteErr.Error(), deleteErrCode)
			return nil
		}

		helpers.InternalServerError(c, err.Error(), errCode)
		return nil
	}

	// Return created iteration
	c.Status(201)
	c.JSON(response.BaseResponse{
		Data: updatedProjectIteration,
		Meta: struct{ Status int }{Status: 200},
	})
	return nil
}

// UpdateProjectIteration calls project microservice to update an iteration on db
// then saves thee new files/keep the old files if remmove is not true. Else deletes
// the old files and update url to null
// Params
// iteration_id: ID of the iteration to be updated
// geojson: geojson files as zip, to be upploaded if remove != true
// removeGeoJson: true to delete old files, false to upload new or keep old files
// tile_3d: 3DTile files as zip, to be upploaded if remove != true
// removeTile3D: true to delete old files, false to upload new or keep old files
// ortho_photo: ortho photo files as zip, to be upploaded if remove != true
// removeOrthoPhoto: true to delete old files, false to upload new or keep old files
func UpdateProjectIteration(c *fiber.Ctx) error {
	var updatedProjectIteration response.IterationResponse

	// Get info from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	isRoot := claims["is_root"].(bool)
	token := c.Cookies("token")
	refreshToken := c.Cookies("refreshToken")

	// Only allow root to create
	if !isRoot {
		helpers.BadRequest(c, "no permission to upload", constants.ERR_PROJECT_ITERATION_UPLOAD_NOT_ALLOWED)
		return nil
	}

	// => *multipart.Form
	form, err := c.MultipartForm()
	if err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get iteration from Project microservice
	iterationID := form.Value["id"][0]
	projectIteration, errCode, err := callGetProjectIteration(iterationID, token, refreshToken)
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}
	var toBeUpdatedProjectIteration request.UpdateIterationRequest
	toBeUpdatedProjectIteration.ID = projectIteration.ID
	mapper.Mapper(&projectIteration, &toBeUpdatedProjectIteration)

	// Get company ID from project's ID
	companyID, errCode, err := callGetCompanyIDFromProjectID(projectIteration.ProjectID, token, refreshToken)
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Get if user wants to remove files, true = remove, false = keep/upload new file
	isRemoveGeoJSON := form.Value["removeGeoJson"][0]
	isRemoveTile3D := form.Value["removeTile3D"][0]
	isRemoveOrthoPhoto := form.Value["removeOrthoPhoto"][0]

	// Get files
	var geoJSONFile, tile3DFile, orthoPhotoFile *multipart.FileHeader
	if len(form.File["geojson"]) != 0 {
		geoJSONFile = form.File["geojson"][0]
	}
	if len(form.File["tile_3d"]) != 0 {
		tile3DFile = form.File["tile_3d"][0]
	}
	if len(form.File["ortho_photo"]) != 0 {
		orthoPhotoFile = form.File["ortho_photo"][0]
	}

	// Check for allowed file types
	if fileCheckErr := allowFileTypeCheck([]*multipart.FileHeader{geoJSONFile, tile3DFile, orthoPhotoFile}); fileCheckErr != nil {
		helpers.BadRequest(c, fileCheckErr.Error(), constants.ERR_FILE_TYPE_NOT_ALLOWED)
		return nil
	}

	// Get temporary file save location
	fileSystemRoot := helpers.GetFileSystemRootLocation()
	saveDirectory := fmt.Sprintf("%s/%s/%s/%s", fileSystemRoot, companyID, projectIteration.ProjectID, projectIteration.ID.String())
	geoJSONTempDirectory := fmt.Sprintf("%s/%s", saveDirectory, "geojson_temp")
	geoJSONDirectory := fmt.Sprintf("%s/%s", saveDirectory, "geojson")
	tile3DTempDirectory := fmt.Sprintf("%s/%s", saveDirectory, "tile_3d_temp")
	tile3DDirectory := fmt.Sprintf("%s/%s", saveDirectory, "tile_3d")
	orthoPhotoTempDirectory := fmt.Sprintf("%s/%s", saveDirectory, "ortho_photo_temp")
	orthoPhotoDirectory := fmt.Sprintf("%s/%s", saveDirectory, "ortho_photo")

	// Create save locations
	if err := os.MkdirAll(geoJSONTempDirectory, os.ModePerm); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}
	if err := os.MkdirAll(tile3DTempDirectory, os.ModePerm); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}
	if err := os.MkdirAll(orthoPhotoTempDirectory, os.ModePerm); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Spawn go processes to save files cocurrenly
	var wg sync.WaitGroup
	errChannel := make(chan error)
	var saveFileErr error

	// Only save and set URL if isRemove is not true
	baseURL := fmt.Sprintf("/%s/%s/%s", companyID, projectIteration.ProjectID, projectIteration.ID.String())
	if isRemoveGeoJSON != "true" {
		if geoJSONFile != nil {
			wg.Add(1)
			go saveAndUnzipFile(geoJSONTempDirectory, geoJSONFile, errChannel, &wg)

			geoJSONURL := fmt.Sprintf("%s/%s", baseURL, "geojson")
			toBeUpdatedProjectIteration.GeoJSONURL = &geoJSONURL
			toBeUpdatedProjectIteration.GeoJSONFileName = &geoJSONFile.Filename
		}
	} else { // Set URL and file name to null
		toBeUpdatedProjectIteration.GeoJSONURL = nil
		toBeUpdatedProjectIteration.GeoJSONFileName = nil
	}
	// Same for 3DTile
	if isRemoveTile3D != "true" {
		if tile3DFile != nil {
			wg.Add(1)
			go saveAndUnzipFile(tile3DTempDirectory, tile3DFile, errChannel, &wg)

			tile3DURL := fmt.Sprintf("%s/%s", baseURL, "tile_3d")
			toBeUpdatedProjectIteration.Tile3DURL = &tile3DURL
			toBeUpdatedProjectIteration.Tile3DFileName = &tile3DFile.Filename
		}
	} else { // Set URL and file name to null
		toBeUpdatedProjectIteration.Tile3DURL = nil
		toBeUpdatedProjectIteration.Tile3DFileName = nil
	}
	// Same for Ortho Photo
	if isRemoveOrthoPhoto != "true" {
		if orthoPhotoFile != nil {
			wg.Add(1)
			go saveAndUnzipFile(orthoPhotoTempDirectory, orthoPhotoFile, errChannel, &wg)

			orthoPhotoURL := fmt.Sprintf("%s/%s", baseURL, "ortho_photo")
			toBeUpdatedProjectIteration.OrthoPhotoURL = &orthoPhotoURL
			toBeUpdatedProjectIteration.OrthoPhotoFileName = &orthoPhotoFile.Filename
		}
	} else { // Set URL and file name to null
		toBeUpdatedProjectIteration.OrthoPhotoURL = nil
		toBeUpdatedProjectIteration.OrthoPhotoFileName = nil
	}

	// Get revision to update
	if len(form.Value["revision"]) > 0 {
		toBeUpdatedProjectIteration.Revision = &form.Value["revision"][0]
	}

	// Update record on db
	updatedProjectIteration, updateErrCode, updateErr := callUpdateProjectIteration(toBeUpdatedProjectIteration, token, refreshToken)

	// Wait for all files to save
	go func() {
		wg.Wait()
		close(errChannel)
	}()
	for err := range errChannel {
		if err != nil {
			saveFileErr = err
		}
	}

	// Failed, revert by deleting the files and update back the old Iteration db record
	if saveFileErr != nil || updateErr != nil {
		// Delete temporary uploaded directories
		os.RemoveAll(geoJSONTempDirectory)
		os.RemoveAll(tile3DTempDirectory)
		os.RemoveAll(orthoPhotoTempDirectory)

		// Update to the old record
		mapper.Mapper(&projectIteration, &toBeUpdatedProjectIteration)
		if _, updateErrCode, updateErr := callUpdateProjectIteration(toBeUpdatedProjectIteration, token, refreshToken); updateErr != nil {
			helpers.InternalServerError(c, updateErr.Error(), updateErrCode)
			return nil
		}

		// Return save error
		if updateErr != nil {
			helpers.InternalServerError(c, updateErr.Error(), updateErrCode)
		} else {
			helpers.InternalServerError(c, saveFileErr.Error())
		}
		return nil
	}

	// Success, if remove is true then delete all files in the folder and remake the directory.
	// Else if remove is false, and uploaded a new file, then remove the old files and rename the
	// temporary folder to the original folder name
	if isRemoveGeoJSON == "true" {
		os.RemoveAll(geoJSONDirectory)
		os.RemoveAll(geoJSONTempDirectory)
		os.MkdirAll(geoJSONDirectory, os.ModePerm)
	} else if geoJSONFile != nil {
		os.RemoveAll(geoJSONDirectory)
		os.Rename(geoJSONTempDirectory, geoJSONDirectory)
	}
	if isRemoveTile3D == "true" {
		os.RemoveAll(tile3DDirectory)
		os.RemoveAll(tile3DTempDirectory)
		os.MkdirAll(tile3DDirectory, os.ModePerm)
	} else if tile3DFile != nil {
		os.RemoveAll(tile3DDirectory)
		os.Rename(tile3DTempDirectory, tile3DDirectory)
	}
	if isRemoveOrthoPhoto == "true" {
		os.RemoveAll(orthoPhotoDirectory)
		os.RemoveAll(orthoPhotoTempDirectory)
		os.MkdirAll(orthoPhotoDirectory, os.ModePerm)
	} else if orthoPhotoFile != nil {
		os.RemoveAll(orthoPhotoDirectory)
		os.Rename(orthoPhotoTempDirectory, orthoPhotoDirectory)
	}

	// Return created iteration
	c.Status(201)
	c.JSON(response.BaseResponse{
		Data: updatedProjectIteration,
		Meta: struct{ Status int }{Status: 200},
	})
	return nil
}

// Delete calls project microservice to delete an iteration from db
// and all its files saved here.
// Params
// id: ID of the project iteration
func DeleteProjectIteration(c *fiber.Ctx) error {
	// Parse request model
	request := request.DeleteByIDRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get info from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	isRoot := claims["is_root"].(bool)
	token := c.Cookies("token")
	refreshToken := c.Cookies("refreshToken")

	// Only allow root to delete
	if !isRoot {
		helpers.BadRequest(c, "no permission to delete", constants.ERR_PROJECT_ITERATION_DELETE_NOT_ALLOWED)
		return nil
	}

	// Get iteration from Project microservice
	projectIteration, errCode, err := callGetProjectIteration(request.ID.String(), token, refreshToken)
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Get company ID from project's ID
	companyID, errCode, err := callGetCompanyIDFromProjectID(projectIteration.ProjectID, token, refreshToken)
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Delete project iteration db record
	errCode, err = callDeleteProjectIteration(request.ID, token, refreshToken)
	if err != nil {
		helpers.InternalServerError(c, err.Error(), errCode)
		return nil
	}

	// Get file save location then delete
	fileSystemRoot := helpers.GetFileSystemRootLocation()
	saveDirectory := fmt.Sprintf("%s/%s/%s/%s", fileSystemRoot, companyID, projectIteration.ProjectID, request.ID.String())
	os.RemoveAll(saveDirectory)

	// Return created iteration
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: "Success",
		Meta: struct{ Status int }{Status: 200},
	})
	return nil
}
