package models

import (
	"time"
)

type Payload struct {
    Id          uint      `json:"id" gorm:"primaryKey;autoIncrement:true;type:bigint"`
	RequestId   string    `json:"request_id" gorm:"not null;type:varchar"`
	Account     string    `json:"account" gorm:"type:varchar"`
	OrgId       string    `json:"org_id" gorm:"varchar"`
	InventoryId string    `json:"inventory_id" gorm:"type:varchar"`
	SystemId    string    `json:"system_id" gorm:"type:varchar"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
    Service     string `gorm:type:varchar"`
    Source     string `gorm:type:varchar"`
    Status      string `gorm:type:varchar"`
	StatusMsg string    `gorm:"type:varchar"`
	Date      time.Time `gorm:"primaryKey;not null"`
}
