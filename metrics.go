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
	default:
		return time.Time{}, fmt.Errorf("unsupported timestamp type: %T", ts)
	}
}

func ConsumeMessage(data []byte) (*Metric, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	metric := &Metric{
		Tags: make(map[string]string),
	}

	for k, v := range raw {
		if strings.HasPrefix(k, "_") {
			continue
		}

		if k == "metric" {
			var temp RawMetric
			if err := json.Unmarshal(v, &temp); err != nil {
				return nil, err
			}

			t, err := ParseTime(temp.Timestamp)
			if err != nil {
				return nil, err
			}

			metric.Timestamp = t
			metric.Name = temp.Name
			metric.Value = temp.Value
		} else {
			var val string
			if err := json.Unmarshal(v, &val); err != nil {
				return nil, err
			}
			metric.Tags[k] = val
		}
	}

	return metric, nil
}
