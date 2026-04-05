FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go mod download && CGO_ENABLED=0 go build -o harvest ./cmd/harvest/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/harvest .
ENV PORT=9810 DATA_DIR=/data
EXPOSE 9810
CMD ["./harvest"]
