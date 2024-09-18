package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/snsiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock SNS Client
type MockSNSClient struct {
	snsiface.ClientAPI
	mock.Mock
}

func (m *MockSNSClient) ListTopics(ctx context.Context, input *sns.ListTopicsInput, opts ...func(*sns.Options)) (*sns.ListTopicsOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sns.ListTopicsOutput), args.Error(1)
}

func (m *MockSNSClient) GetTopicAttributes(ctx context.Context, input *sns.GetTopicAttributesInput, opts ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sns.GetTopicAttributesOutput), args.Error(1)
}

func (m *MockSNSClient) Publish(ctx context.Context, input *sns.PublishInput, opts ...func(*sns.Options)) (*sns.PublishOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sns.PublishOutput), args.Error(1)
}

func TestCheckSNSConnection(t *testing.T) {
	mockSNS := new(MockSNSClient)
	mockSNS.On("ListTopics", mock.Anything, mock.Anything).Return(&sns.ListTopicsOutput{}, nil)

	client := &Client{
		snsClient: mockSNS,
	}

	err := client.CheckSNSConnection(context.Background())
	assert.NoError(t, err)
	mockSNS.AssertExpectations(t)
}

func TestPublishToSNS(t *testing.T) {
	mockSNS := new(MockSNSClient)
	mockSNS.On("Publish", mock.Anything, mock.Anything).Return(&sns.PublishOutput{}, nil)

	client := &Client{
		snsClient: mockSNS,
	}

	err := client.PublishToSNS(context.Background(), "test-arn", "test-message")
	assert.NoError(t, err)
	mockSNS.AssertExpectations(t)
}

func TestTopicExists(t *testing.T) {
	mockSNS := new(MockSNSClient)
	mockSNS.On("GetTopicAttributes", mock.Anything, mock.Anything).Return(&sns.GetTopicAttributesOutput{}, nil)

	client := &Client{
		snsClient: mockSNS,
	}

	exists, err := client.TopicExists("test-arn")
	assert.NoError(t, err)
	assert.True(t, exists)
	mockSNS.AssertExpectations(t)
}
