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

// InitSNSClient initializes the AWS SNS client and verifies credentials
func InitSNSClient(cfg config.Config) error {
    awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(cfg.AWSRegion))
    if err != nil {
        return fmt.Errorf("failed to load AWS configuration: %v", err)
    }

    snsClient = sns.NewFromConfig(awsCfg)

    if err := verifySNSClient(); err != nil {
        return fmt.Errorf("failed to verify SNS client: %v", err)
    }

    return nil
}

// verifySNSClient checks if the SNS client is working by listing topics
func verifySNSClient() error {
    input := &sns.ListTopicsInput{}
    _, err := snsClient.ListTopics(context.TODO(), input)
    if err != nil {
        return fmt.Errorf("error verifying SNS client: %v", err)
    }
    log.Println("AWS SNS client successfully verified.")
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
