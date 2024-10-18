FROM --platform=linux/amd64 golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/alertmanager-sns-forwarder ./cmd/alertmanager-sns-forwarder


FROM --platform=linux/amd64 alpine

RUN addgroup -S alertusergroup && adduser -S alertuser -G alertusergroup

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/alertmanager-sns-forwarder /usr/local/bin/alertmanager-sns-forwarder

RUN chown alertuser:alertusergroup /usr/local/bin/alertmanager-sns-forwarder

USER alertuser

EXPOSE 8080

CMD ["/usr/local/bin/alertmanager-sns-forwarder"]
