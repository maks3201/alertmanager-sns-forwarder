## Overview

**SNS Alert Forwarder** is a lightweight service built in Go that forwards alerts from Prometheus Alertmanager to AWS SNS (Simple Notification Service).

### Features

- **AWS SNS Integration**: Easily forward Prometheus alerts to an SNS topic.
- **Environment Configuration**: Configure the service entirely through environment variables.
- **Lightweight**: Built on top of Alpine for a minimal footprint.

### Environment Variables

- `AWS_ACCESS_KEY_ID`: Your AWS access key.
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret key.
- `AWS_REGION`: The AWS region where SNS is hosted.
- `SNS_TOPIC_ARN`: The ARN of the SNS topic to forward alerts.
- `ALERT_START_TIME`
- `ALERT_END_TIME`

### Compile
```
docker build -f docker/Dockerfile -t sns-alert-forwarder .
```

### Usage

```bash
docker run -e AWS_ACCESS_KEY_ID=your_access_key \
           -e AWS_SECRET_ACCESS_KEY=your_secret_key \
           -e AWS_REGION=your_region \
           -e SNS_TOPIC_ARN=your_topic_arn \
           -e ALERT_START_TIME=08:00 \
           -e ALERT_END_TIME=18:00 \
           -p 80:80 sns-alert-forwarder
