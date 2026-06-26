FROM golang:1.25.0-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hotel-backend ./cmd/server

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/hotel-backend .

EXPOSE 8080

CMD ["./hotel-backend"]
