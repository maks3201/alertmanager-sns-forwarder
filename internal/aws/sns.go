package aws

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/smithy-go"
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

	if err := verifySNSClient(cfg.AWSRegion); err != nil {
		return fmt.Errorf("failed to verify SNS client: %v", err)
	}

	if err := CheckSNSTopicsExistence(cfg); err != nil {
		return fmt.Errorf("SNS topics verification failed: %v", err)
	}

	return nil
}

// verifySNSClient checks if the SNS client is working by listing topics
func verifySNSClient(region string) error {
	input := &sns.ListTopicsInput{}
	_, err := snsClient.ListTopics(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error verifying SNS client: %v", err)
	}
	log.Infof("AWS SNS client successfully verified. Region: %s.", region)
	return nil
}

// CheckSNSTopicsExistence checks if all configured SNS topics exist
func CheckSNSTopicsExistence(cfg config.Config) error {
	for _, topic := range cfg.Topics {
		exists, err := TopicExists(topic.ARN)
		if err != nil {
			return fmt.Errorf("error checking topic %s: %v", topic.Name, err)
		}
		if !exists {
			return fmt.Errorf("SNS topic %s with ARN %s does not exist", topic.Name, topic.ARN)
		}
		log.Infof("SNS topic %s with ARN %s exists.", topic.Name, topic.ARN)
	}
	return nil
}

// TopicExists checks if a topic exists by its ARN
func TopicExists(topicArn string) (bool, error) {
	input := &sns.GetTopicAttributesInput{
		TopicArn: aws.String(topicArn),
	}
	_, err := snsClient.GetTopicAttributes(context.TODO(), input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "InvalidParameter" {
				// Топик не найден или ARN некорректен
				return false, nil
			}
		}
		return false, fmt.Errorf("error getting topic attributes: %v", err)
	}
	return true, nil
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

func CheckSNSConnection() error {
	input := &sns.ListTopicsInput{}
	_, err := snsClient.ListTopics(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error checking SNS connection: %v", err)
	}
	return nil
}

func SimpleSNSConnectionCheck() error {
	input := &sns.ListTopicsInput{}
	_, err := snsClient.ListTopics(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error checking SNS connection: %v", err)
	}
	return nil
}
