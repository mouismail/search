FROM golang:1.22.3 AS builder

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download
COPY . .
RUN go build -o search-app main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/search-app .
EXPOSE 8080
CMD ["./search-app"]