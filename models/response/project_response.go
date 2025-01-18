package response

import (
	"filemanager/models/entity"

	"github.com/google/uuid"
)

type IterationResponse struct {
	entity.BaseEntityModel
	ProjectID          uuid.UUID `json:"project_id"`
	Revision           *string   `json:"revision"`
	GeoJSONURL         *string   `json:"geojson_url"`
	GeoJSONFileName    *string   `json:"geojson_file_name"`
	OrthoPhotoURL      *string   `json:"ortho_photo_url"`
	OrthoPhotoFileName *string   `json:"ortho_photo_file_name"`
	Tile3DURL          *string   `json:"tile_3d_url"`
	Tile3DFileName     *string   `json:"tile_3d_file_name"`
}
