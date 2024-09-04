FROM --platform=linux/amd64 golang:alpine as builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sns-alert-service .

FROM --platform=linux/amd64 alpine

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/sns-alert-service /usr/local/bin/sns-alert-service

EXPOSE 8080

ENV AWS_REGION=us-east-1

CMD ["/usr/local/bin/sns-alert-service"]
