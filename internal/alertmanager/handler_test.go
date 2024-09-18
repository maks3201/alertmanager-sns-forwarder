package alertmanager

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maks3201/sns-alert-service/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAWSClient struct {
	mock.Mock
}

func (m *MockAWSClient) PublishToSNS(ctx context.Context, topicArn string, message string) error {
	args := m.Called(ctx, topicArn, message)
	return args.Error(0)
}

func (m *MockAWSClient) CheckSNSConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestSNSHandler(t *testing.T) {
	cfg := config.Config{
		AlertNames: []string{"TestAlert"},
		Topics: []config.SNSTopicConfig{
			{
				Name:      "TestTopic",
				ARN:       "arn:aws:sns:eu-central-1:123456789012:TestTopic",
				StartTime: "00:00",
				EndTime:   "23:59",
			},
		},
	}

	mockAWSClient := new(MockAWSClient)
	mockAWSClient.On("PublishToSNS", mock.Anything, cfg.Topics[0].ARN, mock.Anything).Return(nil)

	handler := NewHandler(cfg, mockAWSClient)

	alertJSON := `{
        "alerts": [
            {
                "status": "firing",
                "labels": {
                    "alertname": "TestAlert",
                    "instance": "test-instance",
                    "severity": "critical"
                },
                "annotations": {
                    "summary": "Test summary"
                },
                "startsAt": "2021-01-01T00:00:00Z",
                "endsAt": ""
            }
        ]
    }`

	req := httptest.NewRequest("POST", "/alert", bytes.NewBufferString(alertJSON))
	w := httptest.NewRecorder()

	handler.SNSHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "Alerts sent to all available topics")
	mockAWSClient.AssertExpectations(t)
}

func TestIsAlertFiltered(t *testing.T) {
	allowedAlerts := []string{"Alert1", "Alert2"}
	assert.True(t, isAlertFiltered("Alert1", allowedAlerts))
	assert.False(t, isAlertFiltered("Alert3", allowedAlerts))
	assert.True(t, isAlertFiltered("AnyAlert", []string{}))
}

func TestIsTopicAvailable(t *testing.T) {
	startTime, _ := parseTime("15:04", "08:00")
	endTime, _ := parseTime("15:04", "18:00")
	currentTime, _ := parseTime("15:04", "12:00")
	assert.True(t, isTopicAvailable(startTime, endTime, currentTime))

	currentTime, _ = parseTime("15:04", "20:00")
	assert.False(t, isTopicAvailable(startTime, endTime, currentTime))

	startTime, _ = parseTime("15:04", "22:00")
	endTime, _ = parseTime("15:04", "06:00")
	currentTime, _ = parseTime("15:04", "23:00")
	assert.True(t, isTopicAvailable(startTime, endTime, currentTime))
	currentTime, _ = parseTime("15:04", "05:00")
	assert.True(t, isTopicAvailable(startTime, endTime, currentTime))
}
