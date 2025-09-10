package main

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

func SendMetric(writeAPI api.WriteAPIBlocking, metric *Metric) error {
	point := influxdb2.NewPoint(
		metric.Name,
		metric.Tags,
		map[string]interface{}{metric.Name: metric.Value},
		metric.Timestamp,
	)

	return writeAPI.WritePoint(context.Background(), point)
}

