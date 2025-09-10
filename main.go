package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var Log *log.Logger

func main() {
	Log = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "Carrot 🥕 ",
	})

	configPath := os.Getenv("CONFIG_PATH")	
	if configPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			Log.Error("Failed to get working directory", "err", err)
			os.Exit(1)
		}

		configPath = filepath.Join(cwd, "config.yml")
	}

	cfg, err := ReadConfig(configPath)
	if err != nil {
		Log.Error("Cannot read config file", "err", err)
		return
	}

	client := influxdb2.NewClient(cfg.InfluxdbConfig.Url, cfg.InfluxdbConfig.Token)
	writeAPI := client.WriteAPIBlocking(cfg.InfluxdbConfig.Org, cfg.InfluxdbConfig.Bucket)

	msgs, err := ConsumeMessages(cfg)
	if err != nil {
		Log.Error("Cannot consume messages from rabbitmq", "err", err)
		return 
	}

	Log.Info("Waiting for messages...")
	for msg := range msgs {
		unquoted, err := strconv.Unquote(`"` + string(msg.Body) + `"`)
    	if err != nil {
        fmt.Println("Error:", err)
        return
    	}

		Log.Info("Received! ", "body", unquoted)
		metric, err := ConsumeMessage(msg.Body)
		if err != nil {
			Log.Error("Cannot consume rabbit msg", "err", err)
			continue
		}

		if err := SendMetric(writeAPI, metric); err != nil {
			Log.Error("Cannot send metric to influxdb", "err", err)
			continue
		}

		Log.Info("Send new metric to influxdb")
		msg.Ack(false)
	}
}
		
