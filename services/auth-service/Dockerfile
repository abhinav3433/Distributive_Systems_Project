FROM golang:1.26.9-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY proto ./proto
COPY services/auth-service ./services/auth-service
WORKDIR /app/services/auth-service
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/auth-service/main .
EXPOSE 8081 9091
CMD ["./main"]
