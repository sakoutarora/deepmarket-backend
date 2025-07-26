# Use a multi-stage build to reduce the final image size

# 1st stage: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Declare a build argument for the environment
ARG ENV_FILE

# Copy go mod and sum files for dependency caching
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the selected .env file based on the build argument
COPY ${ENV_FILE} .

# Rename the copied file to .env
RUN mv ${ENV_FILE} .env

# Copy the rest of the application source code
COPY . .


# Build the application
RUN go build -o main ./cmd/main.go

# 2nd stage: Create the final, minimal image
FROM alpine:latest

# Adding timezone data
RUN apk add --no-cache tzdata

WORKDIR /app
RUN mkdir /cmd

# Copy only the built binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/.env . 
COPY --from=builder /app/cmd/* cmd/
# Expose the port your application listens on
EXPOSE 8080

# Set environment variables if needed (e.g., for database connection)
# ENV DATABASE_URL=your_database_url

# Run the application
CMD ["./main"]