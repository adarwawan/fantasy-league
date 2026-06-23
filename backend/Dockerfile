FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /bin/server /bin/server
COPY backend/config.yaml config.yaml
EXPOSE 8080
ENTRYPOINT ["/bin/server"]
