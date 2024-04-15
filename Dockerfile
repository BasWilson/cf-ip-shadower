# Use an official Golang runtime as the base image
FROM golang:1.21-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN mkdir bin && go build -o ./bin ./cmd/...

# Start a new stage from scratch
FROM alpine:latest  

# Set the working directory to /app in the container
WORKDIR /app

# Copy the built executable from the builder stage to the /app directory in the final stage
COPY --from=builder /app/bin/shadow .

EXPOSE 1338

# Command to run the executable
CMD ["./shadow"]