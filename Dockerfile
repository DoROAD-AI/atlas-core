# Build Stage
FROM golang:1.23.5 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o atlas .

# Runtime Stage
FROM gcr.io/distroless/base-debian12

# Expose the application port
EXPOSE 3101

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/atlas ./

# Create the data directory and copy the JSON files there
COPY --from=builder /app/data /app/data

# Create the docs directory and copy the JSON files there
COPY --from=builder /app/docs /app/docs

# Command to run the application
ENTRYPOINT ["./atlas"]