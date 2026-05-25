FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o service ./cmd/service/main.go


FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/service .
COPY --from=builder /app/frontend ./frontend

EXPOSE 8081

CMD ["./service"]
