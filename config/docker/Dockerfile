FROM golang:1.27.1-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/avia ./cmd/avia

FROM scratch
WORKDIR /app

COPY --from=builder /out/avia /app/avia
COPY --from=builder /src/config/dev.yaml /app/config/dev.yaml
COPY --from=builder /src/migrations /app/migrations

ENV CONFIG_PATH=/app/config/dev.yaml
EXPOSE 8080
ENTRYPOINT ["/app/avia"]
