package entity

import (
	"time"

	"github.com/google/uuid"
)

type BaseEntityModel struct {
	ID           uuid.UUID  `gorm:"type:VARCHAR(36);column:id;primaryKey" json:"id" mapper:"id"`
	CreatedBy    string     `gorm:"type:VARCHAR(30);column:created_by" json:"created_by"`
	CreatedTime  time.Time  `gorm:"column:created_time" json:"created_time"`
	ModifiedBy   *string    `gorm:"type:VARCHAR(30);column:modified_by" json:"modified_by"`
	ModifiedTime *time.Time `gorm:"column:modified_time" json:"modified_time"`
}

type BaseEntityNoIDModel struct {
	CreatedBy    string     `gorm:"type:VARCHAR(30);column:created_by" json:"created_by"`
	CreatedTime  time.Time  `gorm:"column:created_time" json:"created_time"`
	ModifiedBy   *string    `gorm:"type:VARCHAR(30);column:modified_by" json:"modified_by"`
	ModifiedTime *time.Time `gorm:"column:modified_time" json:"modified_time"`
}
