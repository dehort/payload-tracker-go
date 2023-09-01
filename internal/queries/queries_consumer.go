package queries

import (
	models "github.com/redhatinsights/payload-tracker-go/internal/models/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	StatusColumns = "payload_id, status_id, service_id, source_id, date, inventory_id, system_id, account, org_id"
	PayloadJoins  = "left join Payloads on Payloads.id = PayloadStatuses.payload_id"
)

func GetPayloadByRequestId(db *gorm.DB, request_id string) (result models.Payload, err error) {
	var payload models.Payload
	if results := db.Where("request_id = ?", request_id).First(&payload); results.Error != nil {
		return payload, results.Error
	}

	return payload, nil
}

func UpsertPayloadByRequestId(db *gorm.DB, request_id string, payload models.Payload) (tx *gorm.DB, payloadId uint) {
	result := db.Model(&payload).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "request_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"account", "inventory_id", "system_id", "org_id"}),
		}).
		Create(&payload)
	return result, payload.Id
}

func UpdatePayloadsTable(db *gorm.DB, updates models.Payload, payloads models.Payload) (tx *gorm.DB) {
	return db.Model(&payloads).Omit("request_id", "Id").Updates(updates)
}

func CreatePayloadTableEntry(db *gorm.DB, newPayload models.Payload) (result *gorm.DB, payload models.Payload) {
	results := db.Create(&newPayload)

	return results, newPayload
}

/*
var bufferedPayloadStatus []*models.PayloadStatuses
*/
func InsertPayload(db *gorm.DB, payload *models.Payload) (tx *gorm.DB) {
    return db.Create(payload)
}

/*
func InsertPayloadStatus(db *gorm.DB, payloadStatus *models.PayloadStatuses) (tx *gorm.DB) {
	if (models.Sources{}) == payloadStatus.Source {
		return db.Omit("source_id").Create(&payloadStatus)
	}

    batchSize := 100

    if bufferedPayloadStatus == nil {
        bufferedPayloadStatus = make([]*models.PayloadStatuses, 0, batchSize)
    }

    bufferedPayloadStatus = append(bufferedPayloadStatus, payloadStatus)

    if len(bufferedPayloadStatus) >= batchSize {
        fmt.Println("Bulk create")
        db.CreateInBatches(bufferedPayloadStatus, batchSize)
        //bufferedPayloadStatus = nil
        bufferedPayloadStatus = bufferedPayloadStatus[:0]
    } else {
        fmt.Println("SKIP!!")
    }

    return db

	//return db.Create(&payloadStatus)
}
*/
