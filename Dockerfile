FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

Copy the actual application code

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hotel-backend .



FROM alpine:latest

WORKDIR /root/

Add certificates to make secure HTTPS requests to Stripe/AWS/etc

RUN apk --no-cache add ca-certificates tzdata

Copy the binary from the builder stage

COPY --from=builder /app/hotel-backend .

EXPOSE 8080

CMD ["./hotel-backend"]
