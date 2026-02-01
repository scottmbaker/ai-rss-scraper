# ai-rss-scraper: a program to scrape rss feeds and score them using AI, then email the results
# Scott Baker, https://medium.com/@smbaker
# https://github.com/scottmbaker/ai-rss-scraper

# First container is our build container
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Use CGO_ENABLED=1 for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -a -o ai-rss-scraper ./cmd/ai-rss-scraper

# Second container is our run container.
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/ai-rss-scraper .

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

EXPOSE 8080

ENTRYPOINT ["./ai-rss-scraper"]
