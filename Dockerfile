FROM golang:1.21.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/scraper/tweet

FROM golang:1.21.2-alpine AS runner

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]