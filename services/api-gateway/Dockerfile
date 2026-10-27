FROM golang:1.26.9-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY proto ./proto
COPY services/api-gateway ./services/api-gateway
WORKDIR /app/services/api-gateway
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/api-gateway/main .
EXPOSE 8080
CMD ["./main"]
