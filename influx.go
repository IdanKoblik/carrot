package main

import (
	"context"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

func SendMetric(writeAPI api.WriteAPIBlocking, metrics []*Metric) error {
	var points []*write.Point	

	for _, metric := range metrics {
		point := influxdb2.NewPoint(
			metric.Name,
			metric.Tags,
			map[string]interface{}{metric.Name: metric.Value},
			metric.Timestamp,
		)

		points = append(points, point)
	}

	return writeAPI.WritePoint(context.Background(), points...)
}

