# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build ./cmd/main.go

# Run stage
FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 50051
CMD [ "/app/main" ]