package main

import (
	"fmt"
	"os"
	"path/filepath"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")	
	if configPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
			os.Exit(1)
		}

		configPath = filepath.Join(cwd, "config.yml")
	}

	cfg, err := ReadConfig(configPath)
	if err != nil {
		fmt.Printf("Cannot read config file:\n%v\n", err)
		return
	}

	client := influxdb2.NewClient(cfg.InfluxdbConfig.Url, cfg.InfluxdbConfig.Token)
	writeAPI := client.WriteAPIBlocking(cfg.InfluxdbConfig.Org, cfg.InfluxdbConfig.Bucket)

	msgs, err := ConsumeMessages(cfg)
	if err != nil {
		fmt.Printf("Cannot consume messages from rabbitmq:\n%v\n", err)
		return 
	}

	fmt.Printf("Waiting for messages...\n")
	for msg := range msgs {
		fmt.Printf("Received: %s\n", msg.Body)
		metric, err := ConsumeMessage(msg.Body)
		if err != nil {
			fmt.Printf("Cannot consume rabbit msg:\n%v\n", err)
			continue
		}

		if err := SendMetric(writeAPI, metric); err != nil {
			fmt.Printf("Cannot send metric to influxdb:\n%v\n", err)	
			continue
		}

		msg.Ack(false)
	}
}
		
