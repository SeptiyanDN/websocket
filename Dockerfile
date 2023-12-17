FROM alpine:latest
RUN apk add --no-cache postgresql-client
WORKDIR /app
FROM golang:1.20
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .
# Set environment variables for database connection
ENV DB_HOST=34.101.230.132
ENV DB_PORT=5432
ENV DB_NAME=evolvitech
ENV DB_USER=postgres
ENV DB_PASSWORD=Development

ENV PORT=3500

# Expose the application port
EXPOSE 3500

# Run the application
CMD ["/app/main"]

