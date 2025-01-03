FROM golang:1.23-alpine as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go mod download
RUN go build -o api cmd/api/*.go

FROM alpine:latest
RUN apk add multirun
WORKDIR /app
COPY --from=builder /app/. ./
CMD ["multirun", "/app/api server"]