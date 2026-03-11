# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install git for downloading dependencies
RUN apk add --no-cache git

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o velocitydbgo .

# Run stage
FROM alpine:latest  

WORKDIR /root/

# Install CA certificates for HTTPS requests if needed
RUN apk --no-cache add ca-certificates

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/velocitydbgo .

# Copy the static assets and Swagger docs
COPY --from=builder /app/public ./public
COPY --from=builder /app/docs ./docs

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./velocitydbgo"]
