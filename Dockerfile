# Build stage
FROM golang:1.24 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cmd ./cmd

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates git
WORKDIR /root/

COPY --from=builder /build/cmd .

ENTRYPOINT ["./cmd"]