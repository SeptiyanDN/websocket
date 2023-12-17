# Stage 1: Build the application
FROM golang:1.20 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

# Stage 2: Create a minimal image to run the application
FROM alpine:latest
RUN apk add --no-cache postgresql-client
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

ENV PORT=3500

# Expose the application port
EXPOSE 3500

# Run the application
CMD ["./main"]
