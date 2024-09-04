FROM --platform=linux/amd64 golang:alpine as builder

WORKDIR /app

COPY . .

RUN apk add --no-cache git && go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/sns-alert-service .

FROM --platform=linux/amd64 alpine

RUN apk --no-cache add ca-certificates

WORKDIR /usr/local/bin

COPY --from=builder /app/sns-alert-service /usr/local/bin/sns-alert-service

EXPOSE 80

CMD ["/usr/local/bin/sns-alert-service"]
