FROM golang:1.21.5-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -mod=vendor -o pastey

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/pastey .

COPY .env .env

EXPOSE 8000

CMD [ "./pastey" ]
