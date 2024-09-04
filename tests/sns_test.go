package tests

import (
    "snsalert/internal/aws"
    "testing"
)

func TestPublishToSNS(t *testing.T) {
    // Mock SNS or use a test topic to validate the publish functionality
    err := aws.PublishToSNS("test-topic-arn", "Test message")
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
}
