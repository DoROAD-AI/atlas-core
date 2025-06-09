# Build Stage
FROM golang:1.24.4 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o atlas-core .

# Runtime Stage
FROM gcr.io/distroless/base-debian12

# Expose the application port
EXPOSE 3101

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage with proper ownership
COPY --from=builder --chown=nonroot:nonroot /app/atlas-core ./

# Copy the data directory with proper ownership
COPY --from=builder --chown=nonroot:nonroot /app/data /app/data

# Copy the docs directory with proper ownership
COPY --from=builder --chown=nonroot:nonroot /app/docs /app/docs

# Switch to nonroot user for security
USER nonroot:nonroot

# Command to run the application
ENTRYPOINT ["./atlas-core"]