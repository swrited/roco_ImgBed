# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO (SQLite support)
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/server .

# Create uploads directory
RUN mkdir -p uploads

# Use non-root user for security
RUN adduser -D -g '' appuser && chown -R appuser:appuser /app
USER appuser

EXPOSE 8000

CMD ["./server"]