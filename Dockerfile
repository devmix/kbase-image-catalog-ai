FROM golang:1.24-alpine AS builder

WORKDIR /kbic
COPY . .

# Install dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kbase-catalog cmd/kbase-catalog/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /kbic/

# Copy the binary and config file
COPY --from=builder /kbic/kbase-catalog .
COPY --from=builder /kbic/config.yaml .

EXPOSE 8080

CMD ["./kbase-catalog", "web", "-port", "8080"]