package queries

import (
    "fmt"

	models "github.com/redhatinsights/payload-tracker-go/internal/models/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	StatusColumns = "payload_id, status_id, service_id, source_id, date, inventory_id, system_id, account, org_id"
	PayloadJoins  = "left join Payloads on Payloads.id = PayloadStatuses.payload_id"
)

var (
	g_services map[string]models.Services
	g_statuses map[string]models.Statuses
	sources  []models.Sources
)

func init() {
    g_services = make(map[string]models.Services)
    g_statuses = make(map[string]models.Statuses)
}

func GetServiceByName(db *gorm.DB, service_id string) models.Services {
    fmt.Println("GetServicByName")

/*
    serv, ok := g_services[service_id]
    if ok {
        fmt.Println("\t cached")
        return serv
    }
*/

	var service models.Services
	db.Where("name = ?", service_id).First(&service)
/*
	if (models.Services{}) != service {
        g_services[service_id] = service
    }
*/
	return service
}

func GetStatusByName(db *gorm.DB, status_id string) models.Statuses {
    fmt.Println("GetStatusByName")
/*
    stat, ok := g_statuses[status_id]
    if ok {
        fmt.Println("\t cached")
        return stat
    }
*/

	var status models.Statuses
	db.Where("name = ?", status_id).First(&status)
/*
	if (models.Statuses{}) != status {
        g_statuses[status_id] = status
    }
*/
	return status
}

func GetSourceByName(db *gorm.DB, source_id string) models.Sources {
    fmt.Println("GetSourceByName")
	var source models.Sources
	db.Where("name = ?", source_id).First(&source)
	return source
}

func GetPayloadByRequestId(db *gorm.DB, request_id string) (result models.Payloads, err error) {
	var payload models.Payloads
	if results := db.Where("request_id = ?", request_id).First(&payload); results.Error != nil {
		return payload, results.Error
	}

	return payload, nil
}

func UpsertPayloadByRequestId(db *gorm.DB, request_id string, payload models.Payloads) (tx *gorm.DB, payloadId uint) {
	result := db.Model(&payload).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "request_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"account", "inventory_id", "system_id", "org_id"}),
		}).
		Create(&payload)
	return result, payload.Id
}

func UpdatePayloadsTable(db *gorm.DB, updates models.Payloads, payloads models.Payloads) (tx *gorm.DB) {
	return db.Model(&payloads).Omit("request_id", "Id").Updates(updates)
}

func CreatePayloadTableEntry(db *gorm.DB, newPayload models.Payloads) (result *gorm.DB, payload models.Payloads) {
	results := db.Create(&newPayload)

	return results, newPayload
}

func CreateStatusTableEntry(db *gorm.DB, name string) (result *gorm.DB, status models.Statuses) {
	newStatus := models.Statuses{Name: name}
	results := db.Create(&newStatus)

	return results, newStatus
}

func CreateSourceTableEntry(db *gorm.DB, name string) (result *gorm.DB, source models.Sources) {
	newSource := models.Sources{Name: name}
	results := db.Create(&newSource)

	return results, newSource
}

func CreateServiceTableEntry(db *gorm.DB, name string) (result *gorm.DB, service models.Services) {
    fmt.Println("CreateServiceTableEntry")
	newService := models.Services{Name: name}
	results := db.Create(&newService)

	return results, newService
}

func InsertPayloadStatus(db *gorm.DB, payloadStatus *models.PayloadStatuses) (tx *gorm.DB) {
	if (models.Sources{}) == payloadStatus.Source {
		return db.Omit("source_id").Create(&payloadStatus)
	}
	return db.Create(&payloadStatus)
}
