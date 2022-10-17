FROM golang:alpine3.16 AS builder
WORKDIR /build
COPY . .
RUN go build -o p.lee.io

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /build/p.lee.io .
COPY static static
ENTRYPOINT [ "/app/p.lee.io" ]