package aws

import (
    "context"
    "fmt"
    "log"

    awsconfig "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sns"
    "github.com/maks3201/sns-alert-service/config"
)

var snsClient *sns.Client

// InitSNSClient initializes the AWS SNS client
func InitSNSClient(cfg config.Config) error {
    awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(cfg.AWSRegion))
    if err != nil {
        return fmt.Errorf("failed to load AWS configuration: %v", err)
    }
    snsClient = sns.NewFromConfig(awsCfg)
    log.Printf("Connected to AWS region: %s", awsCfg.Region)
    return nil
}

// PublishToSNS publishes a message to an SNS topic
func PublishToSNS(topicArn string, message string) error {
    _, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
        TopicArn: &topicArn,
        Message:  &message,
    })
    if err != nil {
        return fmt.Errorf("failed to publish message to SNS: %v", err)
    }
    return nil
}
