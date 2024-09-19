package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type SNSTopicConfig struct {
	Name      string `yaml:"name"`
	ARN       string `yaml:"arn"`
	StartTime string `yaml:"start_time"`
	EndTime   string `yaml:"end_time"`
}

type Config struct {
	AWSRegion        string           `yaml:"aws_region"`
	Topics           []SNSTopicConfig `yaml:"sns_topics"`
	AlertNames       []string         `yaml:"alertnames"`
	BatchWaitSeconds int              `yaml:"batch_wait_seconds"`
}

var readFile = os.ReadFile

func LoadConfig() Config {
	file, err := readFile("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// Validate required fields
	if cfg.AWSRegion == "" || len(cfg.Topics) == 0 {
		log.Fatal("Missing required fields in config file")
	}

	if cfg.BatchWaitSeconds <= 0 {
		log.Fatal("batch_wait_seconds must be a positive integer")
	}

	return cfg
}
