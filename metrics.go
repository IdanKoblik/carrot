package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type RawMetric struct {
	Name      string      `json:"name"`
	Value     any         `json:"value"`
	Timestamp any         `json:"time"`
}

type Metric struct {
	Name      string
	Value     any
	Timestamp time.Time
	Tags      map[string]string
}

func ParseTime(ts any) (time.Time, error) {
	switch v := ts.(type) {
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	case float64:
		sec := int64(v)
		nsecFloat := v - float64(sec)
		nsec := int64(nsecFloat * 1e9)
		return time.Unix(sec, nsec), nil
	case int64:
		seconds := v
		return time.Unix(seconds, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported timestamp type: %T", ts)
	}
}

func ConsumeMessage(data []byte) ([]*Metric, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	tags := make(map[string]string)
	for k, v := range raw {
		if k == "metrics" || strings.HasPrefix(k, "_") {
			continue
		}

		var val string
		if err := json.Unmarshal(v, &val); err != nil {
			return nil, fmt.Errorf("invalid tag value for key %s: %v", k, err)
		}

		tags[k] = val
	}

	var rawMetrics []RawMetric
	if err := json.Unmarshal(raw["metrics"], &rawMetrics); err != nil {
		return nil, fmt.Errorf("invalid metrics array: %v", err)
	}

	var metrics []*Metric
	for _, rawMetric := range rawMetrics {
		t, err := ParseTime(rawMetric.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp: %v", err)
		}

		metric := &Metric{
			Name:      rawMetric.Name,
			Value:     rawMetric.Value,
			Timestamp: t,
			Tags:      tags,
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}
