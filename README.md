
# SNS Alert Forwarder

**SNS Alert Forwarder** is a lightweight service built in Go that forwards alerts from [Prometheus Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) to [AWS SNS (Simple Notification Service)](https://aws.amazon.com/sns/). Additionally, it provides Prometheus-compatible metrics for monitoring alert processing and system performance.

## Features

- **AWS SNS Integration**: Forward Prometheus alerts to AWS SNS topics.
- **Prometheus Metrics**: Expose metrics for monitoring alert processing.
- **Configurable Time Windows**: Define active periods for SNS topics.
- **Batch Processing**: Aggregate alerts over a configurable period before sending.
- **Health Checks**: Provide a `/status` endpoint to verify service's AWS SNS connectivity.
- **Docker Support**: Easily build and deploy using Docker.

## Configuration

The service is configured via a `config.yaml` file. Example:

```yaml
aws_region: "eu-central-1"
sns_topics:
  - name: "alerts-topic"
    arn: "arn:aws:sns:eu-central-1:xxxxxxx:alerts-topic"
    start_time: "00:00"
    end_time: "23:59"
alertnames:
  - "CriticalAlert"
batch_wait_seconds: 3
timeouts:
  server:
    read_timeout_seconds: 5
    write_timeout_seconds: 5
  aws:
    api_call_timeout_seconds: 10
```

## Environment Variables

- **`AWS_ACCESS_KEY_ID`**: AWS access key ID.
- **`AWS_SECRET_ACCESS_KEY`**: AWS secret access key.

## Endpoints

- **`/status`**: Health check to verify SNS connectivity.
- **`/alert`**: Receives alerts from Prometheus Alertmanager.
- **`/metrics`**: Exposes Prometheus metrics.

## Metrics

Metrics exposed by the service:

- `alerts_received_total`: Total alerts received.
- `alerts_filtered_total`: Alerts filtered out.
- `alerts_sent_total`: Alerts sent to SNS.
- `sns_send_duration_seconds`: Time taken to send alerts to SNS.

## Build and Deployment

### Using Docker

Build the Docker image:

```bash
docker build -t sns-alert-forwarder .
```

Run the Docker container:

```bash
docker run -e AWS_ACCESS_KEY_ID=your_access_key            -e AWS_SECRET_ACCESS_KEY=your_secret_key            -v $(pwd)/config/config.yaml:/config/config.yaml            -p 8080:8080 sns-alert-forwarder
```

### Sending Test Alerts

Test the alert forwarding by sending a POST request:

```bash
curl -X POST 127.0.0.1:8080/alert -H "Content-Type: application/json" -d @tests/alert.json
```

### Accessing Metrics

Metrics are available at the `/metrics` endpoint:

```bash
curl http://127.0.0.1:8080/metrics
```


# Architecture Overview

The **SNS Alert Forwarder** service operates in a pipeline-like architecture designed to forward alerts from Prometheus Alertmanager to AWS SNS. Below is a high-level overview of how the system works:

## Key Components

1. **Alert Reception**:
   - The service exposes an HTTP endpoint `/alert` which listens for alerts from Prometheus Alertmanager.
   - Alerts are sent in JSON format and are received as batches.
   
2. **Alert Filtering**:
   - Each incoming alert is checked against a configured list of allowed alert names.
   - If an alert does not match one of the allowed names, it is filtered out and not forwarded.

3. **Batch Processing**:
   - Alerts that pass the filtering process are collected into batches.
   - The batching process waits for a configurable period (`batch_wait_seconds`) before sending the batch to AWS SNS.

4. **Time Window Control**:
   - Each SNS topic has a configurable time window (`start_time` and `end_time`), defining when alerts can be forwarded.
   - If the current time is outside the active time window for a topic, the alert is not sent.

5. **AWS SNS Publishing**:
   - The service connects to AWS SNS and forwards the batched alerts to the specified SNS topics.
   - If an SNS topic is unreachable or AWS SNS encounters errors, the system logs the failure but continues processing.

6. **Health Checks**:
   - The `/status` endpoint performs a health check by testing connectivity to AWS SNS.
   - This ensures that the service is properly connected to AWS and ready to forward alerts.

7. **Metrics**:
   - The service exposes a `/metrics` endpoint, providing Prometheus-compatible metrics.
   - Metrics include the number of received, filtered, and sent alerts, along with the duration of sending batches to SNS.

## Workflow Diagram (Optional)

1. Prometheus Alertmanager sends alerts to the `/alert` endpoint.
2. The service filters alerts based on the configured alert names.
3. Alerts are batched and processed according to the time window rules for each SNS topic.
4. Alerts that pass the checks are forwarded to the appropriate AWS SNS topics.
5. The system reports metrics via `/metrics` for monitoring purposes.

## Sequence of Operations

- **Reception**: Alerts are received from Prometheus Alertmanager.
- **Filtering**: Non-matching alerts are filtered out.
- **Batching**: Alerts are batched and queued for sending.
- **Time Control**: Alerts are only sent during specified time windows.
- **Publishing**: Alerts are forwarded to AWS SNS.
- **Monitoring**: Metrics are tracked and exposed for Prometheus.

This modular architecture allows for flexible alert management and ensures scalability and resilience through the use of AWS SNS.
