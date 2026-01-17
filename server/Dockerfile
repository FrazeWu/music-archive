# Build stage
FROM --platform=linux/arm64 golang:1.23-alpine AS builder

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set Go proxy for better network access in China
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GO111MODULE=on
ENV CGO_ENABLED=1

WORKDIR /app

# Copy go mod
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .
# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Runtime stage
FROM --platform=linux/arm64 alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite-libs

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]