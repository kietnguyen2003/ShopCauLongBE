FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o product-service ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/product-service .

EXPOSE 8082

CMD ["./product-service"]