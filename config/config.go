package config

import (
    "log"
    "os"
    "time"
)

type Config struct {
    AWSRegion       string
    SNSTopicARN     string
    AllowedStartTime time.Time
    AllowedEndTime   time.Time
}

func LoadConfig() Config {
    requiredVars := []string{"AWS_REGION", "SNS_TOPIC_ARN", "ALERT_START_TIME", "ALERT_END_TIME"}
    envVars := make(map[string]string)

    for _, key := range requiredVars {
        value, exists := os.LookupEnv(key)
        if !exists || value == "" {
            log.Fatalf("Missing required environment variable: %s", key)
        }
        envVars[key] = value
    }

    layout := "15:04"

    startTime, err := time.Parse(layout, envVars["ALERT_START_TIME"])
    if err != nil {
        log.Fatalf("Error parsing ALERT_START_TIME: %v", err)
    }

    endTime, err := time.Parse(layout, envVars["ALERT_END_TIME"])
    if err != nil {
        log.Fatalf("Error parsing ALERT_END_TIME: %v", err)
    }

    return Config{
        AWSRegion:       envVars["AWS_REGION"],
        SNSTopicARN:     envVars["SNS_TOPIC_ARN"],
        AllowedStartTime: startTime,
        AllowedEndTime:   endTime,
    }
}
