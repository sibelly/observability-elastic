FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/favorite-go .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/favorite-go /opt/favorite-go

CMD ["sh", "-c", "cp /opt/favorite-go /shared/favorite-go && chmod 755 /shared/favorite-go && exec /shared/favorite-go"]