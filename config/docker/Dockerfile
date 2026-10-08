FROM golang:1.27.1-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY services/avia ./services/avia

WORKDIR /src/services/avia

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/avia ./cmd/avia

FROM scratch

WORKDIR /app

COPY --from=builder /out/avia /app/avia
COPY --from=builder /src/services/avia/config/dev.yaml /app/config/dev.yaml

ENV CONFIG_PATH=/app/config/dev.yaml

EXPOSE 50065

ENTRYPOINT ["/app/avia"]
