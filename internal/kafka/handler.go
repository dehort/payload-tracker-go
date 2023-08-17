package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"

	"github.com/redhatinsights/payload-tracker-go/internal/config"
	"github.com/redhatinsights/payload-tracker-go/internal/endpoints"
	l "github.com/redhatinsights/payload-tracker-go/internal/logging"
	models "github.com/redhatinsights/payload-tracker-go/internal/models/db"
	"github.com/redhatinsights/payload-tracker-go/internal/models/message"
	"github.com/redhatinsights/payload-tracker-go/internal/queries"
)

var (
	tableNames = []string{"service", "source", "status"}
)

type MessageHandler interface {
	onMessage(ctx context.Context, msg *kafka.Message)
}

func NewDBBasedMessageHandler(db *gorm.DB, cfg *config.TrackerConfig) MessageHandler {
	return &DBBasedMessageHandler{
		db:     db,
		config: cfg,
	}
}

type DBBasedMessageHandler struct {
	db     *gorm.DB
	config *config.TrackerConfig
}

// OnMessage takes in each payload status message and processes it
func (this *DBBasedMessageHandler) onMessage(ctx context.Context, msg *kafka.Message) {

	callDurationTimer := prometheus.NewTimer(metrics.dbInsertDuration)
	defer callDurationTimer.ObserveDuration()

	l.Log.Debug("Processing Payload Message ", msg.Value)

	payloadStatus := &message.PayloadStatusMessage{}
	sanitizedPayloadStatus := &models.PayloadStatuses{}

	if err := json.Unmarshal(msg.Value, payloadStatus); err != nil {
		// PROBE: Add probe here for error unmarshaling JSON
		if cfg.DebugConfig.LogStatusJson {
			l.Log.Error("ERROR: Unmarshaling Payload Status Event: ", err, " Raw Message: ", string(msg.Value))
		} else {
			l.Log.Error("ERROR: Unmarshaling Payload Status Event: ", err)
		}
		return
	}

	if !validateRequestID(cfg.RequestConfig.ValidateRequestIDLength, payloadStatus.RequestID) {
		return
	}

	// Sanitize the payload
	sanitizePayload(payloadStatus)

	// Upsert into Payloads Table
	payload := createPayload(payloadStatus)

	upsertResult, payloadId := queries.UpsertPayloadByRequestId(this.db, payloadStatus.RequestID, payload)
	if upsertResult.Error != nil {
		l.Log.Error("ERROR Payload table upsert failed: ", upsertResult.Error)
		return
	}
	sanitizedPayloadStatus.PayloadId = payloadId

	// Check if service/source/status are in table
	// this section checks the subsiquent DB tables to see if the service_id, source_id, and status_id exist for the given message
	l.Log.Debug("Adding Status, Sources, and Services to sanitizedPayload")

	// Status & Service: Always defined in the message
	existingStatus := queries.GetStatusByName(this.db, payloadStatus.Status)
	if (models.Statuses{}) == existingStatus {
		statusResult, newStatus := queries.CreateStatusTableEntry(this.db, payloadStatus.Status)
		if statusResult.Error != nil {
			l.Log.Error("Error Creating Statuses Table Entry ERROR: ", statusResult.Error)
			return
		}

		sanitizedPayloadStatus.Status = newStatus
	} else {
		sanitizedPayloadStatus.Status = existingStatus
	}

	existingService := queries.GetServiceByName(this.db, payloadStatus.Service)
	if (models.Services{}) == existingService {
		serviceResult, newService := queries.CreateServiceTableEntry(this.db, payloadStatus.Service)
		if serviceResult.Error != nil {
			l.Log.Error("Error Creating Service Table Entry ERROR: ", serviceResult.Error)
			return
		}

		sanitizedPayloadStatus.Service = newService
	} else {
		sanitizedPayloadStatus.Service = existingService
	}

	// Sources
	if payloadStatus.Source != "" {
		existingSource := queries.GetSourceByName(this.db, payloadStatus.Source)
		if (models.Sources{}) == existingSource {
			result, newSource := queries.CreateSourceTableEntry(this.db, payloadStatus.Source)
			if result.Error != nil {
				l.Log.Error("Error Creating Sources Table Entry ERROR: ", result.Error)
				return
			}

			sanitizedPayloadStatus.Source = newSource
		} else {
			sanitizedPayloadStatus.Source = existingSource
		}
	}

	if payloadStatus.StatusMSG != "" {
		sanitizedPayloadStatus.StatusMsg = payloadStatus.StatusMSG
	}

	// Insert Date
	sanitizedPayloadStatus.Date = payloadStatus.Date.Time

	fmt.Println("inserting payload")
	// Insert payload into DB
	endpoints.ObserveMessageProcessTime(time.Since(start))
	endpoints.IncMessagesProcessed()

	result := queries.InsertPayloadStatus(this.db, sanitizedPayloadStatus)
	fmt.Println("inserted payload")
	fmt.Println("result:", result)
	if result.Error != nil {
		endpoints.IncMessageProcessErrors()
		l.Log.Debug("Failed to insert sanitized PayloadStatus with ERROR: ", result.Error)
		result = queries.InsertPayloadStatus(this.db, sanitizedPayloadStatus)
		if result.Error != nil {
			l.Log.Debug("Failed to re-insert sanitized PayloadStatus with ERROR: ", result.Error)
			result = queries.InsertPayloadStatus(this.db, sanitizedPayloadStatus)
			if result.Error != nil {
				l.Log.Error("Failed final attempt to re-insert PayloadStatus with ERROR: ", result.Error)
			}
		}
	}
}

func validateRequestID(requestIDLength int, requestID string) bool {
	if requestIDLength != 0 {
		if len(requestID) != requestIDLength {
			endpoints.IncInvalidConsumerRequestIDs()
			return false
		}
	}

	return true
}

func sanitizePayload(msg *message.PayloadStatusMessage) {
	// Set default fields to lowercase
	msg.Service = strings.ToLower(msg.Service)
	msg.Status = strings.ToLower(msg.Status)
	if msg.Source != "" {
		msg.Source = strings.ToLower((msg.Source))
	}
}

func createPayload(msg *message.PayloadStatusMessage) (table models.Payloads) {
	payloadTable := models.Payloads{
		Id:          msg.PayloadID,
		RequestId:   msg.RequestID,
		Account:     msg.Account,
		OrgId:       msg.OrgID,
		SystemId:    msg.SystemID,
		CreatedAt:   msg.Date.Time,
		InventoryId: msg.InventoryID,
	}

	return payloadTable
}
