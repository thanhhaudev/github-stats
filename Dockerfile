FROM golang:1.22

WORKDIR /workspace

COPY . .

RUN go mod tidy

ENTRYPOINT ["go", "run", "cmd/main.go"]
