FROM golang:1.25 AS dev 
WORKDIR /app
RUN go install github.com/air-verse/air@v1.65.3
RUN go install golang.org/x/tools/gopls@v0.21.0
RUN go install github.com/go-delve/delve/cmd/dlv@v1.27.0

FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o main ./cmd

FROM alpine:3.20 AS prod
RUN adduser -D appuser
WORKDIR /app
COPY --from=builder /app .
USER root
RUN mv /app/deploy/entrypoint.sh /usr/local/bin/entrypoint.sh && \
    chmod +x /usr/local/bin/entrypoint.sh
USER appuser
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

FROM golang:1.25 AS test
WORKDIR /app
COPY --from=builder /app .
