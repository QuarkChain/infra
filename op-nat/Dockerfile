FROM golang:1.23.5-alpine3.21 AS builder

WORKDIR /app

# Install dependencies first for better caching
RUN apk add --no-cache just

COPY op-nat/ .

RUN just build

FROM alpine:3.21

WORKDIR /app
RUN addgroup -S app && adduser -S app -G app && \
    mkdir -p /devnets && \
    chown -R app:app /devnets

COPY --from=builder /app/bin/op-nat /app/
USER app

VOLUME /devnets
ENTRYPOINT ["/app/op-nat"]
