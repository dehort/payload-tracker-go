package main

import (
	"context"
	"fmt"
	"time"

	"crypto/rand"

	"github.com/go-redis/redis/v8"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/redhatinsights/payload-tracker-go/internal/config"
	"github.com/redhatinsights/payload-tracker-go/internal/db"
	"github.com/redhatinsights/payload-tracker-go/internal/kafka"
	"github.com/redhatinsights/payload-tracker-go/internal/logging"
)

func main() {
	logging.InitLogger()

	cfg := config.Get()

	ctx := context.Background()

	fmt.Println("Db config:", cfg.DatabaseConfig)
	db.DbConnect(cfg)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping(context.TODO()).Result()
	if err != nil {
		logging.Log.Fatal("Unable to connect to redis: ", err)
	}


/*
    handler := kafka.NewRedisBasedMessageHandler(redisClient, cfg)
*/
	handler := kafka.NewDBBasedMessageHandler(db.DB, cfg)

	date := time.Now()

	//    jsonPayload := fmt.Sprintf("{\"date\": \"%s\", \"service\": \"fred\", \"request_id\": \"11111111111111111111111111111111\", \"account\": \"1\", \"inventory_id\": \"1\", \"system_id\": \"1\"}", date.Format(time.RFC3339))

	//    msg := confluent.Message{Value: []byte(jsonPayload)}

	for i := 0; i < 10000; i++ {

		jsonPayload := fmt.Sprintf("{\"date\": \"%s\", \"service\": \"fred\", \"request_id\": \"%s\", \"account\": \"1\", \"inventory_id\": \"1\", \"system_id\": \"1\", \"source\": \"FIXME\"}", date.Format(time.RFC3339), pseudo_uuid())
		msg := confluent.Message{Value: []byte(jsonPayload)}
		//handler.OnMessage(ctx, &msg, cfg)
		handler.OnMessage(ctx, &msg)
	}
}

// Note - NOT RFC4122 compliant
func pseudo_uuid() (uuid string) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%X%X%X%X%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return
}
