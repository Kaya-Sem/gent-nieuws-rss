FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o gent-news-rss

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gent-news-rss .

EXPOSE 8080

CMD ["./gent-news-rss"] 
