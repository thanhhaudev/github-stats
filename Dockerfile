# Build stage
FROM golang:1.24 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/github-stats ./cmd

# Runtime stage
FROM alpine:latest

WORKDIR /root
COPY --from=builder /build/github-stats /root/github-stats
RUN chmod +x /root/github-stats

ENTRYPOINT ["/root/github-stats"]