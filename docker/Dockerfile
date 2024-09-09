FROM --platform=linux/amd64 golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/sns-alert-service ./cmd/sns-alert-service

FROM --platform=linux/amd64 alpine

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/sns-alert-service /usr/local/bin/sns-alert-service

EXPOSE 80

CMD ["/usr/local/bin/sns-alert-service"]
