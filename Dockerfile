FROM alpine:latest
RUN apk add --no-cache postgresql-client
WORKDIR /app
FROM golang:1.20
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

ENV PORT=3500

# Expose the application port
EXPOSE 3500

# Run the application
CMD ["/app/main"]

