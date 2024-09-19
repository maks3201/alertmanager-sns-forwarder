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

type SNSAPI interface {
	ListTopics(ctx context.Context, params *sns.ListTopicsInput, optFns ...func(*sns.Options)) (*sns.ListTopicsOutput, error)
	GetTopicAttributes(ctx context.Context, params *sns.GetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error)
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type SNSClient interface {
	PublishToSNS(ctx context.Context, topicArn string, message string) error
	CheckSNSConnection(ctx context.Context) error
}

type Client struct {
	snsClient SNSAPI
}

func InitSNSClient(cfg config.Config) (*Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %v", err)
	}

	client := &Client{
		snsClient: sns.NewFromConfig(awsCfg),
	}

	if err := client.verifySNSClient(cfg.AWSRegion); err != nil {
		return nil, fmt.Errorf("failed to verify SNS client: %v", err)
	}

	if err := client.CheckSNSTopicsExistence(cfg); err != nil {
		return nil, fmt.Errorf("SNS topics verification failed: %v", err)
	}

	return client, nil
}

func (c *Client) verifySNSClient(region string) error {
	input := &sns.ListTopicsInput{}
	_, err := c.snsClient.ListTopics(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error verifying SNS client: %v", err)
	}
	log.Infof("AWS SNS client successfully verified. Region: %s.", region)
	return nil
}

func (c *Client) CheckSNSTopicsExistence(cfg config.Config) error {
	for _, topic := range cfg.Topics {
		exists, err := c.TopicExists(topic.ARN)
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

func (c *Client) TopicExists(topicArn string) (bool, error) {
	input := &sns.GetTopicAttributesInput{
		TopicArn: aws.String(topicArn),
	}
	_, err := c.snsClient.GetTopicAttributes(context.TODO(), input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "InvalidParameter" {
				return false, nil
			}
		}
		return false, fmt.Errorf("error getting topic attributes: %v", err)
	}
	return true, nil
}

func (c *Client) PublishToSNS(ctx context.Context, topicArn string, message string) error {
	_, err := c.snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(message),
	})
	if err != nil {
		return fmt.Errorf("failed to publish message to SNS: %v", err)
	}
	return nil
}

func (c *Client) CheckSNSConnection(ctx context.Context) error {
	input := &sns.ListTopicsInput{}
	_, err := c.snsClient.ListTopics(ctx, input)
	if err != nil {
		return fmt.Errorf("error checking SNS connection: %v", err)
	}
	return nil
}
