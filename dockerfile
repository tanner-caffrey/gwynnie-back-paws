# Stage 1: Build the Go application
FROM golang:1.21 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./

# Download dependencies (cached if go.mod/go.sum didn't change)
RUN go mod download

# Copy the entire project into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o main .

# Stage 2: Create a smaller runtime image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose a port if your application listens on one (optional)
EXPOSE 8080

RUN ls -l ./main

# Set the default command to run the application
CMD ["./main"]
