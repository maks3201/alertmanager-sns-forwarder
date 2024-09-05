## Overview

**SNS Alert Forwarder** is a lightweight service built in Go that forwards alerts from Prometheus Alertmanager to AWS SNS (Simple Notification Service). The service listens for HTTP requests on port 80, processes incoming alerts, and publishes them to an SNS topic specified via environment variables.

### Features

- **AWS SNS Integration**: Easily forward Prometheus alerts to an SNS topic.
- **JSON Structured Logging**: All logs are output in JSON format, making it easy to integrate with modern logging systems.
- **Environment Configuration**: Configure the service entirely through environment variables.
- **Lightweight**: Built on top of Alpine for a minimal footprint.

### Environment Variables

- `AWS_ACCESS_KEY_ID`: Your AWS access key.
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret key.
- `AWS_REGION`: The AWS region where SNS is hosted.
- `SNS_TOPIC_ARN`: The ARN of the SNS topic to forward alerts.

### Usage

```bash
docker run -e AWS_ACCESS_KEY_ID=your_access_key \
           -e AWS_SECRET_ACCESS_KEY=your_secret_key \
           -e AWS_REGION=your_region \
           -e SNS_TOPIC_ARN=your_topic_arn \
           -p 80:80 sns-alert-forwarder
