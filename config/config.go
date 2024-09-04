package config

import (
    "log"
    "os"
)

type Config struct {
    AWSRegion    string
    SNSTopicARN  string
}

func LoadConfig() Config {
    cfg := Config{
        AWSRegion:   getEnv("AWS_REGION", ""),
        SNSTopicARN: getEnv("SNS_TOPIC_ARN", ""),
    }

    if cfg.AWSRegion == "" || cfg.SNSTopicARN == "" {
        log.Fatal("Missing required environment variables")
    }

    return cfg
}

func getEnv(key string, defaultValue string) string {
    value, exists := os.LookupEnv(key)
    if !exists {
        return defaultValue
    }
    return value
}
