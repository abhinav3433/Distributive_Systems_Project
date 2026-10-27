FROM golang:1.26.9-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY proto ./proto
COPY services/product-service ./services/product-service
WORKDIR /app/services/product-service
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/product-service/main .
EXPOSE 8082 9092
CMD ["./main"]
