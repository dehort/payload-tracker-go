package kafka

import (
	"context"
    "encoding/json"
    "fmt"
    "time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
    "github.com/prometheus/client_golang/prometheus"

	"github.com/redhatinsights/payload-tracker-go/internal/config"
    "github.com/redhatinsights/payload-tracker-go/internal/logging"
)

func NewRedisBasedMessageHandler(redis *redis.Client, cfg *config.TrackerConfig) MessageHandler {
	return &RedisBasedMessageHandler{
		redis:     redis,
		config: cfg,
	}
}

type RedisBasedMessageHandler struct {
    redis *redis.Client
	config *config.TrackerConfig
}

func (this *RedisBasedMessageHandler) onMessage(ctx context.Context, msg *kafka.Message) {

    callDurationTimer := prometheus.NewTimer(metrics.redisInsertDuration)
    defer callDurationTimer.ObserveDuration()

	//logging.Log.Debug("Processing Payload Message ", msg.Value)

    payloadStatus := map[string]interface{}{}

    if err := json.Unmarshal(msg.Value, &payloadStatus); err != nil {
		// PROBE: Add probe here for error unmarshaling JSON
		logging.Log.Error("ERROR: Unmarshaling Payload Status Event: ", err)
		return
	}

    //fmt.Println("payloadStatus:", payloadStatus)

    requestID, err := pluckRequestID(payloadStatus)
    if err != nil {
        logging.Log.Debug("pluck request id failure:", err)
		//endpoints.IncInvalidConsumerRequestIDs()
        return
    }

    logging.Log.Debug("requestID: ", requestID)

    if !isValidRequestID(requestID, this.config) {
        logging.Log.Debug("invalid request id:", err)
		//endpoints.IncInvalidConsumerRequestIDs()
        return
    }

    /*
	// Sanitize the payload
	sanitizePayload(payloadStatus)
    */
    writeMessageToRedis(ctx, this.redis, requestID,  payloadStatus)
}

func pluckRequestID(statusMessage map[string]interface{}) (string, error) {
    requestID, ok := statusMessage["request_id"]
    if !ok {
        return "", fmt.Errorf("request_id not in message")
    }

    requestIDString, ok := requestID.(string)
    if !ok {
        return "", fmt.Errorf("Cannot convert request_id to string")
    }

    return requestIDString, nil
}

func isValidRequestID(requestID string, cfg *config.TrackerConfig) bool {

	if cfg.RequestConfig.ValidateRequestIDLength != 0 {
        fmt.Println("len(requestID):", len(requestID))
        if len(requestID) != cfg.RequestConfig.ValidateRequestIDLength {
			return false
		}
	}

    return true
}

func writeMessageToRedis(ctx context.Context, rdb *redis.Client, requestID string, statusMessage map[string]interface{}) error {

    //logging.Log.Debug("Writing status to redis - ", requestID)
    fmt.Println("Writing status to redis - ", requestID)
    fmt.Println("date: ", statusMessage["date"])

    delete(statusMessage, "request_id")

    statusMessageJsonStr, err := json.Marshal(statusMessage)
    if err != nil {
        return err
    }

    /*
    var timestamp timeutil.Timestamp
    timestamp, err = timeutil.TimestampFromString(statusMessage["date"].(string))
    if err != nil {
        fmt.Println("Couldn't read timestamp")
    }
    */

    t, _ := time.Parse(time.RFC3339, statusMessage["date"].(string))
    timestamp := float64(t.UnixMilli())

    // FIXME: what about expring these??
    err = rdb.ZAdd(ctx, requestID, &redis.Z{ Score: timestamp, Member: statusMessageJsonStr}).Err()
    if err != nil {
        fmt.Println("Failed to write message to redis: ", err)
        return err
    }

    return nil
}
