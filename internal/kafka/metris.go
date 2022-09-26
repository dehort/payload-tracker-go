package kafka

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type messageProcessorMetrics struct {
    dbInsertDuration prometheus.Histogram
    redisInsertDuration prometheus.Histogram
}

var metrics *messageProcessorMetrics

func init() {
    metrics = new(messageProcessorMetrics)

    metrics.dbInsertDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "payload_tracker_sql_insert_duration",
    })

    metrics.redisInsertDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "payload_tracker_redis_insert_duration",
    })
}
