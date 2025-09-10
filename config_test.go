package main

import (
	"os"
	"testing"
)

func TestReadConfig_ValidFile(t *testing.T) {
	yamlData := `
Influx:
  url: "http://localhost:8086"
  token: "my-token"
  org: "my-org"
  bucket: "my-bucket"
Rabbit:
  Channel: "my-channel"
  Host: "localhost"
  Username: "guest"
  Password: "guest"
  Port: 5672
Api:
  Host: "0.0.0.0"
  Port: 8080
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := ReadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("ReadConfig returned error: %v", err)
	}

	// Influx checks
	if cfg.InfluxdbConfig.Url != "http://localhost:8086" {
		t.Errorf("Expected Influx.Url 'http://localhost:8086', got '%s'", cfg.InfluxdbConfig.Url)
	}
	if cfg.InfluxdbConfig.Token != "my-token" {
		t.Errorf("Expected Influx.Token 'my-token', got '%s'", cfg.InfluxdbConfig.Token)
	}
	if cfg.InfluxdbConfig.Org != "my-org" {
		t.Errorf("Expected Influx.Org 'my-org', got '%s'", cfg.InfluxdbConfig.Org)
	}
	if cfg.InfluxdbConfig.Bucket != "my-bucket" {
		t.Errorf("Expected Influx.Bucket 'my-bucket', got '%s'", cfg.InfluxdbConfig.Bucket)
	}

	// Rabbit checks
	if cfg.Rabbit.Channel != "my-channel" {
		t.Errorf("Expected Rabbit.Channel 'my-channel', got '%s'", cfg.Rabbit.Channel)
	}
	if cfg.Api.Port != 8080 {
		t.Errorf("Expected Api.Port 8080, got %d", cfg.Api.Port)
	}
	if cfg.Api.Host != "0.0.0.0" {
		t.Errorf("Expected Api.Host '0.0.0.0', got '%s'", cfg.Api.Host)
	}
}

func TestReadConfig_FileNotFound(t *testing.T) {
	_, err := ReadConfig("nonexistent.yaml")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
}

func TestReadConfig_InvalidYAML(t *testing.T) {
	invalidYAML := `
Influx:
  url: "http://localhost:8086"
  token: "my-token"
  org: "my-org"
  bucket: "my-bucket"
Rabbit:
  Channel: "my-channel"
  Host: localhost
  Username: guest
  Password: guest
  Port: not-a-number
Api:
  Host: 0.0.0.0
  Port: 8080
`

	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(invalidYAML)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	_, err = ReadConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("Expected YAML unmarshal error, got nil")
	}
}

