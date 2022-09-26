package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/redhatinsights/payload-tracker-go/internal/config"
	"github.com/redhatinsights/payload-tracker-go/internal/db"
	"github.com/redhatinsights/payload-tracker-go/internal/endpoints"
	"github.com/redhatinsights/payload-tracker-go/internal/kafka"
	"github.com/redhatinsights/payload-tracker-go/internal/logging"
)

func lubdub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("lubdub"))
}

func main() {
	logging.InitLogger()

	cfg := config.Get()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	logging.Log.Info("Setting up DB")
	db.DbConnect(cfg)

	healthHandler := endpoints.HealthCheckHandler(
		db.DB,
		*cfg,
	)

	// FIXME: move the redis initialization to a method
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping(context.TODO()).Result()
	if err != nil {
		logging.Log.Fatal("Unable to connect to redis: ", err)
	}

	logging.Log.Info("Starting a new kafka consumer...")

	// Webserver is created only for metrics collection
	r := chi.NewRouter()

	// Mount the metrics handler on /metrics
	r.Get("/", lubdub)
	r.Get("/live", healthHandler)
	r.Get("/ready", healthHandler)
	r.Handle("/metrics", promhttp.Handler())

	msrv := http.Server{
		Addr:    ":" + cfg.MetricsPort,
		Handler: r,
	}

	consumer, err := kafka.NewConsumer(ctx, cfg, cfg.KafkaConfig.KafkaTopic)

	if err != nil {
		logging.Log.Fatal("ERROR! ", err)
	}

	go func() {
		if err := msrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	var kafkaMessageHandler kafka.MessageHandler
	switch cfg.MessageProcessorImpl {
	case "db":
		fmt.Println("DB Message Handler")
		kafkaMessageHandler = kafka.NewDBBasedMessageHandler(db.DB, cfg)
	case "redis":
		fmt.Println("Redis Message Handler")
		kafkaMessageHandler = kafka.NewRedisBasedMessageHandler(redisClient, cfg)
	default:
		logging.Log.Fatal("Invalid message processor impl")
	}

	go kafka.NewConsumerEventLoop(ctx, consumer, kafkaMessageHandler)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigchan
	logging.Log.Infof("Caught Signal %v: terminating\n", sig)
	cancel()
	consumer.Close()
}
