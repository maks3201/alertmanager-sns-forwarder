package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type SNSTopicConfig struct {
	Name       string   `yaml:"name"`
	ARN        string   `yaml:"arn"`
	StartTime  string   `yaml:"start_time"`
	EndTime    string   `yaml:"end_time"`
	DaysOfWeek []string `yaml:"days_of_week"`
}

type ServerTimeouts struct {
	ReadTimeoutSeconds       int `yaml:"read_timeout_seconds"`
	WriteTimeoutSeconds      int `yaml:"write_timeout_seconds"`
	IdleTimeoutSeconds       int `yaml:"idle_timeout_seconds"`
	ReadHeaderTimeoutSeconds int `yaml:"read_header_timeout_seconds"`
}

type AWSTimeouts struct {
	DialTimeoutSeconds           int `yaml:"dial_timeout_seconds"`
	TLSHandshakeTimeoutSeconds   int `yaml:"tls_handshake_timeout_seconds"`
	ResponseHeaderTimeoutSeconds int `yaml:"response_header_timeout_seconds"`
	ExpectContinueTimeoutSeconds int `yaml:"expect_continue_timeout_seconds"`
	IdleConnTimeoutSeconds       int `yaml:"idle_conn_timeout_seconds"`
	MaxIdleConns                 int `yaml:"max_idle_conns"`
	APICallTimeoutSeconds        int `yaml:"api_call_timeout_seconds"`
}

type Timeouts struct {
	Server ServerTimeouts `yaml:"server"`
	AWS    AWSTimeouts    `yaml:"aws"`
}

type Config struct {
	AWSRegion        string           `yaml:"aws_region"`
	AWSAccessKey     string           `yaml:"aws_access_key"`
	AWSSecretKey     string           `yaml:"aws_secret_key"`
	Topics           []SNSTopicConfig `yaml:"sns_topics"`
	AlertNames       []string         `yaml:"alertnames"`
	BatchWaitSeconds int              `yaml:"batch_wait_seconds"`
	Timeouts         Timeouts         `yaml:"timeouts"`
	LogLevel         string           `yaml:"log_level"`
}

var readFile = os.ReadFile

func LoadConfig(configFilePath string) Config {
	file, err := readFile(configFilePath)
	if err != nil {
		log.Fatalf("Failed to read config file '%s': %v", configFilePath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config file '%s': %v", configFilePath, err)
	}

	if cfg.AWSRegion == "" || len(cfg.Topics) == 0 {
		log.Fatal("Missing required fields in config file")
	}

	if cfg.BatchWaitSeconds <= 0 {
		log.Fatal("batch_wait_seconds must be a positive integer")
	}

	setLogLevel(cfg.LogLevel)

	setDefaultTimeouts(&cfg)

	return cfg
}

func setLogLevel(logLevel string) {
	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		log.Debug("Log level set to DEBUG")
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.InfoLevel)
		log.Warnf("Invalid log level '%s', defaulting to INFO", logLevel)
	}
}

func setDefaultTimeouts(cfg *Config) {
	if cfg.Timeouts.Server.ReadTimeoutSeconds == 0 {
		cfg.Timeouts.Server.ReadTimeoutSeconds = 5
	}
	if cfg.Timeouts.Server.WriteTimeoutSeconds == 0 {
		cfg.Timeouts.Server.WriteTimeoutSeconds = 5
	}
	if cfg.Timeouts.Server.IdleTimeoutSeconds == 0 {
		cfg.Timeouts.Server.IdleTimeoutSeconds = 60
	}
	if cfg.Timeouts.Server.ReadHeaderTimeoutSeconds == 0 {
		cfg.Timeouts.Server.ReadHeaderTimeoutSeconds = 5
	}

	if cfg.Timeouts.AWS.DialTimeoutSeconds == 0 {
		cfg.Timeouts.AWS.DialTimeoutSeconds = 5
	}
	if cfg.Timeouts.AWS.TLSHandshakeTimeoutSeconds == 0 {
		cfg.Timeouts.AWS.TLSHandshakeTimeoutSeconds = 5
	}
	if cfg.Timeouts.AWS.ResponseHeaderTimeoutSeconds == 0 {
		cfg.Timeouts.AWS.ResponseHeaderTimeoutSeconds = 10
	}
	if cfg.Timeouts.AWS.ExpectContinueTimeoutSeconds == 0 {
		cfg.Timeouts.AWS.ExpectContinueTimeoutSeconds = 1
	}
	if cfg.Timeouts.AWS.IdleConnTimeoutSeconds == 0 {
		cfg.Timeouts.AWS.IdleConnTimeoutSeconds = 90
	}
	if cfg.Timeouts.AWS.MaxIdleConns == 0 {
		cfg.Timeouts.AWS.MaxIdleConns = 100
	}
	if cfg.Timeouts.AWS.APICallTimeoutSeconds == 0 {
		cfg.Timeouts.AWS.APICallTimeoutSeconds = 10
	}
}
