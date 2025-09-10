package main

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    time.Time
		expectError bool
	}{
		{
			name:        "valid RFC3339 timestamp",
			input:       "2023-10-15T14:30:45Z",
			expected:    time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "valid RFC3339 timestamp with timezone",
			input:       "2023-10-15T14:30:45+02:00",
			expected:    time.Date(2023, 10, 15, 12, 30, 45, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "invalid timestamp format",
			input:       "2023-10-15 14:30:45",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "non-string timestamp",
			input:       123456789,
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "nil timestamp",
			input:       nil,
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "empty string timestamp",
			input:       "",
			expected:    time.Time{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTime(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseTime() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ParseTime() unexpected error: %v", err)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("ParseTime() = %v, expected %v", result, tt.expected)
				}
			}
		})
	}
}

func TestConsumeMessage(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    *Metric
		expectError bool
	}{
		{
			name: "valid message with metric and tags",
			input: []byte(`{
				"metric": {
					"name": "cpu_usage",
					"value": 75.5,
					"time": "2023-10-15T14:30:45Z"
				},
				"host": "server1",
				"region": "us-east-1"
			}`),
			expected: &Metric{
				Name:      "cpu_usage",
				Value:     75.5,
				Timestamp: time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
				Tags: map[string]string{
					"host":   "server1",
					"region": "us-east-1",
				},
			},
			expectError: false,
		},
		{
			name: "metric with string value",
			input: []byte(`{
				"metric": {
					"name": "status",
					"value": "healthy",
					"time": "2023-10-15T14:30:45Z"
				},
				"service": "api"
			}`),
			expected: &Metric{
				Name:      "status",
				Value:     "healthy",
				Timestamp: time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
				Tags: map[string]string{
					"service": "api",
				},
			},
			expectError: false,
		},
		{
			name: "metric with integer value",
			input: []byte(`{
				"metric": {
					"name": "request_count",
					"value": 1000,
					"time": "2023-10-15T14:30:45Z"
				}
			}`),
			expected: &Metric{
				Name:      "request_count",
				Value:     float64(1000), // JSON unmarshals numbers as float64
				Timestamp: time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
				Tags:      map[string]string{},
			},
			expectError: false,
		},
		{
			name: "message with underscore prefixed fields (should be ignored)",
			input: []byte(`{
				"metric": {
					"name": "memory_usage",
					"value": 60.2,
					"time": "2023-10-15T14:30:45Z"
				},
				"_internal": "should_be_ignored",
				"_debug": "also_ignored",
				"environment": "prod"
			}`),
			expected: &Metric{
				Name:      "memory_usage",
				Value:     60.2,
				Timestamp: time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
				Tags: map[string]string{
					"environment": "prod",
				},
			},
			expectError: false,
		},
		{
			name: "message without metric field",
			input: []byte(`{
				"host": "server1",
				"region": "us-east-1"
			}`),
			expected: &Metric{
				Name:      "",
				Value:     nil,
				Timestamp: time.Time{},
				Tags: map[string]string{
					"host":   "server1",
					"region": "us-east-1",
				},
			},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			input:       []byte(`{invalid json`),
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid metric JSON - malformed object",
			input: []byte(`{
				"metric": {invalid},
				"host": "server1"
			}`),
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid metric JSON - missing quotes",
			input: []byte(`{
				"metric": {name: "test", value: 123, time: "2023-10-15T14:30:45Z"},
				"host": "server1"
			}`),
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid metric JSON - array instead of object",
			input: []byte(`{
				"metric": ["name", "value", "time"],
				"host": "server1"
			}`),
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid metric JSON - string instead of object",
			input: []byte(`{
				"metric": "not an object",
				"host": "server1"
			}`),
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid metric JSON - number instead of object",
			input: []byte(`{
				"metric": 12345,
				"host": "server1"
			}`),
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid timestamp in metric",
			input: []byte(`{
				"metric": {
					"name": "cpu_usage",
					"value": 75.5,
					"time": "invalid-time"
				}
			}`),
			expected:    nil,
			expectError: true,
		},
		{
			name: "non-string tag value",
			input: []byte(`{
				"metric": {
					"name": "cpu_usage",
					"value": 75.5,
					"time": "2023-10-15T14:30:45Z"
				},
				"count": 123
			}`),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "empty JSON object",
			input:       []byte(`{}`),
			expected: &Metric{
				Name:      "",
				Value:     nil,
				Timestamp: time.Time{},
				Tags:      map[string]string{},
			},
			expectError: false,
		},
		{
			name: "metric with null value",
			input: []byte(`{
				"metric": {
					"name": "null_test",
					"value": null,
					"time": "2023-10-15T14:30:45Z"
				}
			}`),
			expected: &Metric{
				Name:      "null_test",
				Value:     nil,
				Timestamp: time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
				Tags:      map[string]string{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConsumeMessage(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("ConsumeMessage() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ConsumeMessage() unexpected error: %v", err)
				return
			}
			
			if result == nil && tt.expected != nil {
				t.Errorf("ConsumeMessage() returned nil, expected %+v", tt.expected)
				return
			}
			
			if result != nil && tt.expected == nil {
				t.Errorf("ConsumeMessage() returned %+v, expected nil", result)
				return
			}
			
			if result != nil && tt.expected != nil {
				if result.Name != tt.expected.Name {
					t.Errorf("ConsumeMessage() Name = %v, expected %v", result.Name, tt.expected.Name)
				}
				
				if !reflect.DeepEqual(result.Value, tt.expected.Value) {
					t.Errorf("ConsumeMessage() Value = %v (%T), expected %v (%T)", 
						result.Value, result.Value, tt.expected.Value, tt.expected.Value)
				}
				
				if !result.Timestamp.Equal(tt.expected.Timestamp) {
					t.Errorf("ConsumeMessage() Timestamp = %v, expected %v", 
						result.Timestamp, tt.expected.Timestamp)
				}
				
				if !reflect.DeepEqual(result.Tags, tt.expected.Tags) {
					t.Errorf("ConsumeMessage() Tags = %v, expected %v", result.Tags, tt.expected.Tags)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseTime(b *testing.B) {
	timestamp := "2023-10-15T14:30:45Z"
	for i := 0; i < b.N; i++ {
		_, _ = ParseTime(timestamp)
	}
}

func BenchmarkConsumeMessage(b *testing.B) {
	data := []byte(`{
		"metric": {
			"name": "cpu_usage",
			"value": 75.5,
			"time": "2023-10-15T14:30:45Z"
		},
		"host": "server1",
		"region": "us-east-1"
	}`)
	
	for i := 0; i < b.N; i++ {
		_, _ = ConsumeMessage(data)
	}
}

// Test helper functions
func TestRawMetricUnmarshal(t *testing.T) {
	jsonData := []byte(`{
		"name": "test_metric",
		"value": 42.5,
		"time": "2023-10-15T14:30:45Z"
	}`)
	
	var raw RawMetric
	err := json.Unmarshal(jsonData, &raw)
	if err != nil {
		t.Errorf("Failed to unmarshal RawMetric: %v", err)
	}
	
	if raw.Name != "test_metric" {
		t.Errorf("RawMetric Name = %v, expected test_metric", raw.Name)
	}
	
	if raw.Value != 42.5 {
		t.Errorf("RawMetric Value = %v, expected 42.5", raw.Value)
	}
	
	if raw.Timestamp != "2023-10-15T14:30:45Z" {
		t.Errorf("RawMetric Timestamp = %v, expected 2023-10-15T14:30:45Z", raw.Timestamp)
	}
}
