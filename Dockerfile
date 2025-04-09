FROM golang:1.24-alpine3.21 AS builder
WORKDIR /build
COPY . .
RUN go build -o p.lee.io

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /build/p.lee.io .
COPY static static
ENTRYPOINT [ "/app/p.lee.io" ]