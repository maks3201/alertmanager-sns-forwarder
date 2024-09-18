package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	configContent := `
aws_region: "eu-central-1"
sns_topics:
  - name: "TestTopic"
    arn: "arn:aws:sns:eu-central-1:123456789012:TestTopic"
    start_time: "00:00"
    end_time: "23:59"
alertnames:
  - "TestAlert"
`
	err := os.WriteFile("config_test.yaml", []byte(configContent), 0644)
	assert.NoError(t, err)
	defer os.Remove("config_test.yaml")

	originalReadFile := readFile
	readFile = func(filename string) ([]byte, error) {
		return os.ReadFile("config_test.yaml")
	}
	defer func() { readFile = originalReadFile }()

	cfg := LoadConfig()
	assert.Equal(t, "eu-central-1", cfg.AWSRegion)
	assert.Len(t, cfg.Topics, 1)
	assert.Equal(t, "TestTopic", cfg.Topics[0].Name)
	assert.Equal(t, []string{"TestAlert"}, cfg.AlertNames)
}
