# Build stage
FROM golang:1.21.6-alpine3.18 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
COPY db/migration ./db/migration
COPY start.sh .
COPY wait-for.sh .

EXPOSE 8080 9090
CMD ["/app/main"]
ENTRYPOINT ["/app/start.sh"]
