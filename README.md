## Overview

**SNS Alert Forwarder** is a lightweight service built in Go that forwards alerts from Prometheus Alertmanager to AWS SNS (Simple Notification Service).

### Features

- **AWS SNS Integration**: Easily forward Prometheus alerts to an SNS topic.
- **Environment Configuration**: Configure the service entirely through environment variables.
- **Lightweight**: Built on top of Alpine for a minimal footprint.

### Environment Variables

- `AWS_ACCESS_KEY_ID`: Your AWS access key.
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret key.

### Compile
```
docker build -f docker/Dockerfile -t sns-alert-forwarder .
```

### Usage

Modify config/config.yaml file.

```bash
docker run -e AWS_ACCESS_KEY_ID=your_access_key \
           -e AWS_SECRET_ACCESS_KEY=your_secret_key \
           -v $(pwd)/config/config.yaml:/config/config.yaml\
           -p 8080:80 sns-alert-forwarder
```

### Tests
```
curl -X POST 127.1:8080/alert      -H "Content-Type: application/json" \
     -d @tests/alert.json
```

### Golang Tests
```
go test ./... -cover
```
